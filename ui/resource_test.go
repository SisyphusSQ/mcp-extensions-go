package ui_test

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/ui"
)

func TestHTMLResourceThroughOfficialSDK(t *testing.T) {
	ctx := t.Context()
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	resource := &mcp.Resource{URI: "ui://app/home.html", Name: "home", Meta: mcp.Meta{"vendor": map[string]any{"value": "original"}}}
	if err := ui.AddHTMLResource(server, resource, "<h1>你好</h1>", ui.ResourceMetadata{PreferredDisplayMode: ui.Fullscreen}); err != nil {
		t.Fatal(err)
	}
	resource.URI = "ui://app/changed.html"
	resource.Meta["vendor"].(map[string]any)["value"] = "changed"
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ss.Close(); err != nil {
			t.Error(err)
		}
	})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cs.Close(); err != nil {
			t.Error(err)
		}
	})
	listed, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Resources) != 1 || listed.Resources[0].URI != "ui://app/home.html" || listed.Resources[0].MIMEType != ui.MIMEType {
		t.Fatalf("unexpected resources: %#v", listed.Resources)
	}
	read, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "ui://app/home.html"})
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Contents) != 1 {
		t.Fatalf("contents count = %d", len(read.Contents))
	}
	content := read.Contents[0]
	if content.Text != "<h1>你好</h1>" || content.MIMEType != ui.MIMEType || content.URI != "ui://app/home.html" {
		t.Fatalf("unexpected content: %#v", content)
	}
	if content.Meta["vendor"].(map[string]any)["value"] != "original" || content.Meta[ui.OpenAIMetaKey].(map[string]any)["preferredDisplayMode"] != "fullscreen" {
		t.Fatalf("unexpected content metadata: %#v", content.Meta)
	}
}

func TestHTMLResourceRejectsInvalidRegistration(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	for _, resource := range []*mcp.Resource{
		nil, {URI: "file:///etc/passwd", Name: "bad"}, {URI: "ui://app/home", Name: ""},
		{URI: "ui://app/home", Name: "home", MIMEType: "text/plain"},
	} {
		if err := ui.AddHTMLResource(server, resource, "<h1>home</h1>", ui.ResourceMetadata{}); err == nil {
			t.Fatalf("invalid resource accepted: %#v", resource)
		}
	}
}
