package sheets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sheetsutil "github.com/grokify/gogoogle/sheetsutil/v4"
	"github.com/plexusone/omniskill/skill"
	"google.golang.org/api/option"
)

// newSkill returns a Sheets skill whose gogoogle service talks to a fake
// Sheets API. Requests are recorded; the body depends on the path.
func newSkill(t *testing.T) (*Skill, *[]*http.Request) {
	t.Helper()
	var reqs []*http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs = append(reqs, r.Clone(context.Background()))
		w.Header().Set("Content-Type", "application/json")
		body := `{"range":"Sheet1!A1:B2","values":[["a","b"],["1","2"]]}`
		switch {
		case r.URL.Path == "/v4/spreadsheets/ss1":
			body = `{"spreadsheetId":"ss1","properties":{"title":"Round 1","locale":"en_US"},"sheets":[{"properties":{"title":"Sheet1","sheetId":0}}]}`
		case r.URL.Path == "/v4/spreadsheets/ss1/values:batchGet":
			body = `{"valueRanges":[{"range":"A1:B1","values":[["a","b"]]},{"range":"C1","values":[["c"]]}]}`
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	svc, err := sheetsutil.NewService(context.Background(), srv.Client(), option.WithEndpoint(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	return NewWithService(svc), &reqs
}

func tool(t *testing.T, s *Skill, name string) skill.Tool {
	t.Helper()
	for _, tl := range s.Tools() {
		if tl.Name() == name {
			return tl
		}
	}
	t.Fatalf("no tool %q", name)
	return nil
}

func TestGetSpreadsheetMetadata(t *testing.T) {
	s, _ := newSkill(t)
	got, err := tool(t, s, "get_spreadsheet_metadata").Call(context.Background(), map[string]any{"spreadsheet_id": "ss1"})
	if err != nil {
		t.Fatal(err)
	}
	if m := got.(map[string]any); m["title"] != "Round 1" {
		t.Errorf("metadata = %v", m)
	}
}

func TestGetSheetValuesPassesTheRenderOption(t *testing.T) {
	for format, want := range map[string]string{"formatted": "FORMATTED_VALUE", "raw": "UNFORMATTED_VALUE", "typed": "UNFORMATTED_VALUE"} {
		s, reqs := newSkill(t)
		_, err := tool(t, s, "get_sheet_values").Call(context.Background(),
			map[string]any{"spreadsheet_id": "ss1", "range": "Sheet1!A1:B2", "value_format": format})
		if err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		last := (*reqs)[len(*reqs)-1]
		if got := last.URL.Query().Get("valueRenderOption"); got != want {
			t.Errorf("value_format %q sent valueRenderOption=%q, want %q", format, got, want)
		}
	}
}

func TestGetMultipleRanges(t *testing.T) {
	s, reqs := newSkill(t)
	got, err := tool(t, s, "get_multiple_ranges").Call(context.Background(),
		map[string]any{"spreadsheet_id": "ss1", "ranges": []any{"A1:B1", "C1"}})
	if err != nil {
		t.Fatal(err)
	}
	m := got.(map[string]any)
	if rs, ok := m["ranges"].([]map[string]any); !ok || len(rs) != 2 {
		t.Errorf("result = %v", m)
	}
	if q := (*reqs)[len(*reqs)-1].URL.Query()["ranges"]; len(q) != 2 {
		t.Errorf("ranges sent = %v", q)
	}
}
