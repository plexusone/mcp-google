package drive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	driveutil "github.com/grokify/gogoogle/driveutil/v3"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plexusone/omniskill/mcp/oauth2"
	runtime "github.com/plexusone/omniskill/mcp/server"
	"github.com/plexusone/omniskill/skill"
	"google.golang.org/api/option"
)

type call struct {
	method, path string
	body         string
}

// fakeDrive is a Drive API that records calls and answers with status/body.
type fakeDrive struct {
	status int
	body   string
	calls  []call
}

func (f *fakeDrive) service(t *testing.T) *driveutil.Service {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
		}
		f.calls = append(f.calls, call{r.Method, r.URL.Path, string(b)})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		if _, err := w.Write([]byte(f.body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	svc, err := driveutil.NewService(context.Background(), srv.Client(), option.WithEndpoint(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

const created = `{"id":"new1","name":"review","mimeType":"application/vnd.google-apps.document","webViewLink":"https://example.org/d/new1","modifiedTime":"2026-01-02T03:04:05Z"}`

// writeSkill returns a write-mode skill over a fake Drive, with an upload
// root holding review.docx, and the audit log.
func writeSkill(t *testing.T, f *fakeDrive, opts ...Option) (s *Skill, root string, audit *bytes.Buffer) {
	t.Helper()
	root = t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "review.docx"), []byte("PAYLOAD"), 0o600); err != nil {
		t.Fatal(err)
	}
	audit = &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(audit, nil))
	all := append([]Option{WithWrite(root), WithLogger(logger)}, opts...)
	s = NewWithService(f.service(t), all...)
	if err := s.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s, root, audit
}

func writeTool(t *testing.T, s *Skill, name string) skill.Tool {
	t.Helper()
	for _, tl := range s.Tools() {
		if tl.Name() == name {
			return tl
		}
	}
	t.Fatalf("no tool %q", name)
	return nil
}

func toolNames(s *Skill) []string {
	var names []string
	for _, tl := range s.Tools() {
		names = append(names, tl.Name())
	}
	slices.Sort(names)
	return names
}

func errCode(t *testing.T, err error) string {
	t.Helper()
	var te *ToolError
	if !errors.As(err, &te) {
		t.Fatalf("error %v is not a ToolError", err)
	}
	var parsed ToolError
	if jerr := json.Unmarshal([]byte(err.Error()), &parsed); jerr != nil || parsed.Code != te.Code || parsed.Message == "" {
		t.Errorf("error text %q is not {code,message} JSON: %v", err.Error(), jerr)
	}
	return te.Code
}

// INV-1: with write mode off, no tool can change Google data.
func TestDefaultModeRegistersNoWriteTool(t *testing.T) {
	if got := toolNames(New(nil)); !slices.Equal(got, []string{"get_file_comments"}) {
		t.Errorf("default tools = %v", got)
	}
	want := []string{"get_file_comments", "update_file", "upload_file"}
	if got := toolNames(New(nil, WithWrite("."))); !slices.Equal(got, want) {
		t.Errorf("write-mode tools = %v, want %v", got, want)
	}
}

func TestUploadConvertsAndReturnsTheFile(t *testing.T) {
	f := &fakeDrive{status: 200, body: created}
	s, _, audit := writeSkill(t, f)
	got, err := writeTool(t, s, "upload_file").Call(context.Background(),
		map[string]any{"path": "review.docx", "convert_to": "doc", "folder_id": "folder9"})
	if err != nil {
		t.Fatal(err)
	}
	file := got.(*driveutil.File)
	if file.ID != "new1" || file.WebViewLink == "" {
		t.Errorf("file = %+v", file)
	}
	if len(f.calls) != 1 || f.calls[0].method != "POST" || f.calls[0].path != "/upload/drive/v3/files" {
		t.Fatalf("calls = %+v", f.calls)
	}
	for _, want := range []string{driveutil.MimeGoogleDoc, "folder9", "PAYLOAD"} {
		if !strings.Contains(f.calls[0].body, want) {
			t.Errorf("request body lacks %q", want)
		}
	}
	// Audit: tool, file name, and the resulting file ID.
	var rec map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(audit.Bytes()), &rec); err != nil {
		t.Fatalf("audit log %q: %v", audit.String(), err)
	}
	if rec["msg"] != "write_call" || rec["tool"] != "upload_file" || rec["name"] != "review.docx" || rec["file_id"] != "new1" || rec["sub"] != "local" {
		t.Errorf("audit record = %v", rec)
	}
}

func TestUpdateKeepsTheFileID(t *testing.T) {
	f := &fakeDrive{status: 200, body: created}
	s, _, _ := writeSkill(t, f)
	_, err := writeTool(t, s, "update_file").Call(context.Background(),
		map[string]any{"file_id": "https://docs.google.com/document/d/f1/edit", "path": "review.docx", "convert_to": "doc"})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 || f.calls[0].method != "PATCH" || f.calls[0].path != "/upload/drive/v3/files/f1" {
		t.Errorf("calls = %+v", f.calls)
	}
}

