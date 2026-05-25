# Google Slides Tools

## get_presentation

Get metadata about a Google Slides presentation.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `presentation_id` | string | Yes | The ID of the Google Slides presentation |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Presentation title |
| `slide_count` | integer | Number of slides |
| `locale` | string | Presentation locale |
| `revision_id` | string | Current revision ID |

### Example

```json
{
  "name": "get_presentation",
  "arguments": {
    "presentation_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
  }
}
```

---

## list_slides

List all slides in a presentation with their titles and element counts.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `presentation_id` | string | Yes | The ID of the Google Slides presentation |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `slides` | array | Array of slide information |
| `slides[].object_id` | string | Slide's unique identifier |
| `slides[].index` | integer | Zero-based slide index |
| `slides[].title` | string | Slide title (if present) |
| `slides[].element_count` | integer | Number of elements on the slide |

### Example

```json
{
  "name": "list_slides",
  "arguments": {
    "presentation_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
  }
}
```

---

## get_slide

Get the content and elements of a specific slide.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `presentation_id` | string | Yes | The ID of the Google Slides presentation |
| `slide_index` | integer | No* | Zero-based index of the slide |
| `slide_object_id` | string | No* | Object ID of the slide |

*One of `slide_index` or `slide_object_id` must be provided.

### Output

| Field | Type | Description |
|-------|------|-------------|
| `index` | integer | Slide index |
| `title` | string | Slide title |
| `text_content` | array | Array of text strings from the slide |
| `element_summary` | array | Array of element details |
| `element_summary[].object_id` | string | Element's unique identifier |
| `element_summary[].element_type` | string | Type (shape, image, table, etc.) |
| `element_summary[].description` | string | Element description or text preview |
| `images` | array | Array of images on the slide |
| `images[].object_id` | string | Image element ID |
| `images[].content_url` | string | Direct URL to image (~30 min validity) |
| `images[].source_url` | string | Original source URL |
| `images[].alt_text` | string | Image description |

### Example

```json
{
  "name": "get_slide",
  "arguments": {
    "presentation_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
    "slide_index": 0
  }
}
```

---

## get_slide_notes

Get the speaker notes for a specific slide.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `presentation_id` | string | Yes | The ID of the Google Slides presentation |
| `slide_index` | integer | No* | Zero-based index of the slide |
| `slide_object_id` | string | No* | Object ID of the slide |

*One of `slide_index` or `slide_object_id` must be provided.

### Output

| Field | Type | Description |
|-------|------|-------------|
| `notes` | string | Speaker notes text |

### Example

```json
{
  "name": "get_slide_notes",
  "arguments": {
    "presentation_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
    "slide_index": 0
  }
}
```

---

## get_presentation_content

Get all slide content in a single call - ideal for AI analysis of the entire presentation.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `presentation_id` | string | Yes | The ID of the Google Slides presentation |
| `include_notes` | boolean | No | Include speaker notes (default: false) |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Presentation title |
| `slides` | array | Array of slide content |
| `slides[].index` | integer | Zero-based slide index |
| `slides[].object_id` | string | Slide's unique identifier |
| `slides[].title` | string | Slide title (if present) |
| `slides[].text_content` | array | Array of text strings |
| `slides[].images` | array | Array of images |
| `slides[].notes` | string | Speaker notes (if `include_notes` is true) |

### Example

```json
{
  "name": "get_presentation_content",
  "arguments": {
    "presentation_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
    "include_notes": true
  }
}
```

!!! tip "Best for AI Analysis"
    This tool retrieves everything in one call, minimizing round-trips. Use it when you need to analyze the entire presentation.
