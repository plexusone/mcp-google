# Google Docs Tools

!!! note "URL Support"
    All Google Docs tools accept either the document ID or the full URL, including URLs with query strings and anchors like:
    ```
    https://docs.google.com/document/d/abc123/edit?tab=t.0#heading=h.xyz
    ```

## get_document_metadata

Get metadata about a Google Doc document.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `document_id` | string | Yes | The ID or URL of the Google Doc |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Document title |
| `document_id` | string | Document ID |
| `revision_id` | string | Current revision ID |
| `word_count` | integer | Approximate word count |
| `char_count` | integer | Character count |
| `image_count` | integer | Number of images |
| `table_count` | integer | Number of tables |
| `header_count` | integer | Number of headers |
| `footer_count` | integer | Number of footers |

### Example

```json
{
  "name": "get_document_metadata",
  "arguments": {
    "document_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
  }
}
```

---

## get_document_content

Get the full structured content of a document including headings, paragraphs, and optionally images and tables.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `document_id` | string | Yes | The ID or URL of the Google Doc |
| `include_images` | boolean | No | Include image information (default: false) |
| `include_tables` | boolean | No | Include table content (default: false) |
| `include_headers` | boolean | No | Include document headers (default: false) |
| `include_footers` | boolean | No | Include document footers (default: false) |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Document title |
| `sections` | array | Array of content sections |
| `sections[].type` | string | Section type ("heading", "paragraph") |
| `sections[].level` | integer | Heading level (1-6, for headings only) |
| `sections[].text` | string | Section text content |
| `sections[].style_id` | string | Style identifier (e.g., "HEADING_1") |
| `images` | array | Array of images (if requested) |
| `images[].object_id` | string | Image element ID |
| `images[].content_uri` | string | Direct URL to image |
| `images[].source_uri` | string | Original source URL |
| `images[].title` | string | Image title |
| `images[].description` | string | Image description |
| `tables` | array | Array of tables (if requested) |
| `tables[].rows` | integer | Number of rows |
| `tables[].columns` | integer | Number of columns |
| `tables[].cells` | array | 2D array of cell text content |
| `headers` | array | Array of header text (if requested) |
| `footers` | array | Array of footer text (if requested) |

### Example

```json
{
  "name": "get_document_content",
  "arguments": {
    "document_id": "https://docs.google.com/document/d/abc123/edit",
    "include_images": true,
    "include_tables": true
  }
}
```

---

## get_document_text

Get all text content from a document as a single plain text string.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `document_id` | string | Yes | The ID or URL of the Google Doc |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Document title |
| `text` | string | Full document text |

### Example

```json
{
  "name": "get_document_text",
  "arguments": {
    "document_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
  }
}
```

!!! tip "Best for Simple Text Extraction"
    Use this when you just need the raw text content without structure.

---

## get_document_paragraphs

Get text organized by paragraphs.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `document_id` | string | Yes | The ID or URL of the Google Doc |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Document title |
| `paragraphs` | array | Array of paragraph text strings |

### Example

```json
{
  "name": "get_document_paragraphs",
  "arguments": {
    "document_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
  }
}
```

!!! tip "Best for Paragraph-Level Processing"
    Use this when you need to process text paragraph by paragraph, such as for summarization or analysis tasks.
