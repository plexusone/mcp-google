# HTTP Transport & ID-JAG Authorization

By default `mcp-google` serves MCP over stdio, with no network exposure. As of v0.4.0 it can also serve over streamable HTTP, secured as a pure OAuth 2.1 resource server that verifies externally-issued bearer tokens (an enterprise IdP, or an [ID-JAG](https://github.com/aistandardsio/agent-protocols) delegation-chain token) instead of running its own authorization server.

This is **Enterprise-Managed Authorization**: your identity provider issues the tokens and owns the trust relationship; `mcp-google` only verifies them and enforces scope.

## Enabling HTTP mode

HTTP mode is opt-in. Passing `--http` (or setting `MCP_GOOGLE_HTTP_ADDR`) switches the `serve` command from stdio to streamable HTTP:

```bash
mcp-google serve \
  --http :8080 \
  --idjag-issuer https://idp.example.com \
  --idjag-audience https://mcp-google.example.com
```

`--idjag-issuer` is required in HTTP mode — it's the authorization server whose tokens will be accepted. `--idjag-audience` should always be set in production; without it, token audience is not validated and a token minted for a different resource would be accepted.

## Flags

| Flag | Environment variable | Description |
|------|----------------------|-------------|
| `--http` | `MCP_GOOGLE_HTTP_ADDR` | Serve MCP over HTTP on this address (e.g. `:8080`) instead of stdio |
| `--idjag-issuer` | `MCP_GOOGLE_IDJAG_ISSUER` | Authorization server issuer URL for ID-JAG-derived access tokens (required for HTTP mode) |
| `--idjag-audience` | `MCP_GOOGLE_IDJAG_AUDIENCE` | Expected token audience / resource identifier |
| `--idjag-jwks-url` | `MCP_GOOGLE_IDJAG_JWKS_URL` | Pin the issuer's JWKS endpoint, skipping discovery |

## Scopes

Verified tokens that carry a `scope` claim are enforced per tool call. Tokens without a `scope` claim are not gated — the call is allowed, and the identity is still audit-logged.

Authorization uses an explicit map keyed by exact tool name. Every registered tool has exactly one required scope:

| Scope | Tools |
|-------|-------|
| `docs:read` | `get_document_metadata`, `get_document_content`, `get_document_text`, `get_document_paragraphs` |
| `sheets:read` | `get_spreadsheet_metadata`, `list_sheets`, `get_sheet_values`, `get_sheet_data`, `get_multiple_ranges` |
| `slides:read` | `get_presentation`, `list_slides`, `get_slide`, `get_slide_notes`, `get_presentation_content` |
| `drive:read` | `get_file_comments` |
| `drive:write` | `upload_file`, `update_file` (registered only with `--enable-write`) |

!!! warning "Authorization change in v0.5.0"
    `get_document_suggestions` needs `docs:read`, like the other Docs tools. The
    new `get_file_comments` needs `drive:read`, and the upload tools need
    `drive:write`. A deployment that issues scoped tokens must add `drive:read`
    and, if it enables write mode, `drive:write` before callers can use them;
    until then those calls are denied with a message naming the missing scope.
    Tokens with no `scope` claim are not gated, as before. `drive:write` is
    advertised in the resource metadata only when write mode is on.

When the token carries a `scope` claim, a call is denied with an `access denied` error before it reaches the Google API if:

- the token's `scope` claim does not include the tool's required scope (`access denied: tool <name> requires scope <scope>`), or
- the tool has no entry in the map (`access denied: tool <name> has no scope mapping`). Unmapped tools are denied by default rather than allowed.

## Audit logging

Every `tools/call` request in HTTP mode is logged via the configured `*slog.Logger`, including the verified subject, the RFC 8693 `act` delegation chain (if the token represents a delegated call), the client ID, and the granted scope:

```json
{"msg":"tool_call","tool":"get_document_text","sub":"user@example.com","act":"agent-123","client_id":"mcp-google-client","scope":"docs:read"}
```

Denied calls are logged separately as `tool_call_denied`, with the `required_scope` that was missing, or with `reason` set to `unmapped_tool` when the tool has no scope mapping.

## Google credentials still required

ID-JAG / HTTP mode authorizes *access to mcp-google's tools*. It is independent of how `mcp-google` itself authenticates to the Google APIs — you still need one of the credential sources described in [Credentials](credentials.md) (service account, goauth CredentialsSet, or vault-backed).

## stdio is unaffected

If `--http` / `MCP_GOOGLE_HTTP_ADDR` is not set, `mcp-google` behaves exactly as before: it serves over stdio with no token verification, suitable for local use (e.g. Claude Desktop).
