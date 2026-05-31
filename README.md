# Google MCP Server

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/mcp-google/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/mcp-google/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/mcp-google/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/mcp-google/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/mcp-google/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/mcp-google/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/mcp-google
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/mcp-google
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/mcp-google
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/mcp-google
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/mcp-google
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fmcp-google
 [loc-svg]: https://tokei.rs/b1/github/plexusone/mcp-google
 [repo-url]: https://github.com/plexusone/mcp-google
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/mcp-google/blob/main/LICENSE

An MCP (Model Context Protocol) server for reading Google Docs, Sheets, and Slides.

## Features

Read-only access to Google Docs, Sheets, and Slides via MCP tools.

### Google Docs Tools

- **get_document_metadata** - Get document metadata (title, word count, element counts)
- **get_document_content** - Get structured content (headings, paragraphs, images, tables)
- **get_document_text** - Get all text as a single plain text string
- **get_document_paragraphs** - Get text organized by paragraphs

### Google Sheets Tools

- **get_spreadsheet_metadata** - Get spreadsheet metadata (title, sheet count, locale, time zone)
- **list_sheets** - List all sheets with their properties
- **get_sheet_values** - Get cell values from a specific range
- **get_sheet_data** - Get all data from a specific sheet
- **get_multiple_ranges** - Batch get values from multiple ranges

### Google Slides Tools

- **get_presentation** - Get presentation metadata (title, slide count, locale, revision ID)
- **list_slides** - List all slides with titles and element counts
- **get_slide** - Get slide content and element details by index or object ID
- **get_slide_notes** - Get speaker notes by slide index or object ID
- **get_presentation_content** - Get all slides' text and images in one call (ideal for AI)

## Architecture

