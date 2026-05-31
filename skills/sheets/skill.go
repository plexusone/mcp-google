// Package sheets provides an omniskill Skill for reading Google Sheets spreadsheets.
//
// This package can be used standalone with mcp-google or composed
// with other skills in a multi-service MCP server.
package sheets

import (
	"context"
	"fmt"
	"net/http"

	sheetsutil "github.com/grokify/gogoogle/sheetsutil/v4"
	"github.com/plexusone/omniskill/skill"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Skill provides Google Sheets reading tools.
type Skill struct {
	httpClient   *http.Client
	spreadsheets *sheets.SpreadsheetsService
}

// New creates a new Sheets skill with the given authenticated HTTP client.
func New(httpClient *http.Client) *Skill {
	return &Skill{httpClient: httpClient}
}

// Name returns the skill identifier.
func (s *Skill) Name() string {
	return "sheets"
}

// Description returns what this skill does.
func (s *Skill) Description() string {
	return "Read Google Sheets spreadsheets including cell values, metadata, and sheet properties"
}

// Init initializes the Google Sheets API client.
func (s *Skill) Init(ctx context.Context) error {
	svc, err := sheets.NewService(ctx, option.WithHTTPClient(s.httpClient))
	if err != nil {
		return fmt.Errorf("failed to create Sheets service: %w", err)
	}
	s.spreadsheets = sheets.NewSpreadsheetsService(svc)
	return nil
}

// Close releases resources (no-op for this skill).
func (s *Skill) Close() error {
	return nil
}

// Tools returns all tools provided by this skill.
func (s *Skill) Tools() []skill.Tool {
	return []skill.Tool{
		s.getSpreadsheetMetadataTool(),
		s.listSheetsTool(),
		s.getSheetValuesTool(),
		s.getSheetDataTool(),
		s.getMultipleRangesTool(),
	}
}

// Ensure Skill implements skill.Skill.
var _ skill.Skill = (*Skill)(nil)

func (s *Skill) getSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	return s.spreadsheets.Get(spreadsheetID).Context(ctx).Do()
}

func (s *Skill) getSpreadsheetMetadataTool() skill.Tool {
	return skill.NewTool(
		"get_spreadsheet_metadata",
		"Get metadata about a Google Sheets spreadsheet including title, sheet count, locale, and time zone. Accepts spreadsheet ID or full URL.",
		map[string]skill.Parameter{
			"spreadsheet_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Sheets spreadsheet",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			spreadsheetIDOrURL, _ := params["spreadsheet_id"].(string)
			if spreadsheetIDOrURL == "" {
				return nil, fmt.Errorf("spreadsheet_id is required")
			}

			spreadsheetID, err := sheetsutil.ParseSpreadsheetURL(spreadsheetIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid spreadsheet ID or URL: %w", err)
			}

			ss, err := s.getSpreadsheet(ctx, spreadsheetID)
			if err != nil {
				return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
			}

			return map[string]any{
				"spreadsheet_id": ss.SpreadsheetId,
				"title":          ss.Properties.Title,
				"locale":         ss.Properties.Locale,
				"time_zone":      ss.Properties.TimeZone,
				"sheet_count":    len(ss.Sheets),
				"url":            ss.SpreadsheetUrl,
			}, nil
		},
	)
}

func (s *Skill) listSheetsTool() skill.Tool {
	return skill.NewTool(
		"list_sheets",
		"List all sheets in a Google Sheets spreadsheet with their properties. Accepts spreadsheet ID or full URL.",
		map[string]skill.Parameter{
			"spreadsheet_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Sheets spreadsheet",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			spreadsheetIDOrURL, _ := params["spreadsheet_id"].(string)
			if spreadsheetIDOrURL == "" {
				return nil, fmt.Errorf("spreadsheet_id is required")
			}

			spreadsheetID, err := sheetsutil.ParseSpreadsheetURL(spreadsheetIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid spreadsheet ID or URL: %w", err)
			}

			ss, err := s.getSpreadsheet(ctx, spreadsheetID)
			if err != nil {
				return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
			}

			sheetsInfo := make([]map[string]any, len(ss.Sheets))
			for i, sheet := range ss.Sheets {
				props := sheet.Properties
				sheetsInfo[i] = map[string]any{
					"index":        props.Index,
					"sheet_id":     props.SheetId,
					"title":        props.Title,
					"sheet_type":   props.SheetType,
					"hidden":       props.Hidden,
					"row_count":    props.GridProperties.RowCount,
					"column_count": props.GridProperties.ColumnCount,
				}
				if props.GridProperties.FrozenRowCount > 0 {
					sheetsInfo[i]["frozen_row_count"] = props.GridProperties.FrozenRowCount
				}
				if props.GridProperties.FrozenColumnCount > 0 {
					sheetsInfo[i]["frozen_column_count"] = props.GridProperties.FrozenColumnCount
				}
			}

			return map[string]any{
				"spreadsheet_id": ss.SpreadsheetId,
				"title":          ss.Properties.Title,
				"sheets":         sheetsInfo,
			}, nil
		},
	)
}

