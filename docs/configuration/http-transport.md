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

| Scope | Gates |
|-------|-------|
| `docs:read` | Google Docs tools (`get_document_*`) |
| `sheets:read` | Google Sheets tools (`get_spreadsheet_*`, `list_sheets`, `get_sheet_*`, `get_multiple_ranges`) |
| `slides:read` | Google Slides tools (`get_presentation*`, `list_slides`, `get_slide*`) |

A call to a tool whose required scope isn't present in the token's `scope` claim is denied with an `access denied` error before it reaches the Google API.

## Audit logging

Every `tools/call` request in HTTP mode is logged via the configured `*slog.Logger`, including the verified subject, the RFC 8693 `act` delegation chain (if the token represents a delegated call), the client ID, and the granted scope:

```json
{"msg":"tool_call","tool":"get_document_text","sub":"user@example.com","act":"agent-123","client_id":"mcp-google-client","scope":"docs:read"}
```

Denied calls are logged separately as `tool_call_denied` with the required scope that was missing.

## Google credentials still required

ID-JAG / HTTP mode authorizes *access to mcp-google's tools*. It is independent of how `mcp-google` itself authenticates to the Google APIs — you still need one of the credential sources described in [Credentials](credentials.md) (service account, goauth CredentialsSet, or vault-backed).

## stdio is unaffected

If `--http` / `MCP_GOOGLE_HTTP_ADDR` is not set, `mcp-google` behaves exactly as before: it serves over stdio with no token verification, suitable for local use (e.g. Claude Desktop).
