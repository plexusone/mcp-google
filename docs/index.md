# Google MCP Server

An MCP (Model Context Protocol) server for reading Google Docs, Sheets, and Slides.

## Features

- **Read-only access** to Google Docs, Sheets, and Slides via MCP tools
- **14 tools** for comprehensive document access
- **Composable architecture** built on [omniskill](https://github.com/plexusone/omniskill)
- **Multiple credential options** including vault-backed credentials via [omnitoken](https://github.com/plexusone/omnitoken)

## Google Docs Tools

| Tool | Description |
|------|-------------|
| `get_document_metadata` | Get document metadata (title, word count, element counts) |
| `get_document_content` | Get structured content (headings, paragraphs, images, tables) |
| `get_document_text` | Get all text as a single plain text string |
| `get_document_paragraphs` | Get text organized by paragraphs |

## Google Sheets Tools

| Tool | Description |
|------|-------------|
| `get_spreadsheet_metadata` | Get spreadsheet metadata (title, sheet count, locale, time zone) |
| `list_sheets` | List all sheets with their properties |
| `get_sheet_values` | Get cell values from a specific range |
| `get_sheet_data` | Get all data from a specific sheet |
| `get_multiple_ranges` | Batch get values from multiple ranges |

## Google Slides Tools

| Tool | Description |
|------|-------------|
| `get_presentation` | Get presentation metadata (title, slide count, locale, revision ID) |
| `list_slides` | List all slides with titles and element counts |
| `get_slide` | Get slide content and element details by index or object ID |
| `get_slide_notes` | Get speaker notes by slide index or object ID |
| `get_presentation_content` | Get all slides' text and images in one call (ideal for AI) |

## Why Composable?

This server is built on **omniskill**, making its Google Docs, Sheets, and Slides skills reusable building blocks. You can:

1. **Use standalone** - Run `mcp-google` as a focused Google tools server
2. **Compose with others** - Import the skills into a multi-service MCP server combining Google + Slack + Jira + GitHub

```go
import (
    "github.com/grokify/mcp-google/skills/docs"
    "github.com/grokify/mcp-google/skills/sheets"
    "github.com/grokify/mcp-google/skills/slides"
)

// Add to any omniskill-based server
rt.RegisterSkill(docs.New(httpClient))
rt.RegisterSkill(sheets.New(httpClient))
rt.RegisterSkill(slides.New(httpClient))
```

See [Architecture](architecture/overview.md) for more details.

## Quick Start

```bash
# Install
go install github.com/grokify/mcp-google/cmd/mcp-google@latest

# Run with Google service account
mcp-google --credentials /path/to/service-account.json
```

See [Getting Started](getting-started/installation.md) for full setup instructions.