func (s *Skill) getSheetValuesTool() skill.Tool {
	return skill.NewTool(
		"get_sheet_values",
		"Get cell values from a specific range in a Google Sheets spreadsheet. Accepts spreadsheet ID or full URL.",
		map[string]skill.Parameter{
			"spreadsheet_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Sheets spreadsheet",
				Required:    true,
			},
			"range": {
				Type:        "string",
				Description: "A1 notation range (e.g., 'Sheet1!A1:D10', 'A:D', 'A1:D10'). Sheet name in range takes precedence over sheet_index/sheet_name params.",
				Required:    true,
			},
			"sheet_index": {
				Type:        "integer",
				Description: "Zero-based sheet index. Used when range doesn't include sheet name.",
				Required:    false,
			},
			"sheet_name": {
				Type:        "string",
				Description: "Sheet name. Used when range doesn't include sheet name. Mutually exclusive with sheet_index.",
				Required:    false,
			},
			"value_format": {
				Type:        "string",
				Description: "Output format: 'formatted' (default, display values), 'typed' (with type info), 'raw' (underlying values as strings)",
				Required:    false,
				Default:     "formatted",
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			spreadsheetIDOrURL, _ := params["spreadsheet_id"].(string)
			if spreadsheetIDOrURL == "" {
				return nil, fmt.Errorf("spreadsheet_id is required")
			}

			spreadsheetID, err := sheetsutil.ParseSpreadsheetURL(spreadsheetIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid spreadsheet ID or URL: %w", err)
			}

			rangeStr, _ := params["range"].(string)
			if rangeStr == "" {
				return nil, fmt.Errorf("range is required")
			}

			valueFormat, _ := params["value_format"].(string)
			if valueFormat == "" {
				valueFormat = "formatted"
			}

			// Build the full range with sheet name if needed
			fullRange := rangeStr
			if !containsSheetName(rangeStr) {
				sheetName, err := s.resolveSheetName(ctx, spreadsheetID, params)
				if err != nil {
					return nil, err
				}
				if sheetName != "" {
					fullRange = fmt.Sprintf("'%s'!%s", escapeSheetName(sheetName), rangeStr)
				}
			}

			// Determine value render option based on format
			valueRenderOption := "FORMATTED_VALUE"
			if valueFormat == "raw" || valueFormat == "typed" {
				valueRenderOption = "UNFORMATTED_VALUE"
			}

			vr, err := s.spreadsheets.Values.Get(spreadsheetID, fullRange).
				ValueRenderOption(valueRenderOption).
				Context(ctx).
				Do()
			if err != nil {
				return nil, fmt.Errorf("failed to get values: %w", err)
			}

			result := map[string]any{
				"spreadsheet_id": spreadsheetID,
				"range":          vr.Range,
			}

			switch valueFormat {
			case "typed":
				result["values"] = sheetsutil.ParseValueRange(vr)
			case "raw":
				result["values"] = sheetsutil.ExtractRawValues(vr)
			default: // "formatted"
				result["values"] = sheetsutil.ExtractFormattedValues(vr)
			}

			return result, nil
		},
	)
}

