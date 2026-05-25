# Release Notes v0.1.0

Initial release of google-mcp-server, a read-only MCP (Model Context Protocol) server for Google Slides presentations.

## Highlights

- Read-only MCP server enabling AI assistants to read and understand Google Slides presentations
- 5 tools for comprehensive slide content extraction
- Dual authentication support: Google service account and goauth CredentialsSet

## MCP Tools

| Tool | Description |
|------|-------------|
| `get_presentation` | Get presentation metadata (title, slide count, locale, revision) |
| `list_slides` | List all slides with titles and element counts |
| `get_slide` | Get content and elements for a specific slide |
| `get_slide_notes` | Get speaker notes for a specific slide |
| `get_presentation_content` | Get all slides' text and images in one call (ideal for AI analysis) |

## Authentication

Two authentication methods are supported:

1. **Google Service Account**: Use `-credentials /path/to/service-account.json`
2. **goauth CredentialsSet**: Use `-goauth-credentials-file` and `-goauth-credentials-account` flags

## Usage with Claude Desktop

```json
{
  "mcpServers": {
    "google-slides": {
      "command": "/path/to/google-mcp-server",
      "args": ["-credentials", "/path/to/service-account.json"]
    }
  }
}
```

## Dependencies

- `github.com/grokify/gogoogle` v0.7.0 - Google Slides reading utilities
- `github.com/grokify/goauth` v0.23.28 - OAuth2 authentication
- `github.com/modelcontextprotocol/go-sdk` v1.2.0 - MCP SDK for Go
- `google.golang.org/api` v0.265.0 - Google APIs
