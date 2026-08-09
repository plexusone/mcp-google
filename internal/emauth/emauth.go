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
// claim, tool calls require the matching skill scope; tokens without a
// scope claim are not gated (identity is still audit-logged).
const (
	ScopeDocs   = "docs:read"
	ScopeSheets = "sheets:read"
	ScopeSlides = "slides:read"
)

// RequiredScope maps a tool name to the scope that authorizes it. Returns
// an empty string for tools with no scope mapping (allowed by default).
func RequiredScope(toolName string) string {
	switch {
	case strings.Contains(toolName, "document"):
		return ScopeDocs
	case strings.Contains(toolName, "spreadsheet"), strings.Contains(toolName, "sheet"):
		return ScopeSheets
	case strings.Contains(toolName, "presentation"), strings.Contains(toolName, "slide"):
		return ScopeSlides
	default:
		return ""
	}
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
// verified identity on every tool call and denies calls whose token scope
// does not cover the tool. Non-tool methods pass through untouched.
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
				if required := RequiredScope(toolName); required != "" && !HasScope(info.Scope, required) {
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
