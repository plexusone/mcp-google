# Tools Overview

Google MCP Server provides 14 tools for reading Google Docs, Sheets, and Slides.

## Tool Categories

| Category | Tools | Description |
|----------|-------|-------------|
| [Google Docs](docs.md) | 4 tools | Read documents, content, and paragraphs |
| [Google Sheets](sheets.md) | 5 tools | Read spreadsheets, sheets, and cell values |
| [Google Slides](slides.md) | 5 tools | Read presentations, slides, and speaker notes |

## Tools at a Glance

### Google Docs

| Tool | Purpose |
|------|---------|
| `get_document_metadata` | Document metadata (title, word count) |
| `get_document_content` | Structured content with optional images/tables |
| `get_document_text` | Plain text extraction |
| `get_document_paragraphs` | Text organized by paragraphs |

### Google Sheets

| Tool | Purpose |
|------|---------|
| `get_spreadsheet_metadata` | Spreadsheet metadata (title, sheet count) |
| `list_sheets` | List all sheets with properties |
| `get_sheet_values` | Cell values from a specific range |
| `get_sheet_data` | All data from a specific sheet |
| `get_multiple_ranges` | Batch get values from multiple ranges |

### Google Slides

| Tool | Purpose |
|------|---------|
| `get_presentation` | Presentation metadata (title, slide count) |
| `list_slides` | List all slides with titles |
| `get_slide` | Single slide content and elements |
| `get_slide_notes` | Speaker notes for a slide |
| `get_presentation_content` | All content in one call (ideal for AI) |

## Common Patterns

### AI Analysis of Entire Document

For AI analysis, use the bulk retrieval tools:

- **Docs**: `get_document_content` with all include flags
- **Sheets**: `get_sheet_data` with `include_metadata: true`
- **Slides**: `get_presentation_content` with `include_notes: true`

### Targeted Retrieval

For specific content:

- **Docs**: Use `get_document_paragraphs` for structured text
- **Sheets**: Use `get_sheet_values` for specific ranges
- **Slides**: Use `list_slides` then `get_slide` for specific slides

### Metadata Only

For quick metadata checks:

- **Docs**: `get_document_metadata`
- **Sheets**: `get_spreadsheet_metadata`
- **Slides**: `get_presentation`