func (s *Skill) getSheetDataTool() skill.Tool {
	return skill.NewTool(
		"get_sheet_data",
		"Get all data from a specific sheet in a Google Sheets spreadsheet. Accepts spreadsheet ID or full URL.",
		map[string]skill.Parameter{
			"spreadsheet_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Sheets spreadsheet",
				Required:    true,
			},
			"sheet_index": {
				Type:        "integer",
				Description: "Zero-based sheet index (default: 0, first sheet)",
				Required:    false,
				Default:     0,
			},
			"sheet_name": {
				Type:        "string",
				Description: "Sheet name. Mutually exclusive with sheet_index and sheet_gid.",
				Required:    false,
			},
			"sheet_gid": {
				Type:        "integer",
				Description: "Sheet GID (from URL). Mutually exclusive with sheet_index and sheet_name.",
				Required:    false,
			},
			"value_format": {
				Type:        "string",
				Description: "Output format: 'formatted' (default), 'typed', 'raw'",
				Required:    false,
				Default:     "formatted",
			},
			"include_metadata": {
				Type:        "boolean",
				Description: "Include sheet metadata (row/column counts, frozen rows/cols)",
				Required:    false,
				Default:     false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			spreadsheetIDOrURL, _ := params["spreadsheet_id"].(string)
			if spreadsheetIDOrURL == "" {
				return nil, fmt.Errorf("spreadsheet_id is required")
			}

			spreadsheetID, err := sheetsutil.ParseSpreadsheetURL(spreadsheetIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid spreadsheet ID or URL: %w", err)
			}

			valueFormat, _ := params["value_format"].(string)
			if valueFormat == "" {
				valueFormat = "formatted"
			}

			includeMetadata, _ := params["include_metadata"].(bool)

			// Get spreadsheet to find sheet info
			ss, err := s.getSpreadsheet(ctx, spreadsheetID)
			if err != nil {
				return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
			}

			// Resolve sheet name from various parameters
			sheetName, sheetProps, err := s.resolveSheetWithProps(ss, params)
			if err != nil {
				return nil, err
			}

			// Build range for entire sheet
			fullRange := fmt.Sprintf("'%s'", escapeSheetName(sheetName))

			// Determine value render option based on format
			valueRenderOption := "FORMATTED_VALUE"
			if valueFormat == "raw" || valueFormat == "typed" {
				valueRenderOption = "UNFORMATTED_VALUE"
			}

			vr, err := s.spreadsheets.Values.Get(spreadsheetID, fullRange).
				ValueRenderOption(valueRenderOption).
				Context(ctx).
				Do()
			if err != nil {
				return nil, fmt.Errorf("failed to get values: %w", err)
			}

			result := map[string]any{
				"spreadsheet_id": spreadsheetID,
				"sheet_name":     sheetName,
				"range":          vr.Range,
			}

			switch valueFormat {
			case "typed":
				result["values"] = sheetsutil.ParseValueRange(vr)
			case "raw":
				result["values"] = sheetsutil.ExtractRawValues(vr)
			default: // "formatted"
				result["values"] = sheetsutil.ExtractFormattedValues(vr)
			}

			if includeMetadata && sheetProps != nil {
				metadata := map[string]any{
					"sheet_id":     sheetProps.SheetId,
					"index":        sheetProps.Index,
					"row_count":    sheetProps.GridProperties.RowCount,
					"column_count": sheetProps.GridProperties.ColumnCount,
				}
				if sheetProps.GridProperties.FrozenRowCount > 0 {
					metadata["frozen_row_count"] = sheetProps.GridProperties.FrozenRowCount
				}
				if sheetProps.GridProperties.FrozenColumnCount > 0 {
					metadata["frozen_column_count"] = sheetProps.GridProperties.FrozenColumnCount
				}
				result["metadata"] = metadata
			}

			return result, nil
		},
	)
}

func (s *Skill) getMultipleRangesTool() skill.Tool {
	return skill.NewTool(
		"get_multiple_ranges",
		"Get cell values from multiple ranges in a single request. Accepts spreadsheet ID or full URL.",
		map[string]skill.Parameter{
			"spreadsheet_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Sheets spreadsheet",
				Required:    true,
			},
			"ranges": {
				Type:        "array",
				Description: "Array of A1 notation ranges (e.g., ['Sheet1!A1:D10', 'Sheet2!A1:B5'])",
				Required:    true,
			},
			"value_format": {
				Type:        "string",
				Description: "Output format: 'formatted' (default), 'typed', 'raw'",
				Required:    false,
				Default:     "formatted",
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			spreadsheetIDOrURL, _ := params["spreadsheet_id"].(string)
			if spreadsheetIDOrURL == "" {
				return nil, fmt.Errorf("spreadsheet_id is required")
			}

			spreadsheetID, err := sheetsutil.ParseSpreadsheetURL(spreadsheetIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid spreadsheet ID or URL: %w", err)
			}

			rangesParam, _ := params["ranges"].([]any)
			if len(rangesParam) == 0 {
				return nil, fmt.Errorf("ranges is required and must not be empty")
			}

			ranges := make([]string, len(rangesParam))
			for i, r := range rangesParam {
				rstr, ok := r.(string)
				if !ok {
					return nil, fmt.Errorf("ranges[%d] must be a string", i)
				}
				ranges[i] = rstr
			}

			valueFormat, _ := params["value_format"].(string)
			if valueFormat == "" {
				valueFormat = "formatted"
			}

			// Determine value render option based on format
			valueRenderOption := "FORMATTED_VALUE"
			if valueFormat == "raw" || valueFormat == "typed" {
				valueRenderOption = "UNFORMATTED_VALUE"
			}

			resp, err := s.spreadsheets.Values.BatchGet(spreadsheetID).
				Ranges(ranges...).
				ValueRenderOption(valueRenderOption).
				Context(ctx).
				Do()
			if err != nil {
				return nil, fmt.Errorf("failed to get values: %w", err)
			}

			rangeResults := make([]map[string]any, len(resp.ValueRanges))
			for i, vr := range resp.ValueRanges {
				rangeResult := map[string]any{
					"range": vr.Range,
				}

				switch valueFormat {
				case "typed":
					rangeResult["values"] = sheetsutil.ParseValueRange(vr)
				case "raw":
					rangeResult["values"] = sheetsutil.ExtractRawValues(vr)
				default: // "formatted"
					rangeResult["values"] = sheetsutil.ExtractFormattedValues(vr)
				}

				rangeResults[i] = rangeResult
			}

			return map[string]any{
				"spreadsheet_id": resp.SpreadsheetId,
				"ranges":         rangeResults,
			}, nil
		},
	)
}

