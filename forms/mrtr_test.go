package forms_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
)

func extendedForm(t *testing.T) *forms.Form {
	t.Helper()
	f, err := forms.New(forms.Schema{Type: "object", Properties: map[string]forms.Field{
		"code":   {Type: "string", Pattern: "^[A-Z]{3}$", Suggestions: []forms.Option{{Const: "ABC", Title: "Example"}}},
		"images": {Type: "array", Items: &forms.Field{Type: "string", Format: "uri"}, Input: &forms.ResourceInput{Type: "resource", Options: []*forms.ResourceOption{{Resource: mcp.Resource{URI: "file:///image.png", Name: "Image", Meta: mcp.Meta{"openai/thumbnail": map[string]any{"src": "https://example.com/image.png"}}}}}}},
	}, Required: []string{"code", "images"}})
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func modernRequest() *mcp.CallToolRequest {
	return &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Meta: mcp.Meta{
		mcp.MetaKeyProtocolVersion:    "2026-07-28",
		mcp.MetaKeyClientCapabilities: map[string]any{"elicitation": map[string]any{"form": map[string]any{}}, "extensions": map[string]any{forms.CapabilityKey: map[string]any{"form": map[string]any{}}}},
	}}}
}

func TestRequestInputWireAndValidation(t *testing.T) {
	f := extendedForm(t)
	meta := mcp.Meta{"vendor": map[string]any{"id": "original"}}
	req := modernRequest()
	pending, answer, err := f.RequestInput(t.Context(), req, forms.RequestOptions{Key: "details", Message: "Choose", RequestState: "opaque", Meta: meta})
	if err != nil || answer != nil || pending.InputRequests == nil {
		t.Fatalf("pending: %#v %#v %v", pending, answer, err)
	}
	meta["vendor"].(map[string]any)["id"] = "changed"
	encoded, err := json.Marshal(pending)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		ResultType    string `json:"resultType"`
		RequestState  string `json:"requestState"`
		InputRequests map[string]struct {
			Method string           `json:"method"`
			Params mcp.ElicitParams `json:"params"`
		} `json:"inputRequests"`
	}
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	input := result.InputRequests["details"]
	if result.ResultType != "" || result.RequestState != "opaque" || input.Method != "elicitation/create" || input.Params.Meta["vendor"].(map[string]any)["id"] != "original" {
		t.Fatalf("wire = %s", encoded)
	}
	if !reflect.DeepEqual(input.Params.RequestedSchema, map[string]any{"type": "object", "properties": map[string]any{}}) {
		t.Fatal("core schema is not empty")
	}
	form := input.Params.Meta[forms.CapabilityKey].(map[string]any)["requestedSchema"].(map[string]any)
	if form["properties"].(map[string]any)["code"].(map[string]any)["x-openai-suggestions"] == nil {
		t.Fatal("lost extended schema")
	}
	for _, action := range []string{"accept", "cancel", "decline"} {
		submitted := &forms.Answer{Action: action}
		if action == "accept" {
			submitted.Content = map[string]any{"code": "ABC", "images": []string{"file:///image.png"}}
		}
		req.Params.InputResponses = mcp.InputResponseMap{"details": submitted}
		pending, answer, err = f.RequestInput(t.Context(), req, forms.RequestOptions{Key: "details", Message: "Choose"})
		if err != nil || pending != nil || answer != submitted {
			t.Fatalf("answer = %#v, %v", answer, err)
		}
	}
	for _, content := range []map[string]any{nil, {"code": "bad", "images": []string{"file:///image.png"}}, {"code": "ABC", "images": []string{"file:///unlisted.png"}}, {"code": "ABC"}} {
		req.Params.InputResponses = mcp.InputResponseMap{"details": &forms.Answer{Action: "accept", Content: content}}
		if _, _, err := f.RequestInput(t.Context(), req, forms.RequestOptions{Key: "details", Message: "Choose"}); !errors.Is(err, forms.ErrInvalidAnswer) {
			t.Fatalf("invalid content accepted: %v", err)
		}
	}
	req.Params.InputResponses = mcp.InputResponseMap{"details": &mcp.ListRootsResult{}}
	if _, _, err := f.RequestInput(t.Context(), req, forms.RequestOptions{Key: "details", Message: "Choose"}); !errors.Is(err, forms.ErrInvalidAnswer) {
		t.Fatalf("wrong response type: %v", err)
	}
}

