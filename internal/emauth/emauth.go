// Package emauth provides MCP Enterprise-Managed Authorization support for
// mcp-google: audit logging and scope-based tool gating driven by the
// identity verified from an ID-JAG-derived access token.
//
// The identity (subject, RFC 8693 act delegation chain, scope) is placed in
// the request context by omniskill's ExternalBearerMiddleware; this package
// consumes it at the MCP method layer via a receiving middleware.
package emauth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plexusone/omniskill/mcp/oauth2"
)

// Scopes understood by mcp-google. When the verified token carries a scope
// claim, every tool call requires the tool's mapped scope; tokens without a
// scope claim are not gated (identity is still audit-logged).
const (
	ScopeDocs   = "docs:read"
	ScopeSheets = "sheets:read"
	ScopeSlides = "slides:read"
)

// toolScopes is the authorization map: every registered tool, keyed by its
// exact name, maps to exactly one required policy scope. A tool that is not
// listed here is denied whenever the token carries a scope claim, so a new
// tool must be added to this table before scoped callers can use it.
var toolScopes = map[string]string{
	// Google Docs
	"get_document_metadata":    ScopeDocs,
	"get_document_content":     ScopeDocs,
	"get_document_text":        ScopeDocs,
	"get_document_paragraphs":  ScopeDocs,
	"get_document_suggestions": ScopeDocs,

	// Google Sheets
	"get_spreadsheet_metadata": ScopeSheets,
	"list_sheets":              ScopeSheets,
	"get_sheet_values":         ScopeSheets,
	"get_sheet_data":           ScopeSheets,
	"get_multiple_ranges":      ScopeSheets,

	// Google Slides
	"get_presentation":         ScopeSlides,
	"list_slides":              ScopeSlides,
	"get_slide":                ScopeSlides,
	"get_slide_notes":          ScopeSlides,
	"get_presentation_content": ScopeSlides,
}

// RequiredScope returns the policy scope that authorizes the tool with the
// given exact name, and whether the tool is mapped at all. Unmapped tools
// report ok == false and are denied when a scope claim is present.
func RequiredScope(toolName string) (scope string, ok bool) {
	scope, ok = toolScopes[toolName]
	return scope, ok
}

// ToolScopes returns a copy of the tool → scope authorization map.
func ToolScopes() map[string]string {
	out := make(map[string]string, len(toolScopes))
	for tool, scope := range toolScopes {
		out[tool] = scope
	}
	return out
}

// HasScope reports whether the space-delimited scope string contains the
// required scope.
func HasScope(scopes, required string) bool {
	for _, s := range strings.Fields(scopes) {
		if s == required {
			return true
		}
	}
	return false
}

// Middleware returns an MCP receiving middleware that audit-logs the
// verified identity on every tool call and, when the token carries a scope
// claim, denies calls to unmapped tools and calls whose token scope does not
// cover the tool's mapped scope. Non-tool methods pass through untouched.
func Middleware(logger *slog.Logger) mcp.Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method != "tools/call" {
				return next(ctx, method, req)
			}

			toolName := ""
			if callReq, ok := req.(*mcp.CallToolRequest); ok && callReq.Params != nil {
				toolName = callReq.Params.Name
			}

			info := oauth2.GetTokenInfoContext(ctx)
			if info == nil {
				// No verified identity in context (e.g. stdio transport):
				// nothing to gate or audit.
				return next(ctx, method, req)
			}

			logger.Info("tool_call",
				"tool", toolName,
				"sub", info.Subject,
				"act", info.Actor,
				"client_id", info.ClientID,
				"scope", info.Scope,
			)

			if info.Scope != "" {
				required, ok := RequiredScope(toolName)
				if !ok {
					logger.Warn("tool_call_denied",
						"tool", toolName,
						"sub", info.Subject,
						"act", info.Actor,
						"reason", "unmapped_tool",
					)
					return nil, fmt.Errorf("access denied: tool %s has no scope mapping", toolName)
				}
				if !HasScope(info.Scope, required) {
					logger.Warn("tool_call_denied",
						"tool", toolName,
						"sub", info.Subject,
						"act", info.Actor,
						"required_scope", required,
					)
					return nil, fmt.Errorf("access denied: tool %s requires scope %s", toolName, required)
				}
			}

			return next(ctx, method, req)
		}
	}
}
