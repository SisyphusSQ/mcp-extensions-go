package main

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/ui"
)

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (b bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	request := req.Clone(req.Context())
	request.Header.Set("Authorization", "Bearer "+b.token)
	return b.base.RoundTrip(request)
}

func TestAuthenticatedMCPOverHTTP(t *testing.T) {
	token := rand.Text() + rand.Text()
	handler, err := newHandler(token)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := mcp.NewClient(&mcp.Implementation{Name: "integration-test", Version: "0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint: server.URL + "/mcp", DisableStandaloneSSE: true, MaxRetries: -1,
		HTTPClient: &http.Client{Transport: bearerTransport{token: token, base: server.Client().Transport}, Timeout: 5 * time.Second},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Error(err)
		}
	})
	tools, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 5 {
		t.Fatalf("unexpected tools: %#v", tools.Tools)
	}
	for _, tool := range tools.Tools {
		if tool.Name == "open_workspace" && tool.Meta["ui"].(map[string]any)["resourceUri"] != appURI {
			t.Fatal("tool did not expose app binding")
		}
	}
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "open_workspace", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || result.StructuredContent.(map[string]any)["message"] != "Welcome to your workspace" {
		t.Fatalf("unexpected tool result: %#v", result)
	}
	resource, err := session.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: appURI})
	if err != nil {
		t.Fatal(err)
	}
	if len(resource.Contents) != 1 || resource.Contents[0].Text != appHTML || resource.Contents[0].MIMEType != ui.MIMEType {
		t.Fatalf("unexpected resource: %#v", resource)
	}
	updated, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "settings.update", Arguments: map[string]any{"set": map[string]any{"showGrid": true}}})
	if err != nil || updated.IsError {
		t.Fatalf("settings update: %#v %v", updated, err)
	}
	read, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "settings.read", Arguments: map[string]any{}})
	if err != nil || read.IsError || read.StructuredContent.(map[string]any)["values"].(map[string]any)["showGrid"] != true {
		t.Fatalf("settings did not persist across stateless requests: %#v %v", read, err)
	}
	mentions, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "search_mentions", Arguments: map[string]any{"query": "bolt"}})
	if err != nil || mentions.IsError || len(mentions.StructuredContent.(map[string]any)["items"].([]any)) != 1 {
		t.Fatalf("mentions: %#v %v", mentions, err)
	}
	part, err := session.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: "parts://bolt"})
	if err != nil || part.Contents[0].Text != "Demo part: bolt" {
		t.Fatalf("mention resource: %#v %v", part, err)
	}
}

func TestHTTPAuthenticationAndLimits(t *testing.T) {
	token := rand.Text() + rand.Text()
	handler, err := newHandler(token)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, authorization, body, origin string
		want                              int
	}{
		{name: "missing credential", body: "{}", want: http.StatusUnauthorized},
		{name: "wrong credential", authorization: "Bearer " + rand.Text(), body: "{}", want: http.StatusUnauthorized},
		{name: "cross origin", authorization: "Bearer " + token, body: "{}", origin: "https://other.example", want: http.StatusForbidden},
		{name: "oversized body", authorization: "Bearer " + token, body: strings.Repeat("x", (1<<20)+1), want: http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/mcp", strings.NewReader(tt.body))
			req.Header.Set("Authorization", tt.authorization)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
				req.Header.Set("Sec-Fetch-Site", "cross-site")
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.want, response.Body.String())
			}
			if strings.Contains(response.Body.String(), token) {
				t.Fatal("response disclosed credential")
			}
		})
	}
	if _, err := newHandler(""); err == nil {
		t.Fatal("missing startup credential accepted")
	}
}

func TestRunStopsOnCancellation(t *testing.T) {
	// Cancel before startup to verify that the listener and Serve goroutine still exit.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := run(ctx, "127.0.0.1:0", rand.Text()+rand.Text()); err != nil {
		t.Fatal(err)
	}
}
