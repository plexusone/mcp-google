# CLI Reference

The `mcp-google` binary can run as an MCP server or call the same Google Slides and Docs tools directly from the command line.

## Discover Commands

```bash
mcp-google --help
mcp-google <command> --help
```

The CLI is generated with Cobra, so `--help` is the easiest way to reference the current command list and flags.

## Server Commands

| Command | Description |
|---------|-------------|
| `mcp-google` | Start the MCP server using stdio |
| `mcp-google serve` | Start the MCP server explicitly |
| `mcp-google version` | Print version information |

## Google Slides Commands

| Command | MCP tool |
|---------|----------|
| `get-presentation <presentation-id>` | `get_presentation` |
| `list-slides <presentation-id>` | `list_slides` |
| `get-slide <presentation-id>` | `get_slide` |
| `get-slide-notes <presentation-id>` | `get_slide_notes` |
| `get-presentation-content <presentation-id>` | `get_presentation_content` |

## Google Docs Commands

| Command | MCP tool |
|---------|----------|
| `get-document-metadata <document-id-or-url>` | `get_document_metadata` |
| `get-document-content <document-id-or-url>` | `get_document_content` |
| `get-document-text <document-id-or-url>` | `get_document_text` |
| `get-document-paragraphs <document-id-or-url>` | `get_document_paragraphs` |

## Global Flags

| Flag | Environment variable | Description |
|------|----------------------|-------------|
| `--credentials` | `GOOGLE_CREDENTIALS_FILE` | Google service account credentials JSON file |
| `--goauth-credentials-file` | `GOAUTH_CREDENTIALS_FILE` | goauth CredentialsSet JSON file |
| `--goauth-credentials-account` | `GOAUTH_CREDENTIALS_ACCOUNT` | Account key in the goauth CredentialsSet |
| `--vault` | `OMNITOKEN_VAULT_URI` | Vault URI for credentials |
| `--credentials-name` | `OMNITOKEN_CREDENTIALS_NAME` | Credential name in the vault |
| `-o, --output` | | Output format: `json` or `pretty` |

## Examples

```bash
mcp-google get-document-metadata https://docs.google.com/document/d/abc123/edit \
  --credentials /path/to/service-account.json

mcp-google get-document-content abc123 \
  --include-metadata \
  --include-images \
  --include-tables \
  --include-headers \
  --include-footers \
  -o pretty

mcp-google get-slide presentation123 --index 0 -o pretty
```
