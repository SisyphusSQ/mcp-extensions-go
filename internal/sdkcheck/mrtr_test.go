// Package sdkcheck tests public SDK extension seams without adding an RPC stack.
package sdkcheck

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
)

// These continuation rules belong to this business fixture, not the SDK.
// The fixture is connection-scoped and deliberately not an HTTP persistence API.
type continuation struct {
	owner   *mcp.ServerSession
	expires time.Time
	round   int
	result  *mcp.CallToolResult
}

type workflow struct {
	mu      sync.Mutex
	pending map[string]*continuation
	writes  int
}

func newWorkflow() (*mcp.Server, *workflow) {
	w := &workflow{pending: make(map[string]*continuation)}
	s := mcp.NewServer(&mcp.Implementation{Name: "mrtr-investigation", Version: "0"}, nil)
	s.AddTool(&mcp.Tool{Name: "review", InputSchema: map[string]any{"type": "object"}}, w.call)
	return s, w
}

func toolError(message string) (*mcp.CallToolResult, error) {
	r := new(mcp.CallToolResult)
	r.SetError(fmt.Errorf("%s", message))
	return r, nil
}

func (w *workflow) call(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	token := req.Params.RequestState
	if token == "" {
		if len(req.Params.InputResponses) != 0 {
			return toolError("responses without a continuation")
		}
		token = rand.Text()
		w.pending[token] = &continuation{owner: req.Session, expires: time.Now().Add(time.Minute)}
	}
	c := w.pending[token]
	if c == nil || c.owner != req.Session {
		return toolError("unknown continuation or wrong session")
	}
	if !time.Now().Before(c.expires) {
		return toolError("continuation expired")
	}
	if c.result != nil {
		return c.result, nil
	}
	id := "choose"
	if c.round == 1 {
		id = "confirm"
	}
	if len(req.Params.InputResponses) != 0 {
		response, ok := req.Params.InputResponses[id].(*mcp.ElicitResult)
		if !ok || len(req.Params.InputResponses) != 1 {
			return toolError("wrong input request id or type")
		}
		switch response.Action {
		case "cancel", "decline":
			c.result = &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: response.Action}}}
			return c.result, nil
		case "accept":
			if c.round == 0 && response.Content["part"] != "bolt" || c.round == 1 && response.Content["approved"] != true {
				return toolError("invalid form response")
			}
		default:
			return toolError("invalid elicitation action")
		}
		c.round++
		if c.round == 2 {
			w.writes++
			c.result = &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "saved"}}}
			return c.result, nil
		}
		id = "confirm"
	}
	field, schema := "part", map[string]any{"type": "string", "enum": []string{"bolt"}}
	if c.round == 1 {
		field, schema = "approved", map[string]any{"type": "boolean"}
	}
	return &mcp.CallToolResult{RequestState: token, InputRequests: mcp.InputRequestMap{id: &mcp.ElicitParams{
		Mode: "form", Message: "Review part", RequestedSchema: map[string]any{"type": "object", "properties": map[string]any{field: schema}, "required": []string{field}},
	}}}, nil
}

func manualClient(t *testing.T, server *mcp.Server) *mcp.ClientSession {
	return testutil.Connect(t, server, &mcp.ClientOptions{MultiRoundTrip: &mcp.MultiRoundTripOptions{Disabled: true}}, "2026-07-28")
}

