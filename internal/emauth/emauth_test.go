package emauth

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plexusone/omniskill/mcp/oauth2"
)

func TestRequiredScope(t *testing.T) {
	tests := map[string]string{
		"get_document_text":        ScopeDocs,
		"get_document_metadata":    ScopeDocs,
		"get_spreadsheet_metadata": ScopeSheets,
		"get_sheet_values":         ScopeSheets,
		"get_presentation":         ScopeSlides,
		"get_slide_notes":          ScopeSlides,
		"unknown_tool":             "",
	}
	for tool, want := range tests {
		if got := RequiredScope(tool); got != want {
			t.Errorf("RequiredScope(%q) = %q, want %q", tool, got, want)
		}
	}
}

func TestHasScope(t *testing.T) {
	if !HasScope("docs:read sheets:read", "docs:read") {
		t.Error("expected docs:read to be found")
	}
	if HasScope("docs:read", "slides:read") {
		t.Error("did not expect slides:read to be found")
	}
	if HasScope("", "docs:read") {
		t.Error("empty scope string should not match")
	}
}

func callViaMiddleware(t *testing.T, ctx context.Context, toolName string) error {
	t.Helper()
	var called bool
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		called = true
		return &mcp.CallToolResult{}, nil
	}
	handler := Middleware(nil)(next)

	req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: toolName}}
	_, err := handler(ctx, "tools/call", req)
	if err == nil && !called {
		t.Fatal("next handler not called despite no error")
	}
	return err
}

func TestMiddleware(t *testing.T) {
	t.Run("no_identity_passes", func(t *testing.T) {
		if err := callViaMiddleware(t, context.Background(), "get_document_text"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("scoped_token_allows_matching_tool", func(t *testing.T) {
		ctx := oauth2.SetTokenInfoContext(context.Background(), &oauth2.TokenInfo{
			Subject: "user:alice",
			Actor:   []string{"agent:worker"},
			Scope:   "docs:read",
		})
		if err := callViaMiddleware(t, ctx, "get_document_text"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("scoped_token_denies_other_tool", func(t *testing.T) {
		ctx := oauth2.SetTokenInfoContext(context.Background(), &oauth2.TokenInfo{
			Subject: "user:alice",
			Scope:   "docs:read",
		})
		err := callViaMiddleware(t, ctx, "get_sheet_values")
		if err == nil {
			t.Fatal("expected denial for sheets tool with docs-only scope")
		}
		if !strings.Contains(err.Error(), ScopeSheets) {
			t.Errorf("expected required scope in error, got %v", err)
		}
	})

	t.Run("unscoped_token_not_gated", func(t *testing.T) {
		ctx := oauth2.SetTokenInfoContext(context.Background(), &oauth2.TokenInfo{
			Subject: "user:bob",
		})
		if err := callViaMiddleware(t, ctx, "get_sheet_values"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
