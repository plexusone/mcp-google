# Tools Overview

Google MCP Server provides 9 tools for reading Google Slides and Google Docs.

## Tool Categories

| Category | Tools | Description |
|----------|-------|-------------|
| [Google Slides](slides.md) | 5 tools | Read presentations, slides, and speaker notes |
| [Google Docs](docs.md) | 4 tools | Read documents, content, and paragraphs |

## Tools at a Glance

### Google Slides

| Tool | Purpose |
|------|---------|
| `get_presentation` | Presentation metadata (title, slide count) |
| `list_slides` | List all slides with titles |
| `get_slide` | Single slide content and elements |
| `get_slide_notes` | Speaker notes for a slide |
| `get_presentation_content` | All content in one call (ideal for AI) |

### Google Docs

| Tool | Purpose |
|------|---------|
| `get_document_metadata` | Document metadata (title, word count) |
| `get_document_content` | Structured content with optional images/tables |
| `get_document_text` | Plain text extraction |
| `get_document_paragraphs` | Text organized by paragraphs |

## Common Patterns

### AI Analysis of Entire Document

For AI analysis, use the bulk retrieval tools:

- **Slides**: `get_presentation_content` with `include_notes: true`
- **Docs**: `get_document_content` with all include flags

### Targeted Retrieval

For specific content:

- **Slides**: Use `list_slides` then `get_slide` for specific slides
- **Docs**: Use `get_document_paragraphs` for structured text

### Metadata Only

For quick metadata checks:

- **Slides**: `get_presentation`
- **Docs**: `get_document_metadata`
