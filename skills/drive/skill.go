// Package drive provides an omniskill Skill for Google Drive: reading the
// comments on a file, and (when explicitly enabled) uploading and updating
// files.
//
// All Google access goes through gogoogle's driveutil package.
package drive

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"

	driveutil "github.com/grokify/gogoogle/driveutil/v3"
	oscfs "github.com/grokify/oscompat/fs"
	"github.com/plexusone/omniskill/skill"
)

// Skill provides Google Drive tools.
type Skill struct {
	httpClient *http.Client
	svc        *driveutil.Service
	write      writeConfig
	logger     *slog.Logger
}

// New creates a Drive skill with the given authenticated HTTP client. By
// default it is read-only: it registers no tool that can change Google data.
func New(httpClient *http.Client, opts ...Option) *Skill {
	s := &Skill{httpClient: httpClient, logger: slog.Default()}
	for _, o := range opts {
		o(s)
	}
	return s
}

// NewWithService creates a Drive skill that uses an existing gogoogle
// service, for callers that build the service themselves (tests use a fake
// endpoint).
func NewWithService(svc *driveutil.Service, opts ...Option) *Skill {
	s := New(nil, opts...)
	s.svc = svc
	return s
}

// Name returns the skill identifier.
func (s *Skill) Name() string { return "drive" }

// Version returns the skill version.
func (s *Skill) Version() string { return "0.5.0" }

// Description returns what this skill does.
func (s *Skill) Description() string {
	if s.write.enabled {
		return "Read the comments on Google Drive files, and upload or update files under the configured upload root"
	}
	return "Read the comments on Google Drive files, including Docs, Sheets, and Slides"
}

// Init creates the gogoogle Drive service unless one was supplied, and in
// write mode resolves the upload root, which must be an existing directory.
func (s *Skill) Init(ctx context.Context) error {
	if s.write.enabled {
		root, err := oscfs.RealDir(s.write.root)
		if err != nil {
			return fmt.Errorf("upload root: %w", err)
		}
		s.write.realRoot = root
	}
	if s.svc != nil {
		return nil
	}
	svc, err := driveutil.NewService(ctx, s.httpClient)
	if err != nil {
		return fmt.Errorf("failed to create Drive service: %w", err)
	}
	s.svc = svc
	return nil
}

// Close releases resources (no-op for this skill).
func (s *Skill) Close() error { return nil }

// Tools returns all tools provided by this skill.
//
// The write tools, upload_file and update_file, are present only when the skill
// was built with WithWrite.
func (s *Skill) Tools() []skill.Tool {
	tools := []skill.Tool{
		s.getFileCommentsTool(),
	}
	if s.write.enabled {
		tools = append(tools, s.uploadFileTool(), s.updateFileTool())
	}
	return tools
}

// Ensure Skill implements skill.Skill.
var _ skill.Skill = (*Skill)(nil)

var (
	// A Drive, Docs, Sheets, or Slides URL carries the file ID after /d/.
	rxPathID = regexp.MustCompile(`/d/([A-Za-z0-9_-]+)`)
	// A bare file ID.
	rxID = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// fileID extracts a Drive file ID from an ID or from a Drive, Docs, Sheets, or
// Slides URL (/d/<id>/..., or an open?id=<id> link).
func fileID(input string) (string, error) {
	if rxID.MatchString(input) {
		return input, nil
	}
	u, err := url.Parse(input)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		if m := rxPathID.FindStringSubmatch(u.Path); m != nil {
			return m[1], nil
		}
		if id := u.Query().Get("id"); rxID.MatchString(id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("not a Drive file ID or URL: %q", input)
}

func (s *Skill) getFileCommentsTool() skill.Tool {
	return skill.NewTool(
		"get_file_comments",
		"List the comments on a Google Drive file, including Docs, Sheets, and Slides. Each comment has its author, the text it is anchored to (quoted_text), whether it is resolved, its replies, and timestamps. Suggested edits are not comments; use get_document_suggestions for those. Accepts a file ID or a Drive, Docs, Sheets, or Slides URL.",
		map[string]skill.Parameter{
			"file_id": {
				Type:        "string",
				Description: "The ID or URL of the Drive file",
				Required:    true,
			},
			"include_deleted": {
				Type:        "boolean",
				Description: "Include deleted comments",
				Required:    false,
				Default:     false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			idOrURL, _ := params["file_id"].(string)
			if idOrURL == "" {
				return nil, fmt.Errorf("file_id is required")
			}
			id, err := fileID(idOrURL)
			if err != nil {
				return nil, err
			}
			includeDeleted, _ := params["include_deleted"].(bool)

			comments, err := s.svc.ListComments(ctx, id, driveutil.ListCommentsOptions{IncludeDeleted: includeDeleted})
			if err != nil {
				return nil, fmt.Errorf("failed to list comments: %w", err)
			}
			if comments == nil {
				comments = []driveutil.Comment{}
			}
			// MCP structured results are objects, so the list is a field.
			return map[string]any{
				"file_id":  id,
				"comments": comments,
			}, nil
		},
	)
}