func TestRequestInputRequiresBothCapabilitiesAndModernProtocol(t *testing.T) {
	f := extendedForm(t)
	for _, mutate := range []func(*mcp.CallToolRequest){
		func(req *mcp.CallToolRequest) { req.Params.Meta[mcp.MetaKeyProtocolVersion] = "2025-11-25" },
		func(req *mcp.CallToolRequest) { delete(req.Params.Meta, mcp.MetaKeyClientCapabilities) },
		func(req *mcp.CallToolRequest) {
			delete(req.Params.Meta[mcp.MetaKeyClientCapabilities].(map[string]any), "elicitation")
		},
		func(req *mcp.CallToolRequest) {
			delete(req.Params.Meta[mcp.MetaKeyClientCapabilities].(map[string]any), "extensions")
		},
		func(req *mcp.CallToolRequest) {
			req.Params.Meta[mcp.MetaKeyClientCapabilities].(map[string]any)["extensions"] = map[string]any{forms.CapabilityKey: map[string]any{"form": nil}}
		},
	} {
		req := modernRequest()
		mutate(req)
		if _, _, err := f.RequestInput(t.Context(), req, forms.RequestOptions{Key: "form", Message: "Choose"}); !errors.Is(err, forms.ErrUnsupportedClient) {
			t.Fatalf("missing support accepted: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := f.RequestInput(ctx, modernRequest(), forms.RequestOptions{Key: "form", Message: "Choose"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestOfficialClientCompletesTwoRoundsConcurrently(t *testing.T) {
	f := extendedForm(t)
	confirmation, err := forms.New(forms.Schema{Type: "object", Properties: map[string]forms.Field{"confirmed": {Type: "boolean"}}, Required: []string{"confirmed"}})
	if err != nil {
		t.Fatal(err)
	}
	s := mcp.NewServer(&mcp.Implementation{Name: "forms", Version: "0"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "choose"}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Label string `json:"label"`
	}) (*mcp.CallToolResult, map[string]any, error) {
		if req.Params.RequestState == args.Label+":confirm" {
			pending, answer, err := confirmation.RequestInput(ctx, req, forms.RequestOptions{Key: "confirm", Message: "Confirm"})
			if err != nil || pending != nil {
				return pending, nil, err
			}
			return nil, map[string]any{"label": args.Label, "action": answer.Action, "content": answer.Content}, nil
		}
		pending, answer, err := f.RequestInput(ctx, req, forms.RequestOptions{Key: "details", Message: args.Label})
		if err != nil || pending != nil {
			return pending, nil, err
		}
		if answer.Action != "accept" {
			return nil, map[string]any{"action": answer.Action}, nil
		}
		pending, _, err = confirmation.RequestInput(ctx, req, forms.RequestOptions{Key: "confirm", Message: "Confirm", RequestState: args.Label + ":confirm"})
		return pending, nil, err
	})
	var mu sync.Mutex
	var messages []string
	cs := testutil.Connect(t, s, &mcp.ClientOptions{Capabilities: &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{Form: &mcp.FormElicitationCapabilities{}}, Extensions: map[string]any{forms.CapabilityKey: map[string]any{"form": map[string]any{}}}}, ElicitationHandler: func(_ context.Context, req *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
		mu.Lock()
		messages = append(messages, req.Params.Message)
		mu.Unlock()
		if req.Params.Meta[forms.CapabilityKey] == nil {
			t.Error("extended metadata did not reach client")
		}
		if req.Params.Message == "Confirm" {
			return &mcp.ElicitResult{Action: "accept", Content: map[string]any{"confirmed": true}}, nil
		}
		return &mcp.ElicitResult{Action: "accept", Content: map[string]any{"code": "ABC", "images": []string{"file:///image.png"}}}, nil
	}}, "2026-07-28")
	var wg sync.WaitGroup
	for _, label := range []string{"first", "second"} {
		wg.Go(func() {
			result, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "choose", Arguments: map[string]any{"label": label}})
			if err != nil || result.IsError {
				t.Errorf("%s: %#v %v", label, result, err)
				return
			}
			if result.StructuredContent.(map[string]any)["label"] != label {
				t.Error("crossed concurrent result")
			}
		})
	}
	wg.Wait()
	if len(messages) != 4 {
		t.Fatalf("rounds = %#v", messages)
	}
}
