package docs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	docsutil "github.com/grokify/gogoogle/docsutil/v1"
	"github.com/plexusone/omniskill/skill"
	gdocs "google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// newSkill returns a Docs skill whose gogoogle service talks to a fake
// Docs API that answers every request with body.
func newSkill(t *testing.T, body string) (*Skill, *[]*http.Request) {
	t.Helper()
	var reqs []*http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs = append(reqs, r.Clone(context.Background()))
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	api, err := gdocs.NewService(context.Background(), option.WithHTTPClient(srv.Client()), option.WithEndpoint(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	svc := &docsutil.Service{DocsService: api, DocumentsService: gdocs.NewDocumentsService(api)}
	s := NewWithService(svc)
	if err := s.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s, &reqs
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

const plainDoc = `{"documentId":"d1","title":"Plan","body":{"content":[
  {"paragraph":{"elements":[{"textRun":{"content":"Hello world\n"}}]}}]}}`

func TestGetDocumentTextGoesThroughGogoogle(t *testing.T) {
	s, reqs := newSkill(t, plainDoc)
	got, err := tool(t, s, "get_document_text").Call(context.Background(),
		map[string]any{"document_id": "https://docs.google.com/document/d/d1/edit"})
	if err != nil {
		t.Fatal(err)
	}
	m := got.(map[string]any)
	if m["title"] != "Plan" || m["text"] != "Hello world\n" {
		t.Errorf("result = %v", m)
	}
	if len(*reqs) != 1 || (*reqs)[0].URL.Path != "/v1/documents/d1" {
		t.Errorf("requests = %v", *reqs)
	}
}

func TestGetDocumentMetadata(t *testing.T) {
	s, _ := newSkill(t, plainDoc)
	got, err := tool(t, s, "get_document_metadata").Call(context.Background(), map[string]any{"document_id": "d1"})
	if err != nil {
		t.Fatal(err)
	}
	if m := got.(map[string]any); m["title"] != "Plan" || m["word_count"] != 2 {
		t.Errorf("metadata = %v", m)
	}
}

func TestDocumentToolsRejectMissingID(t *testing.T) {
	s, _ := newSkill(t, plainDoc)
	for _, tl := range s.Tools() {
		if _, err := tl.Call(context.Background(), map[string]any{}); err == nil {
			t.Errorf("%s accepted a call with no document_id", tl.Name())
		}
	}
}

// suggestionDoc is a documents.get response with SUGGESTIONS_INLINE: a
// paragraph with a deletion and an insertion, and a table cell that is wholly
// inserted.
const suggestionDoc = `{"documentId":"d1","title":"Plan","body":{"content":[
  {"sectionBreak":{}},
  {"paragraph":{"elements":[
    {"textRun":{"content":"We will "}},
    {"textRun":{"content":"quickly ","suggestedDeletionIds":["s.del1"]}},
    {"textRun":{"content":"train "}},
    {"textRun":{"content":"all ","suggestedInsertionIds":["s.ins1"]}},
    {"textRun":{"content":"staff.\n"}}]}},
  {"table":{"tableRows":[{"tableCells":[
    {"content":[{"paragraph":{"elements":[{"textRun":{"content":"Monthly count\n","suggestedInsertionIds":["s.ins2"]}}]}}]}]}]}}
]}}`

func TestGetDocumentSuggestions(t *testing.T) {
	s, reqs := newSkill(t, suggestionDoc)
	got, err := tool(t, s, "get_document_suggestions").Call(context.Background(), map[string]any{"document_id": "d1"})
	if err != nil {
		t.Fatal(err)
	}
	sc, ok := got.(*docsutil.SuggestionContent)
	if !ok {
		t.Fatalf("result is %T, want the gogoogle SuggestionContent", got)
	}
	if sc.DocumentID != "d1" || len(sc.Blocks) != 2 {
		t.Fatalf("content = %+v", sc)
	}
	if p := sc.Blocks[0]; p.Original != "We will quickly train staff.\n" || p.Accepted != "We will train all staff.\n" {
		t.Errorf("paragraph original %q accepted %q", p.Original, p.Accepted)
	}
	if c := sc.Blocks[1]; c.Location != "body/table/2/r/0/c/0" || c.Original != "" || c.Accepted != "Monthly count\n" {
		t.Errorf("cell = %+v", c)
	}
	if got := strings.Join(sc.SuggestionIDs, ","); got != "s.del1,s.ins1,s.ins2" {
		t.Errorf("suggestion IDs = %s", got)
	}
	if q := (*reqs)[0].URL.Query().Get("suggestionsViewMode"); q != "SUGGESTIONS_INLINE" {
		t.Errorf("suggestionsViewMode = %q, want SUGGESTIONS_INLINE", q)
	}
}

// The tool output is the gogoogle JSON, unchanged, so clients see the same
// fields as the gogoogle command line.
func TestGetDocumentSuggestionsJSONShape(t *testing.T) {
	s, _ := newSkill(t, suggestionDoc)
	got, err := tool(t, s, "get_document_suggestions").Call(context.Background(), map[string]any{"document_id": "d1"})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"document_id", "blocks", "suggestion_ids"} {
		if _, ok := m[key]; !ok {
			t.Errorf("output lacks %q: %s", key, raw)
		}
	}
	block := m["blocks"].([]any)[0].(map[string]any)
	for _, key := range []string{"location", "segments", "original", "accepted"} {
		if _, ok := block[key]; !ok {
			t.Errorf("block lacks %q", key)
		}
	}
}
