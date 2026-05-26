# Google MCP Server

An MCP (Model Context Protocol) server for reading Google Slides presentations and Google Docs documents.

## Features

- **Read-only access** to Google Slides and Google Docs via MCP tools
- **9 tools** for comprehensive document access
- **Composable architecture** built on [omniskill](https://github.com/plexusone/omniskill)
- **Multiple credential options** including vault-backed credentials via [omnitoken](https://github.com/plexusone/omnitoken)

## Google Slides Tools

| Tool | Description |
|------|-------------|
| `get_presentation` | Get presentation metadata (title, slide count, locale, revision ID) |
| `list_slides` | List all slides with titles and element counts |
| `get_slide` | Get slide content and element details by index or object ID |
| `get_slide_notes` | Get speaker notes by slide index or object ID |
| `get_presentation_content` | Get all slides' text and images in one call (ideal for AI) |

## Google Docs Tools

| Tool | Description |
|------|-------------|
| `get_document_metadata` | Get document metadata (title, word count, element counts) |
| `get_document_content` | Get structured content (headings, paragraphs, images, tables) |
| `get_document_text` | Get all text as a single plain text string |
| `get_document_paragraphs` | Get text organized by paragraphs |

## Why Composable?

This server is built on **omniskill**, making its Google Slides and Docs skills reusable building blocks. You can:

1. **Use standalone** - Run `mcp-google` as a focused Google tools server
2. **Compose with others** - Import the skills into a multi-service MCP server combining Google + Slack + Jira + GitHub

```go
import (
    "github.com/grokify/mcp-google/skills/slides"
    "github.com/grokify/mcp-google/skills/docs"
)

// Add to any omniskill-based server
rt.RegisterSkill(slides.New(httpClient))
rt.RegisterSkill(docs.New(httpClient))
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
