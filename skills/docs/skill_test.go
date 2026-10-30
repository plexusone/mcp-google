package docs

import (
	"context"
	"net/http"
	"net/http/httptest"
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