func TestAuditRecordsTheCallerAndFailures(t *testing.T) {
	f := &fakeDrive{status: 200, body: created}
	s, _, audit := writeSkill(t, f)
	ctx := oauth2.SetTokenInfoContext(context.Background(), &oauth2.TokenInfo{Subject: "user:alice", ClientID: "client-1", Scope: "drive:write"})
	if _, err := writeTool(t, s, "upload_file").Call(ctx, map[string]any{"path": "../secret", "name": "x"}); err == nil {
		t.Fatal("a path outside the root was accepted")
	}
	out := audit.String()
	for _, want := range []string{`"msg":"write_call_failed"`, `"sub":"user:alice"`, `"client_id":"client-1"`, `"code":"path_outside_root"`, `"tool":"upload_file"`} {
		if !strings.Contains(out, want) {
			t.Errorf("audit log lacks %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, "PAYLOAD") || strings.Contains(out, "secret") && strings.Contains(out, os.TempDir()) {
		t.Errorf("audit log leaks content or server paths:\n%s", out)
	}
}

// Every refused request is refused before anything is sent to Google.
func TestRefusalsNeverReachDrive(t *testing.T) {
	f := &fakeDrive{status: 200, body: created}
	s, root, _ := writeSkill(t, f, WithMaxUploadBytes(5))
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.docx"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o750); err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		params map[string]any
		code   string
	}{
		"parent directory":    {map[string]any{"path": "../x.docx"}, CodePathOutsideRoot},
		"deep parent":         {map[string]any{"path": "a/../../x.docx"}, CodePathOutsideRoot},
		"absolute outside":    {map[string]any{"path": filepath.Join(outside, "secret.docx")}, CodePathOutsideRoot},
		"rooted path":         {map[string]any{"path": "/etc/passwd"}, CodePathOutsideRoot},
		"drive-relative path": {map[string]any{"path": "C:x.docx"}, CodePathOutsideRoot},
		"no such file":        {map[string]any{"path": "missing.docx"}, CodeNotFound},
		"a directory":         {map[string]any{"path": "dir"}, CodeInvalidRequest},
		"too large":           {map[string]any{"path": "review.docx"}, CodeFileTooLarge}, // 7 bytes > the limit of 5
		"missing path":        {map[string]any{}, CodeInvalidRequest},
		"unsupported convert": {map[string]any{"path": "review.docx", "convert_to": "pdf"}, CodeUnsupportedConversion},
	}
	for name, c := range cases {
		_, err := writeTool(t, s, "upload_file").Call(context.Background(), c.params)
		if err == nil {
			t.Errorf("%s: accepted", name)
			continue
		}
		if got := errCode(t, err); got != c.code {
			t.Errorf("%s: code %q, want %q (%v)", name, got, c.code, err)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("refused requests reached Drive: %+v", f.calls)
	}
}

func TestSymlinkOutOfTheRootIsRefused(t *testing.T) {
	f := &fakeDrive{status: 200, body: created}
	s, root, _ := writeSkill(t, f)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.docx"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("cannot create symbolic links here: %v", err)
	}
	for _, p := range []string{"link/secret.docx", "link"} {
		_, err := writeTool(t, s, "upload_file").Call(context.Background(), map[string]any{"path": p})
		if err == nil || errCode(t, err) != CodePathOutsideRoot {
			t.Errorf("%s: err = %v, want path_outside_root", p, err)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("a link out of the root reached Drive: %+v", f.calls)
	}
}

func TestDriveErrorsGetStableCodes(t *testing.T) {
	for status, code := range map[int]string{404: CodeNotFound, 403: CodePermissionDenied, 500: CodeAPIError} {
		f := &fakeDrive{status: status, body: `{"error":{"code":` + http.StatusText(status) + `}}`}
		f.body = `{"error":{"code":` + itoa(status) + `,"message":"nope"}}`
		s, _, _ := writeSkill(t, f)
		_, err := writeTool(t, s, "update_file").Call(context.Background(), map[string]any{"file_id": "f1", "path": "review.docx"})
		if err == nil {
			t.Fatalf("status %d: no error", status)
		}
		if got := errCode(t, err); got != code {
			t.Errorf("status %d: code %q, want %q", status, got, code)
		}
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func TestInitRejectsAMissingUploadRoot(t *testing.T) {
	s := NewWithService(nil, WithWrite(filepath.Join(t.TempDir(), "missing")))
	if err := s.Init(context.Background()); err == nil {
		t.Error("a missing upload root was accepted")
	}
}

// End to end over MCP: the tools are listed only in write mode, a refusal
// comes back as isError with the {code,message} JSON, and in default mode the
// write tools do not exist.
func TestOverMCP(t *testing.T) {
	ctx := context.Background()
	list := func(rt *runtime.Runtime) []string {
		var names []string
		for _, tl := range rt.ListTools() {
			names = append(names, tl.Name)
		}
		slices.Sort(names)
		return names
	}
	impl := &mcp.Implementation{Name: "test", Version: "0"}

	off := runtime.New(impl, nil)
	off.RegisterSkill(New(nil))
	if got := list(off); slices.Contains(got, "upload_file") || slices.Contains(got, "update_file") {
		t.Fatalf("default mode lists write tools: %v", got)
	}
	_, cs, err := off.InMemorySession(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "upload_file", Arguments: map[string]any{"path": "x.docx"}}); err == nil && !res.IsError {
		t.Error("default mode served upload_file")
	}

	f := &fakeDrive{status: 200, body: created}
	s, _, _ := writeSkill(t, f)
	on := runtime.New(impl, nil)
	on.RegisterSkill(s)
	if got := list(on); !slices.Contains(got, "upload_file") || !slices.Contains(got, "update_file") {
		t.Fatalf("write mode lacks the write tools: %v", got)
	}
	_, cs2, err := on.InMemorySession(ctx)
	if err != nil {
		t.Fatal(err)
	}
	res, err := cs2.CallTool(ctx, &mcp.CallToolParams{Name: "upload_file", Arguments: map[string]any{"path": "../x.docx"}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError || len(res.Content) == 0 {
		t.Fatalf("a refused path was not an error result: %+v", res)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	var te ToolError
	if err := json.Unmarshal([]byte(text), &te); err != nil || te.Code != CodePathOutsideRoot {
		t.Errorf("error content %q, want path_outside_root JSON (%v)", text, err)
	}
}
