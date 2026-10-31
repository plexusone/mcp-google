package slides

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grokify/gogoogle/slidesutil/v1"
	"github.com/plexusone/omniskill/skill"
	"google.golang.org/api/option"
	gslides "google.golang.org/api/slides/v1"
)

const deck = `{"presentationId":"p1","title":"Deck","locale":"en","revisionId":"r1",
  "slides":[{"objectId":"s1"},{"objectId":"s2"}]}`

func newSkill(t *testing.T) (*Skill, *[]string) {
	t.Helper()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(deck)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	api, err := gslides.NewService(context.Background(), option.WithHTTPClient(srv.Client()), option.WithEndpoint(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	svc := &slidesutil.GoogleSlidesService{SlidesService: api, PresentationsService: gslides.NewPresentationsService(api)}
	return NewWithService(svc), &paths
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

func TestGetPresentation(t *testing.T) {
	s, paths := newSkill(t)
	got, err := tool(t, s, "get_presentation").Call(context.Background(), map[string]any{"presentation_id": "p1"})
	if err != nil {
		t.Fatal(err)
	}
	m := got.(map[string]any)
	if m["title"] != "Deck" || m["slide_count"] != 2 {
		t.Errorf("result = %v", m)
	}
	if len(*paths) != 1 || (*paths)[0] != "/v1/presentations/p1" {
		t.Errorf("paths = %v", *paths)
	}
}

func TestListSlides(t *testing.T) {
	s, _ := newSkill(t)
	got, err := tool(t, s, "list_slides").Call(context.Background(), map[string]any{"presentation_id": "p1"})
	if err != nil {
		t.Fatal(err)
	}
	slides, ok := got.(map[string]any)["slides"].([]map[string]any)
	if !ok || len(slides) != 2 {
		t.Fatalf("slides = %v, want the two slides of the deck", got)
	}
	for i, want := range []string{"s1", "s2"} {
		if slides[i]["object_id"] != want || slides[i]["index"] != i {
			t.Errorf("slide %d = %v, want object_id %s at index %d", i, slides[i], want, i)
		}
	}
}
