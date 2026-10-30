package drive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	driveutil "github.com/grokify/gogoogle/driveutil/v3"
	"github.com/plexusone/omniskill/skill"
	"google.golang.org/api/option"
)

// newSkill returns a Drive skill whose gogoogle service talks to a fake
// Drive API that answers each request with the next body in bodies.
func newSkill(t *testing.T, bodies ...string) (*Skill, *[]*http.Request) {
	t.Helper()
	var reqs []*http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := len(reqs)
		reqs = append(reqs, r.Clone(context.Background()))
		if i >= len(bodies) {
			i = len(bodies) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(bodies[i])); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	svc, err := driveutil.NewService(context.Background(), srv.Client(), option.WithEndpoint(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	return NewWithService(svc), &reqs
}

func commentsTool(t *testing.T, s *Skill) skill.Tool {
	t.Helper()
	for _, tl := range s.Tools() {
		if tl.Name() == "get_file_comments" {
			return tl
		}
	}
	t.Fatal(`no tool "get_file_comments"`)
	return nil
}

const (
	page1 = `{"comments":[{"id":"c1","content":"Tighten this","quotedFileContent":{"value":"the plan"},
	  "resolved":false,"author":{"displayName":"Reviewer A"},"createdTime":"2026-01-02T03:04:05Z","modifiedTime":"2026-01-02T03:04:05Z",
	  "replies":[{"id":"r1","content":"Agreed","author":{"displayName":"Owner"},"createdTime":"2026-01-03T03:04:05Z","modifiedTime":"2026-01-03T03:04:05Z"}]}],
	  "nextPageToken":"p2"}`
	page2 = `{"comments":[{"id":"c2","content":"Looks right","quotedFileContent":{"value":"the goal"},
	  "resolved":true,"author":{"displayName":"Reviewer B"},"createdTime":"2026-01-04T03:04:05Z","modifiedTime":"2026-01-04T03:04:05Z"}]}`
)

func TestGetFileCommentsReturnsEveryPage(t *testing.T) {
	s, reqs := newSkill(t, page1, page2)
	got, err := commentsTool(t, s).Call(context.Background(),
		map[string]any{"file_id": "https://docs.google.com/document/d/file1/edit"})
	if err != nil {
		t.Fatal(err)
	}
	m := got.(map[string]any)
	if m["file_id"] != "file1" {
		t.Errorf("file_id = %v", m["file_id"])
	}
	comments := m["comments"].([]driveutil.Comment)
	if len(comments) != 2 {
		t.Fatalf("got %d comments, want both pages", len(comments))
	}
	if c := comments[0]; c.QuotedText != "the plan" || c.Resolved || len(c.Replies) != 1 || c.Replies[0].Content != "Agreed" || c.Author.DisplayName != "Reviewer A" {
		t.Errorf("first comment = %+v", c)
	}
	if c := comments[1]; c.QuotedText != "the goal" || !c.Resolved {
		t.Errorf("second comment = %+v", c)
	}
	if len(*reqs) != 2 || (*reqs)[0].URL.Path != "/files/file1/comments" {
		t.Errorf("want one request per page, got %d (%v)", len(*reqs), (*reqs)[0].URL.Path)
	}
	if (*reqs)[1].URL.Query().Get("pageToken") != "p2" {
		t.Errorf("second request did not follow the page token: %s", (*reqs)[1].URL.RawQuery)
	}
}

func TestGetFileCommentsEmptyIsAnEmptyList(t *testing.T) {
	s, _ := newSkill(t, `{"comments":[]}`)
	got, err := commentsTool(t, s).Call(context.Background(), map[string]any{"file_id": "file1"})
	if err != nil {
		t.Fatal(err)
	}
	if c := got.(map[string]any)["comments"].([]driveutil.Comment); c == nil || len(c) != 0 {
		t.Errorf("comments = %#v, want an empty non-nil list", c)
	}
}

func TestGetFileCommentsIncludeDeleted(t *testing.T) {
	s, reqs := newSkill(t, `{"comments":[]}`)
	if _, err := commentsTool(t, s).Call(context.Background(),
		map[string]any{"file_id": "file1", "include_deleted": true}); err != nil {
		t.Fatal(err)
	}
	if (*reqs)[0].URL.Query().Get("includeDeleted") != "true" {
		t.Errorf("query = %s", (*reqs)[0].URL.RawQuery)
	}
}

func TestFileID(t *testing.T) {
	for in, want := range map[string]string{
		"1AbC_d-9": "1AbC_d-9",
		"https://docs.google.com/document/d/DOC1/edit":      "DOC1",
		"https://docs.google.com/spreadsheets/d/SHEET1/e":   "SHEET1",
		"https://drive.google.com/file/d/FILE1/view":        "FILE1",
		"https://drive.google.com/open?id=OPEN1":            "OPEN1",
		"https://docs.google.com/presentation/d/DECK1/edit": "DECK1",
	} {
		if got, err := fileID(in); err != nil || got != want {
			t.Errorf("fileID(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "not a url", "https://example.org/x", "a/b", "file://d/x"} {
		if _, err := fileID(in); err == nil {
			t.Errorf("fileID(%q) was accepted", in)
		}
	}
}

func TestGetFileCommentsRequiresAFileID(t *testing.T) {
	s, _ := newSkill(t, `{"comments":[]}`)
	if _, err := commentsTool(t, s).Call(context.Background(), map[string]any{}); err == nil {
		t.Error("a call with no file_id was accepted")
	}
}
