package emauth

import (
	"bytes"
	"context"
	"log/slog"
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
	}
	for tool, want := range tests {
		got, ok := RequiredScope(tool)
		if !ok {
			t.Errorf("RequiredScope(%q) reported unmapped, want %q", tool, want)
			continue
		}
		if got != want {
			t.Errorf("RequiredScope(%q) = %q, want %q", tool, got, want)
		}
	}

	// Exact-name matching: names that merely contain a mapped substring
	// must not inherit a scope.
	for _, tool := range []string{"unknown_tool", "get_file_permissions", "delete_file", "get_document", "GET_DOCUMENT_TEXT"} {
		if got, ok := RequiredScope(tool); ok {
			t.Errorf("RequiredScope(%q) = %q, want unmapped", tool, got)
		}
	}
}

func TestToolScopesReturnsCopy(t *testing.T) {
	m := ToolScopes()
	m["get_document_text"] = "tampered"
	if got, ok := RequiredScope("get_document_text"); !ok || got != ScopeDocs {
		t.Fatalf("mutating ToolScopes() result changed the map: got %q, %v", got, ok)
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
	return callViaMiddlewareWithLogger(t, ctx, toolName, nil)
}

func callViaMiddlewareWithLogger(t *testing.T, ctx context.Context, toolName string, logger *slog.Logger) error {
	t.Helper()
	var called bool
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		called = true
		return &mcp.CallToolResult{}, nil
	}
	handler := Middleware(logger)(next)

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

	t.Run("scoped_token_denies_unmapped_tool", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, nil))
		ctx := oauth2.SetTokenInfoContext(context.Background(), &oauth2.TokenInfo{
			Subject: "user:alice",
			Scope:   "docs:read sheets:read slides:read",
		})
		err := callViaMiddlewareWithLogger(t, ctx, "get_file_permissions", logger)
		if err == nil {
			t.Fatal("expected denial for unmapped tool with scoped token")
		}
		if !strings.Contains(err.Error(), "access denied: tool get_file_permissions") {
			t.Errorf("unexpected denial message: %v", err)
		}
		logs := buf.String()
		if !strings.Contains(logs, `"msg":"tool_call_denied"`) || !strings.Contains(logs, `"tool":"get_file_permissions"`) {
			t.Errorf("expected tool_call_denied audit log for unmapped tool, got %s", logs)
		}
	})

	t.Run("scoped_token_missing_scope_is_audit_logged", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, nil))
		ctx := oauth2.SetTokenInfoContext(context.Background(), &oauth2.TokenInfo{
			Subject: "user:alice",
			Scope:   "sheets:read",
		})
		err := callViaMiddlewareWithLogger(t, ctx, "get_slide", logger)
		if err == nil {
			t.Fatal("expected denial for slides tool with sheets-only scope")
		}
		want := "access denied: tool get_slide requires scope " + ScopeSlides
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
		if !strings.Contains(buf.String(), `"required_scope":"slides:read"`) {
			t.Errorf("expected required_scope in denial audit log, got %s", buf.String())
		}
	})

	t.Run("no_identity_unmapped_tool_passes", func(t *testing.T) {
		if err := callViaMiddleware(t, context.Background(), "get_file_permissions"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unscoped_token_unmapped_tool_not_gated", func(t *testing.T) {
		ctx := oauth2.SetTokenInfoContext(context.Background(), &oauth2.TokenInfo{
			Subject: "user:bob",
		})
		if err := callViaMiddleware(t, ctx, "get_file_permissions"); err != nil {
			t.Fatalf("unexpected error: %v", err)
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