This server is built on [omniskill](https://github.com/plexusone/omniskill), making its Google Docs, Sheets, and Slides skills **composable building blocks** that can be reused in multi-service MCP servers.

### Composable Skills

The skills in this repository can be imported and combined with other skills:

```go
import (
    "github.com/grokify/mcp-google/skills/docs"
    "github.com/grokify/mcp-google/skills/sheets"
    "github.com/grokify/mcp-google/skills/slides"
    runtime "github.com/plexusone/omniskill/mcp/server"
)

// Create runtime
rt := runtime.New(&mcp.Implementation{
    Name:    "work-mcp-server",
    Version: "v1.0.0",
}, nil)

// Add Google skills
docsSkill := docs.New(googleHTTPClient)
docsSkill.Init(ctx)
rt.RegisterSkill(docsSkill)

sheetsSkill := sheets.New(googleHTTPClient)
sheetsSkill.Init(ctx)
rt.RegisterSkill(sheetsSkill)

slidesSkill := slides.New(googleHTTPClient)
slidesSkill.Init(ctx)
rt.RegisterSkill(slidesSkill)

// Add other skills (Slack, Jira, GitHub, etc.)
rt.RegisterSkill(slackSkill)
rt.RegisterSkill(jiraSkill)

// Run server
rt.ServeStdio(ctx)
```

This enables building unified MCP servers that combine multiple services while keeping each service's implementation modular and maintainable.

## Requirements

- Go 1.24+
- Google Cloud service account with Docs, Sheets, and Slides API access

## Installation

```bash
go install github.com/grokify/mcp-google/cmd/mcp-google@latest
```

Or build from source:

```bash
git clone https://github.com/grokify/mcp-google.git
cd mcp-google
go build ./cmd/mcp-google
```

## Setup

### 1. Create a Google Cloud Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Docs API, Google Sheets API, and Google Slides API
4. Create a service account with no special roles
5. Download the JSON credentials file

### 2. Share Documents with the Service Account

Share any documents, spreadsheets, or presentations you want to access with the service account's email address (found in the credentials JSON as `client_email`).

## Usage

### Option 1: Google Service Account Credentials

Use a standard Google Cloud service account JSON file:

```bash
mcp-google --credentials /path/to/service-account.json
```

Or using an environment variable:

```bash
export GOOGLE_CREDENTIALS_FILE=/path/to/service-account.json
mcp-google
```

### Option 2: goauth CredentialsSet

Use a [goauth](https://github.com/grokify/goauth) CredentialsSet file, which can store multiple credentials:

```bash
mcp-google --goauth-credentials-file /path/to/credentials.json --goauth-credentials-account myaccount
```

Or using environment variables:

```bash
export GOAUTH_CREDENTIALS_FILE=/path/to/credentials.json
export GOAUTH_CREDENTIALS_ACCOUNT=myaccount
mcp-google
```

The CredentialsSet entry should be of type `gcpsa` with appropriate scopes:

```json
{
  "credentials": {
    "myaccount": {
      "type": "gcpsa",
      "gcpsa": {
        "gcpCredentials": {
          "type": "service_account",
          "project_id": "...",
          "private_key_id": "...",
          "private_key": "...",
          "client_email": "...",
          "client_id": "..."
        },
        "scopes": [
          "https://www.googleapis.com/auth/documents.readonly",
          "https://www.googleapis.com/auth/spreadsheets.readonly",
          "https://www.googleapis.com/auth/presentations.readonly",
          "https://www.googleapis.com/auth/drive.readonly"
        ]
      }
    }
  }
}
```

### Option 3: Vault-Backed Credentials

Use [omnitoken](https://github.com/plexusone/omnitoken) with [omnivault-desktop](https://github.com/plexusone/omnivault-desktop) for secure credential storage in password managers.

Supported vault providers:

| Provider | URI Pattern | Requirements |
|----------|-------------|--------------|
| 1Password | `op://vault` | `OP_SERVICE_ACCOUNT_TOKEN` env var |
| Bitwarden | `bw://org-id` | `BW_ACCESS_TOKEN` and `BW_ORGANIZATION_ID` env vars |
| File | `file:///path` | None |
| Env | `env://PREFIX_` | None |

#### 1Password Example

```bash
export OP_SERVICE_ACCOUNT_TOKEN="ops_..."
mcp-google --vault op://MyVault --credentials-name google
```

#### Bitwarden Example

```bash
export BW_ACCESS_TOKEN="..."
export BW_ORGANIZATION_ID="..."
mcp-google --vault bw://org-id --credentials-name google
```

#### File Vault Example

```bash
mcp-google --vault file:///path/to/secrets --credentials-name google
```

### CLI Tool Commands

The CLI exposes one subcommand per MCP tool, plus `serve` and `version`.

```bash
mcp-google --help
mcp-google get-document-metadata <document-id-or-url> --credentials /path/to/service-account.json
mcp-google get-document-content <document-id-or-url> --include-metadata --include-images --include-tables -o pretty
mcp-google get-spreadsheet-metadata <spreadsheet-id-or-url> --credentials /path/to/service-account.json
mcp-google get-sheet-data <spreadsheet-id-or-url> --sheet-name "Sheet1" -o pretty
mcp-google get-presentation <presentation-id> --credentials /path/to/service-account.json
```

Use `mcp-google <command> --help` for command-specific flags. See [docs/cli.md](docs/cli.md) for the full CLI reference.

### Claude Desktop Configuration

Add to your Claude Desktop configuration (`claude_desktop_config.json`):

#### With Google Service Account

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "GOOGLE_CREDENTIALS_FILE": "/path/to/service-account.json"
      }
    }
  }
}
```

#### With goauth CredentialsSet

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "GOAUTH_CREDENTIALS_FILE": "/path/to/credentials.json",
        "GOAUTH_CREDENTIALS_ACCOUNT": "myaccount"
      }
    }
  }
}
```