func invoke(t *testing.T, cs *mcp.ClientSession, state, id, action string, content map[string]any) *mcp.CallToolResult {
	t.Helper()
	params := &mcp.CallToolParams{Name: "review", Arguments: map[string]any{}, RequestState: state}
	if id != "" {
		params.InputResponses = mcp.InputResponseMap{id: &mcp.ElicitResult{Action: action, Content: content}}
	}
	res, err := cs.CallTool(t.Context(), params)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestStandardMRTRResumeReplayExpiryAndIsolation(t *testing.T) {
	server, flow := newWorkflow()
	cs := manualClient(t, server)
	other := manualClient(t, server)
	first := invoke(t, cs, "", "", "", nil)
	if !first.NeedsInput() || first.InputRequests["choose"] == nil || first.RequestState == "" {
		t.Fatalf("first round = %#v", first)
	}
	if res := invoke(t, other, first.RequestState, "choose", "accept", map[string]any{"part": "bolt"}); !res.IsError {
		t.Fatal("cross-session resume accepted")
	}
	if res := invoke(t, cs, first.RequestState, "wrong", "accept", nil); !res.IsError {
		t.Fatal("unknown request id accepted")
	}
	second := invoke(t, cs, first.RequestState, "choose", "accept", map[string]any{"part": "bolt"})
	if !second.NeedsInput() || second.InputRequests["confirm"] == nil {
		t.Fatalf("second round = %#v", second)
	}
	if res := invoke(t, cs, first.RequestState, "choose", "accept", map[string]any{"part": "bolt"}); !res.IsError {
		t.Fatal("stale previous-round input accepted")
	}
	for range 2 {
		res := invoke(t, cs, second.RequestState, "confirm", "accept", map[string]any{"approved": true})
		if res.NeedsInput() || res.IsError || res.Content[0].(*mcp.TextContent).Text != "saved" {
			t.Fatalf("completion = %#v", res)
		}
	}
	flow.mu.Lock()
	writes := flow.writes
	flow.mu.Unlock()
	if writes != 1 {
		t.Fatal("business fixture did not deduplicate completion")
	}
	expired := invoke(t, cs, "", "", "", nil)
	flow.mu.Lock()
	flow.pending[expired.RequestState].expires = time.Now().Add(-time.Second)
	flow.mu.Unlock()
	if res := invoke(t, cs, expired.RequestState, "choose", "accept", map[string]any{"part": "bolt"}); !res.IsError {
		t.Fatal("expired continuation accepted")
	}
	for _, action := range []string{"cancel", "decline"} {
		start := invoke(t, cs, "", "", "", nil)
		res := invoke(t, cs, start.RequestState, "choose", action, nil)
		if res.IsError || res.NeedsInput() || res.Content[0].(*mcp.TextContent).Text != action {
			t.Fatalf("%s = %#v", action, res)
		}
	}
}

func TestStandardMRTRAutomaticAndConcurrent(t *testing.T) {
	server, flow := newWorkflow()
	var forms atomic.Int64
	cs := testutil.Connect(t, server, &mcp.ClientOptions{ElicitationHandler: func(_ context.Context, req *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
		forms.Add(1)
		properties := req.Params.RequestedSchema.(map[string]any)["properties"].(map[string]any)
		if properties["part"] != nil {
			return &mcp.ElicitResult{Action: "accept", Content: map[string]any{"part": "bolt"}}, nil
		}
		return &mcp.ElicitResult{Action: "accept", Content: map[string]any{"approved": true}}, nil
	}}, "2026-07-28")
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "review", Arguments: map[string]any{}})
			if err != nil || res.IsError || res.NeedsInput() {
				t.Errorf("automatic MRTR = %#v %v", res, err)
			}
		})
	}
	wg.Wait()
	flow.mu.Lock()
	writes := flow.writes
	flow.mu.Unlock()
	if writes != 8 || forms.Load() != 16 {
		t.Fatalf("concurrency crossed calls: writes=%d forms=%d", writes, forms.Load())
	}
}

func TestStandardMRTRCancellationDuringInput(t *testing.T) {
	server, _ := newWorkflow()
	entered := make(chan struct{})
	cs := testutil.Connect(t, server, &mcp.ClientOptions{ElicitationHandler: func(ctx context.Context, _ *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	}}, "2026-07-28")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "review", Arguments: map[string]any{}})
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("input handler was not reached")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled MRTR reported success")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("canceled MRTR did not finish")
	}
}

