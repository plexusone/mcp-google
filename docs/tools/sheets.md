# Google Sheets Tools

!!! note "URL Support"
    All Google Sheets tools accept either the spreadsheet ID or the full URL, including URLs with query strings and fragments like:
    ```
    https://docs.google.com/spreadsheets/d/abc123/edit#gid=0
    ```

## get_spreadsheet_metadata

Get metadata about a Google Sheets spreadsheet.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The ID or URL of the Google Sheets spreadsheet |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `spreadsheet_id` | string | Spreadsheet ID |
| `title` | string | Spreadsheet title |
| `locale` | string | Spreadsheet locale |
| `time_zone` | string | Spreadsheet time zone |
| `sheet_count` | integer | Number of sheets |
| `url` | string | Spreadsheet URL |

### Example

```json
{
  "name": "get_spreadsheet_metadata",
  "arguments": {
    "spreadsheet_id": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
  }
}
```

---

## list_sheets

List all sheets in a spreadsheet with their properties.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The ID or URL of the Google Sheets spreadsheet |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `spreadsheet_id` | string | Spreadsheet ID |
| `title` | string | Spreadsheet title |
| `sheets` | array | Array of sheet information |
| `sheets[].index` | integer | Zero-based sheet index |
| `sheets[].sheet_id` | integer | Sheet GID |
| `sheets[].title` | string | Sheet name |
| `sheets[].sheet_type` | string | Sheet type (GRID, CHART, etc.) |
| `sheets[].hidden` | boolean | Whether the sheet is hidden |
| `sheets[].row_count` | integer | Number of rows |
| `sheets[].column_count` | integer | Number of columns |
| `sheets[].frozen_row_count` | integer | Number of frozen rows (if any) |
| `sheets[].frozen_column_count` | integer | Number of frozen columns (if any) |

### Example

```json
{
  "name": "list_sheets",
  "arguments": {
    "spreadsheet_id": "https://docs.google.com/spreadsheets/d/abc123/edit"
  }
}
```

---

## get_sheet_values

Get cell values from a specific range.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The ID or URL of the Google Sheets spreadsheet |
| `range` | string | Yes | A1 notation range (e.g., 'Sheet1!A1:D10', 'A:D', 'A1:D10') |
| `sheet_index` | integer | No | Zero-based sheet index. Used when range doesn't include sheet name. |
| `sheet_name` | string | No | Sheet name. Used when range doesn't include sheet name. Mutually exclusive with sheet_index. |
| `value_format` | string | No | Output format: 'formatted' (default), 'typed', 'raw' |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `spreadsheet_id` | string | Spreadsheet ID |
| `range` | string | Resolved range with sheet name |
| `values` | array | 2D array of cell values (format depends on value_format) |

### Value Formats

**`formatted` (default):** Display values as strings

```json
{
  "values": [
    ["Name", "Date", "Amount"],
    ["Alice", "Jan 15, 2024", "$1,234.56"]
  ]
}
```

**`typed`:** Values with type information

```json
{
  "values": [
    [
      {"type": "string", "formatted_value": "Name"},
      {"type": "string", "formatted_value": "Date"},
      {"type": "string", "formatted_value": "Amount"}
    ],
    [
      {"type": "string", "formatted_value": "Alice"},
      {"type": "number", "formatted_value": "45306", "number_value": 45306},
      {"type": "number", "formatted_value": "1234.56", "number_value": 1234.56}
    ]
  ]
}
```

**`raw`:** Underlying values as strings

```json
{
  "values": [
    ["Name", "Date", "Amount"],
    ["Alice", "45306", "1234.56"]
  ]
}
```

### Example

```json
{
  "name": "get_sheet_values",
  "arguments": {
    "spreadsheet_id": "abc123",
    "range": "Sheet1!A1:D10",
    "value_format": "typed"
  }
}
```

---

## get_sheet_data

Get all data from a specific sheet.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The ID or URL of the Google Sheets spreadsheet |
| `sheet_index` | integer | No | Zero-based sheet index (default: 0) |
| `sheet_name` | string | No | Sheet name. Mutually exclusive with sheet_index and sheet_gid. |
| `sheet_gid` | integer | No | Sheet GID (from URL). Mutually exclusive with sheet_index and sheet_name. |
| `value_format` | string | No | Output format: 'formatted' (default), 'typed', 'raw' |
| `include_metadata` | boolean | No | Include sheet metadata (default: false) |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `spreadsheet_id` | string | Spreadsheet ID |
| `sheet_name` | string | Sheet name |
| `range` | string | Data range |
| `values` | array | 2D array of cell values |
| `metadata` | object | Sheet metadata (if requested) |
| `metadata.sheet_id` | integer | Sheet GID |
| `metadata.index` | integer | Zero-based sheet index |
| `metadata.row_count` | integer | Total row count |
| `metadata.column_count` | integer | Total column count |
| `metadata.frozen_row_count` | integer | Frozen rows (if any) |
| `metadata.frozen_column_count` | integer | Frozen columns (if any) |

### Example

```json
{
  "name": "get_sheet_data",
  "arguments": {
    "spreadsheet_id": "abc123",
    "sheet_name": "Data",
    "include_metadata": true
  }
}
```

!!! tip "Best for AI Analysis"
    Use this tool with `include_metadata: true` to get complete sheet data in one call, ideal for AI analysis of spreadsheet content.

---

## get_multiple_ranges

Get cell values from multiple ranges in a single request.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spreadsheet_id` | string | Yes | The ID or URL of the Google Sheets spreadsheet |
| `ranges` | array | Yes | Array of A1 notation ranges |
| `value_format` | string | No | Output format: 'formatted' (default), 'typed', 'raw' |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `spreadsheet_id` | string | Spreadsheet ID |
| `ranges` | array | Array of range results |
| `ranges[].range` | string | Resolved range |
| `ranges[].values` | array | 2D array of cell values |

### Example

```json
{
  "name": "get_multiple_ranges",
  "arguments": {
    "spreadsheet_id": "abc123",
    "ranges": ["Sheet1!A1:D10", "Sheet2!A1:B5", "Summary!A:A"],
    "value_format": "formatted"
  }
}
```

!!! tip "Efficient Batch Retrieval"
    Use this tool when you need data from multiple ranges to minimize API calls and improve performance.