// Helper functions

// containsSheetName checks if a range string includes a sheet name (contains '!').
func containsSheetName(rangeStr string) bool {
	return len(rangeStr) > 0 && (rangeStr[0] == '\'' || hasExclamation(rangeStr))
}

func hasExclamation(s string) bool {
	for _, c := range s {
		if c == '!' {
			return true
		}
	}
	return false
}

// escapeSheetName escapes single quotes in sheet names.
func escapeSheetName(name string) string {
	// Sheet names containing single quotes need them doubled
	result := ""
	for _, c := range name {
		if c == '\'' {
			result += "''"
		} else {
			result += string(c)
		}
	}
	return result
}

// resolveSheetName resolves the sheet name from parameters.
func (s *Skill) resolveSheetName(ctx context.Context, spreadsheetID string, params map[string]any) (string, error) {
	// Check for sheet_name parameter
	if sheetName, ok := params["sheet_name"].(string); ok && sheetName != "" {
		return sheetName, nil
	}

	// Check for sheet_index parameter
	var sheetIndex *int
	if idx, ok := params["sheet_index"]; ok && idx != nil {
		switch v := idx.(type) {
		case float64:
			i := int(v)
			sheetIndex = &i
		case int:
			sheetIndex = &v
		}
	}

	if sheetIndex == nil {
		return "", nil // No sheet specified, use default
	}

	// Get spreadsheet to find sheet name by index
	ss, err := s.getSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return "", fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	if *sheetIndex < 0 || *sheetIndex >= len(ss.Sheets) {
		return "", fmt.Errorf("sheet_index %d out of range (0-%d)", *sheetIndex, len(ss.Sheets)-1)
	}

	return ss.Sheets[*sheetIndex].Properties.Title, nil
}

// resolveSheetWithProps resolves sheet name and properties from parameters.
func (s *Skill) resolveSheetWithProps(ss *sheets.Spreadsheet, params map[string]any) (string, *sheets.SheetProperties, error) {
	// Check for sheet_name parameter
	if sheetName, ok := params["sheet_name"].(string); ok && sheetName != "" {
		for _, sheet := range ss.Sheets {
			if sheet.Properties.Title == sheetName {
				return sheetName, sheet.Properties, nil
			}
		}
		return "", nil, fmt.Errorf("sheet not found: %s", sheetName)
	}

	// Check for sheet_gid parameter
	var sheetGID *int64
	if gid, ok := params["sheet_gid"]; ok && gid != nil {
		switch v := gid.(type) {
		case float64:
			i := int64(v)
			sheetGID = &i
		case int64:
			sheetGID = &v
		case int:
			i := int64(v)
			sheetGID = &i
		}
	}

	if sheetGID != nil {
		for _, sheet := range ss.Sheets {
			if sheet.Properties.SheetId == *sheetGID {
				return sheet.Properties.Title, sheet.Properties, nil
			}
		}
		return "", nil, fmt.Errorf("sheet not found with gid: %d", *sheetGID)
	}

	// Check for sheet_index parameter (default: 0)
	sheetIndex := 0
	if idx, ok := params["sheet_index"]; ok && idx != nil {
		switch v := idx.(type) {
		case float64:
			sheetIndex = int(v)
		case int:
			sheetIndex = v
		}
	}

	if sheetIndex < 0 || sheetIndex >= len(ss.Sheets) {
		return "", nil, fmt.Errorf("sheet_index %d out of range (0-%d)", sheetIndex, len(ss.Sheets)-1)
	}

	sheet := ss.Sheets[sheetIndex]
	return sheet.Properties.Title, sheet.Properties, nil
}
