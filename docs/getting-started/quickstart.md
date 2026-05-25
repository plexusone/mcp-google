# Quick Start

## Running the Server

### With Google Service Account

```bash
mcp-google --credentials /path/to/service-account.json
```

### With Environment Variable

```bash
export GOOGLE_CREDENTIALS_FILE=/path/to/service-account.json
mcp-google
```

## Finding Document IDs

### Presentations

The presentation ID is in the URL:

```
https://docs.google.com/presentation/d/PRESENTATION_ID_HERE/edit
```

### Documents

The document ID is in the URL:

```
https://docs.google.com/document/d/DOCUMENT_ID_HERE/edit
```

!!! note "URLs Supported"
    Google Docs tools accept either the document ID or the full URL, including URLs with query strings and anchors:
    ```
    https://docs.google.com/document/d/DOCUMENT_ID_HERE/edit?tab=t.0#heading=h.xyz
    ```

## Example Tool Calls

Once the server is running with an MCP client, you can call tools:

### Get Presentation Metadata

```json
{
  "name": "get_presentation",
  "arguments": {
    "presentation_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
  }
}
```

### Get Document Text

```json
{
  "name": "get_document_text",
  "arguments": {
    "document_id": "https://docs.google.com/document/d/1abc123/edit"
  }
}
```

### Get All Slide Content with Notes

```json
{
  "name": "get_presentation_content",
  "arguments": {
    "presentation_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
    "include_notes": true
  }
}
```

## Next Steps

- [Tools Reference](../tools/overview.md) - Full documentation for all 9 tools
- [Configuration](../configuration/credentials.md) - All credential options
- [Claude Desktop](../configuration/claude-desktop.md) - Integration with Claude Desktop