type extensionInput struct{ *mcp.ElicitParams }

func TestOpenAIInputRequestMapBoundary(t *testing.T) {
	if _, err := json.Marshal(mcp.InputRequestMap{"pick": &extensionInput{ElicitParams: &mcp.ElicitParams{Mode: "form", Message: "Pick"}}}); err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("custom embedded input request unexpectedly encoded: %v", err)
	}
	var inputs mcp.InputRequestMap
	err := json.Unmarshal([]byte(`{"pick":{"method":"openai/elicitation/create","params":{"mode":"form","message":"Pick"}}}`), &inputs)
	if err == nil || !strings.Contains(err.Error(), "unsupported InputRequest method") {
		t.Fatalf("custom input request unexpectedly decoded: %v", err)
	}
}

func TestSendingMiddlewareCannotEnableOpenAIMethod(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	server.AddSendingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method == "elicitation/create" {
				return next(ctx, "openai/elicitation/create", req)
			}
			return next(ctx, method, req)
		}
	})
	mcp.AddTool(server, &mcp.Tool{Name: "elicit"}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		_, err := req.Session.Elicit(ctx, &mcp.ElicitParams{Message: "Pick", RequestedSchema: map[string]any{"type": "object"}})
		return nil, nil, err
	})
	var received atomic.Int64
	cs := testutil.Connect(t, server, &mcp.ClientOptions{ElicitationHandler: func(context.Context, *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
		received.Add(1)
		return &mcp.ElicitResult{Action: "cancel"}, nil
	}}, "2025-11-25")
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "elicit", Arguments: map[string]any{}})
	if err != nil || !res.IsError || received.Load() != 0 {
		t.Fatalf("sending boundary = %#v %v, received=%d", res, err, received.Load())
	}
	t.Logf("public sending middleware rejection: %s", res.Content[0].(*mcp.TextContent).Text)
	if err := mcp.AddReceivingCustomMethod(server, "tools/call", func(context.Context, *mcp.ServerSession, *mcp.ParamsBase) (*mcp.ResultBase, error) { return nil, nil }); err == nil {
		t.Fatal("custom registration shadowed a standard method")
	}
}

// ResultBase permits encoding extension results on a receiving middleware seam.
// This probe is intentionally private: it is not a complete forms API.
type adaptedResult struct {
	mcp.ResultBase
	result *mcp.CallToolResult
}

func (r *adaptedResult) MarshalJSON() ([]byte, error) {
	encoded, err := json.Marshal(r.result)
	if err != nil {
		return nil, err
	}
	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		return nil, err
	}
	for _, value := range wire["inputRequests"].(map[string]any) {
		value.(map[string]any)["method"] = "openai/elicitation/create"
	}
	return json.Marshal(wire)
}

func TestReceivingResultAdaptationReachesWireButNotTypedClient(t *testing.T) {
	server, _ := newWorkflow()
	server.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			res, err := next(ctx, method, req)
			if tool, ok := res.(*mcp.CallToolResult); err == nil && ok && tool.NeedsInput() {
				return &adaptedResult{result: tool}, nil
			}
			return res, err
		}
	})
	host := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}))
	t.Cleanup(host.Close)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, host.URL, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"review","arguments":{},"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{"extensions":{"openai/elicitation":{"form":{}}}}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2026-07-28")
	req.Header.Set("Mcp-Method", "tools/call")
	req.Header.Set("Mcp-Name", "review")
	response, err := host.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), `"openai/elicitation/create"`) || !strings.Contains(string(body), `"resultType":"input_required"`) {
		t.Fatalf("receiving seam did not encode extension result: %s", body)
	}
	cs := manualClient(t, server)
	_, err = cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "review", Arguments: map[string]any{}})
	if err == nil || !strings.Contains(err.Error(), "unsupported InputRequest method") {
		t.Fatalf("typed client unexpectedly consumed adapted result: %v", err)
	}
}
