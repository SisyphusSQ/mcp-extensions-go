package forms

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// CapabilityKey identifies the client's OpenAI form capability and schema metadata.
const CapabilityKey = "openai/elicitation"

// ErrUnsupportedClient means the request lacks modern extended-form support.
var ErrUnsupportedClient = errors.New("client does not support OpenAI form MRTR")

// RequestOptions identifies a form round. RequestState is opaque caller-owned
// continuation state, echoed by the client; it is not authenticated by this helper.
// The caller must authorize every request and validate any continuation before use.
type RequestOptions struct {
	Key             string
	Message         string
	RequestState    string
	Meta            mcp.Meta
	SelectionPolicy SelectionPolicy
}

// RequestInput returns either an input-required tool result or a validated answer
// for Key from the client's retry. It requires MCP 2026-07-28 or later and both
// elicitation.form and extensions["openai/elicitation"].form. It never sends a
// legacy request, stores state, reads resources, or changes submitted answers.
// Other response keys may belong to other rounds and are left to the caller.
func (f *Form) RequestInput(ctx context.Context, req *mcp.CallToolRequest, options RequestOptions) (*mcp.CallToolResult, *Answer, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if f == nil || req == nil || req.Params == nil || strings.TrimSpace(options.Key) == "" || strings.TrimSpace(options.Message) == "" {
		return nil, nil, fmt.Errorf("%w: form, tool request, key and message are required", ErrInvalidDefinition)
	}
	caps := req.ClientCapabilities()
	if req.ProtocolVersion() < "2026-07-28" || caps == nil || caps.Elicitation == nil || caps.Elicitation.Form == nil {
		return nil, nil, ErrUnsupportedClient
	}
	extension, err := wire.Convert[map[string]any](caps.Extensions[CapabilityKey])
	if err != nil || extension == nil {
		return nil, nil, ErrUnsupportedClient
	}
	form, ok := extension["form"].(map[string]any)
	if !ok || form == nil {
		return nil, nil, ErrUnsupportedClient
	}
	if response, exists := req.Params.InputResponses[options.Key]; exists {
		answer, ok := response.(*mcp.ElicitResult)
		if !ok {
			return nil, nil, fmt.Errorf("%w: expected an elicitation answer for %q", ErrInvalidAnswer, options.Key)
		}
		if err := f.ValidateResult(answer, options.SelectionPolicy); err != nil {
			return nil, nil, err
		}
		return nil, answer, nil
	}
	meta, err := wire.Clone(options.Meta)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: form metadata: %v", ErrInvalidDefinition, err)
	}
	if meta == nil {
		meta = make(mcp.Meta)
	}
	meta[CapabilityKey] = map[string]any{"requestedSchema": f.Schema()}
	return &mcp.CallToolResult{
		RequestState: options.RequestState,
		InputRequests: mcp.InputRequestMap{options.Key: &mcp.ElicitParams{
			Mode: "form", Message: options.Message,
			RequestedSchema: map[string]any{"type": "object", "properties": map[string]any{}},
			Meta:            meta,
		}},
	}, nil, nil
}