#### With 1Password

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "OP_SERVICE_ACCOUNT_TOKEN": "ops_...",
        "OMNITOKEN_VAULT_URI": "op://MyVault",
        "OMNITOKEN_CREDENTIALS_NAME": "google"
      }
    }
  }
}
```

See [docs/configuration/claude-desktop.md](docs/configuration/claude-desktop.md) for more options including Bitwarden.

## Google Docs Tools

### get_document_metadata

Get metadata about a document.

**Input:**

- `document_id` (required) - The ID or URL of the Google Doc

**Output:**

- `title` - Document title
- `document_id` - Document ID
- `revision_id` - Current revision ID
- `word_count` - Approximate word count
- `char_count` - Character count
- `image_count` - Number of images
- `table_count` - Number of tables
- `header_count` - Number of headers
- `footer_count` - Number of footers

### get_document_content

Get the full structured content of a document.

**Input:**

- `document_id` (required) - The ID or URL of the Google Doc
- `include_images` (optional) - Include image information (default: false)
- `include_tables` (optional) - Include table content (default: false)
- `include_headers` (optional) - Include document headers (default: false)
- `include_footers` (optional) - Include document footers (default: false)

**Output:**

- `title` - Document title
- `sections` - Array of content sections:
  - `type` - Section type ("heading", "paragraph")
  - `level` - Heading level (1-6, for headings only)
  - `text` - Section text content
  - `style_id` - Style identifier (e.g., "HEADING_1", "NORMAL_TEXT")
- `images` - Array of images (if requested):
  - `object_id` - Image element ID
  - `content_uri` - Direct URL to image
  - `source_uri` - Original source URL
  - `title` - Image title
  - `description` - Image description
- `tables` - Array of tables (if requested):
  - `rows` - Number of rows
  - `columns` - Number of columns
  - `cells` - 2D array of cell text content
- `headers` - Array of header text (if requested)
- `footers` - Array of footer text (if requested)

### get_document_text

Get all text from a document as a single plain text string.

**Input:**

- `document_id` (required) - The ID or URL of the Google Doc

**Output:**

- `title` - Document title
- `text` - Full document text

### get_document_paragraphs

Get text organized by paragraphs.

**Input:**

- `document_id` (required) - The ID or URL of the Google Doc

**Output:**

- `title` - Document title
- `paragraphs` - Array of paragraph text strings

## Google Sheets Tools

### get_spreadsheet_metadata

Get metadata about a spreadsheet.

**Input:**

- `spreadsheet_id` (required) - The ID or URL of the Google Sheets spreadsheet

**Output:**

- `spreadsheet_id` - Spreadsheet ID
- `title` - Spreadsheet title
- `locale` - Spreadsheet locale
- `time_zone` - Spreadsheet time zone
- `sheet_count` - Number of sheets
- `url` - Spreadsheet URL

### list_sheets

List all sheets in a spreadsheet with their properties.

**Input:**

- `spreadsheet_id` (required) - The ID or URL of the Google Sheets spreadsheet

**Output:**

- `spreadsheet_id` - Spreadsheet ID
- `title` - Spreadsheet title
- `sheets` - Array of sheet information:
  - `index` - Sheet index
  - `sheet_id` - Sheet GID
  - `title` - Sheet title
  - `sheet_type` - Sheet type (GRID, etc.)
  - `hidden` - Whether sheet is hidden
  - `row_count` - Number of rows
  - `column_count` - Number of columns
  - `frozen_row_count` - Number of frozen rows (if any)
  - `frozen_column_count` - Number of frozen columns (if any)

### get_sheet_values

Get cell values from a specific range.

**Input:**

- `spreadsheet_id` (required) - The ID or URL of the Google Sheets spreadsheet
- `range` (required) - A1 notation range (e.g., 'Sheet1!A1:D10', 'A:D', 'A1:D10')
- `sheet_index` (optional) - Zero-based sheet index (used when range doesn't include sheet name)
- `sheet_name` (optional) - Sheet name (mutually exclusive with sheet_index)
- `value_format` (optional) - Output format: 'formatted' (default), 'typed', 'raw'

**Output:**

- `spreadsheet_id` - Spreadsheet ID
- `range` - Resolved range
- `values` - 2D array of cell values (format depends on value_format)

### get_sheet_data

Get all data from a specific sheet.

**Input:**

- `spreadsheet_id` (required) - The ID or URL of the Google Sheets spreadsheet
- `sheet_index` (optional) - Zero-based sheet index (default: 0)
- `sheet_name` (optional) - Sheet name (mutually exclusive with sheet_index and sheet_gid)
- `sheet_gid` (optional) - Sheet GID from URL (mutually exclusive with sheet_index and sheet_name)
- `value_format` (optional) - Output format: 'formatted' (default), 'typed', 'raw'
- `include_metadata` (optional) - Include sheet metadata (default: false)

**Output:**

- `spreadsheet_id` - Spreadsheet ID
- `sheet_name` - Sheet name
- `range` - Data range
- `values` - 2D array of cell values
- `metadata` - Sheet metadata (if requested)

### get_multiple_ranges

Batch get values from multiple ranges in a single request.

**Input:**

- `spreadsheet_id` (required) - The ID or URL of the Google Sheets spreadsheet
- `ranges` (required) - Array of A1 notation ranges
- `value_format` (optional) - Output format: 'formatted' (default), 'typed', 'raw'

**Output:**

- `spreadsheet_id` - Spreadsheet ID
- `ranges` - Array of range results, each containing:
  - `range` - Resolved range
  - `values` - 2D array of cell values

## Google Slides Tools

### get_presentation

Get metadata about a presentation.

**Input:**

- `presentation_id` (required) - The ID of the Google Slides presentation

**Output:**

- `title` - Presentation title
- `slide_count` - Number of slides
- `locale` - Presentation locale
- `revision_id` - Current revision ID

### list_slides

List all slides in a presentation.

**Input:**

- `presentation_id` (required) - The ID of the Google Slides presentation

**Output:**

- `slides` - Array of slide information:
  - `object_id` - Slide's unique identifier
  - `index` - Zero-based slide index
  - `title` - Slide title (if present)
  - `element_count` - Number of elements on the slide

### get_slide

Get the content of a specific slide.

**Input:**

- `presentation_id` (required) - The ID of the Google Slides presentation
- `slide_index` (optional) - Zero-based slide index
- `slide_object_id` (optional) - Slide's object ID

One of `slide_index` or `slide_object_id` must be provided.

**Output:**

- `text_content` - Array of text strings from the slide
- `element_summary` - Array of element details:
  - `object_id` - Element's unique identifier
  - `element_type` - Type of element (shape, image, table, etc.)
  - `description` - Element description or text preview

### get_slide_notes

Get the speaker notes for a specific slide.

**Input:**

- `presentation_id` (required) - The ID of the Google Slides presentation
- `slide_index` (optional) - Zero-based slide index
- `slide_object_id` (optional) - Slide's object ID

One of `slide_index` or `slide_object_id` must be provided.

**Output:**

- `notes` - Speaker notes text

### get_presentation_content

Get all slide content in a single call - ideal for AI analysis of the entire presentation.

**Input:**

- `presentation_id` (required) - The ID of the Google Slides presentation
- `include_notes` (optional) - Include speaker notes for each slide (default: false)

**Output:**

- `title` - Presentation title
- `slides` - Array of slide content:
  - `index` - Zero-based slide index
  - `object_id` - Slide's unique identifier
  - `title` - Slide title (if present)
  - `text_content` - Array of text strings from the slide
  - `images` - Array of images:
    - `object_id` - Image element ID
    - `content_url` - Direct URL to image (valid ~30 minutes)
    - `source_url` - Original source URL (if available)
    - `alt_text` - Image description
  - `notes` - Speaker notes (if `include_notes` is true)

## Finding Document IDs

### Documents

The document ID is in the URL when viewing a document:

```
https://docs.google.com/document/d/DOCUMENT_ID_HERE/edit
```

**Note:** Google Docs tools accept either the document ID or the full URL, including URLs with query strings and anchors:

```
https://docs.google.com/document/d/DOCUMENT_ID_HERE/edit?tab=t.0#heading=h.xyz
```

### Spreadsheets

The spreadsheet ID is in the URL when viewing a spreadsheet:

```
https://docs.google.com/spreadsheets/d/SPREADSHEET_ID_HERE/edit
```

**Note:** Google Sheets tools accept either the spreadsheet ID or the full URL. URLs may include sheet GID and range:

```
https://docs.google.com/spreadsheets/d/SPREADSHEET_ID_HERE/edit#gid=123&range=A1:D10
```

### Presentations

The presentation ID is in the URL when viewing a presentation:

```
https://docs.google.com/presentation/d/PRESENTATION_ID_HERE/edit
```

## License

MIT
