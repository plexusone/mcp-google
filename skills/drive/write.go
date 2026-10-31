package drive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	driveutil "github.com/grokify/gogoogle/driveutil/v3"
	oscfs "github.com/grokify/oscompat/fs"
	"github.com/plexusone/omniskill/mcp/oauth2"
	"github.com/plexusone/omniskill/skill"
)

// DefaultMaxUploadBytes is the largest file the write tools will send.
const DefaultMaxUploadBytes int64 = 50 << 20

// Error codes returned by the write tools.
const (
	CodeNotFound              = "not_found"
	CodePermissionDenied      = "permission_denied"
	CodeUnsupportedConversion = "unsupported_conversion"
	CodePathOutsideRoot       = "path_outside_root"
	CodeWriteDisabled         = "write_disabled"
	CodeFileTooLarge          = "file_too_large"
	CodeInvalidRequest        = "invalid_request"
	CodeAPIError              = "api_error"
)

// ToolError is a write-tool failure with a stable code. Its text is the JSON
// {"code": ..., "message": ...}, which the MCP server returns to the client
// with isError set.
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ToolError) Error() string {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return string(b)
}

// WriteDisabledError is the error for a write request when write mode is off.
func WriteDisabledError() error {
	return toolErr(CodeWriteDisabled, "write tools are disabled; start with --enable-write (or MCP_GOOGLE_ENABLE_WRITE=true)")
}

func toolErr(code, format string, args ...any) *ToolError {
	return &ToolError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// writeConfig is the opt-in write mode.
type writeConfig struct {
	enabled  bool
	root     string // as configured
	realRoot string // resolved in Init
	maxBytes int64
	maxSet   bool // WithMaxUploadBytes ran, so the default does not apply
}

// Option configures a Drive skill.
type Option func(*Skill)

// WithWrite turns on the write tools, upload_file and update_file. Local
// paths they take are confined to root, which must be an existing directory.
func WithWrite(root string) Option {
	return func(s *Skill) {
		s.write.enabled = true
		s.write.root = root
		if !s.write.maxSet {
			s.write.maxBytes = DefaultMaxUploadBytes
		}
	}
}

// WithMaxUploadBytes sets the largest file the write tools will send.
func WithMaxUploadBytes(n int64) Option {
	return func(s *Skill) {
		s.write.maxBytes = n
		s.write.maxSet = true
	}
}

// WithLogger sets the logger that receives the audit record of each write call.
func WithLogger(l *slog.Logger) Option {
	return func(s *Skill) {
		if l != nil {
			s.logger = l
		}
	}
}

var convertEnum = []any{"doc", "sheet", "slides"}

func (s *Skill) uploadFileTool() skill.Tool {
	return skill.NewTool(
		"upload_file",
		"Upload a local file to Google Drive, optionally converting it to a Google Doc, Sheet, or Slides file. The path is relative to the server's upload root; paths that leave it are refused. Returns the new file's ID, name, and link. Needs the drive:write scope.",
		map[string]skill.Parameter{
			"path":      {Type: "string", Description: "Path of the local file, relative to the upload root", Required: true},
			"name":      {Type: "string", Description: "Name for the Drive file (default: the file name, without its extension when converting)"},
			"folder_id": {Type: "string", Description: "ID of the Drive folder to put the file in"},
			"convert_to": {
				Type:        "string",
				Description: "Convert to a Google file: doc (.docx, .odt, .txt, .html), sheet (.csv, .xlsx, .ods), or slides (.pptx, .odp). Omit to store the file unchanged",
				Enum:        convertEnum,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			p, err := stringParams(params, "path", "name", "folder_id", "convert_to")
			if err != nil {
				s.audit(ctx, "upload_file", "", "", nil, err)
				return nil, err
			}
			label := p["name"]
			if label == "" {
				label = filepath.Base(p["path"])
			}
			f, err := s.upload(ctx, p["path"], p["name"], p["folder_id"], "", p["convert_to"])
			s.audit(ctx, "upload_file", label, "", f, err)
			if err != nil {
				return nil, err
			}
			return f, nil
		},
	)
}

func (s *Skill) updateFileTool() skill.Tool {
	return skill.NewTool(
		"update_file",
		"Replace the content of an existing Google Drive file with a local file, keeping its ID, link, sharing, and comments. The path is relative to the server's upload root; paths that leave it are refused. Needs the drive:write scope.",
		map[string]skill.Parameter{
			"file_id": {Type: "string", Description: "The ID or URL of the Drive file to update", Required: true},
			"path":    {Type: "string", Description: "Path of the local file, relative to the upload root", Required: true},
			"convert_to": {
				Type:        "string",
				Description: "Convert to a Google file: doc, sheet, or slides. Omit to store the file unchanged",
				Enum:        convertEnum,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			p, err := stringParams(params, "file_id", "path", "convert_to")
			if err != nil {
				s.audit(ctx, "update_file", "", "", nil, err)
				return nil, err
			}
			id, err := fileID(p["file_id"])
			if err != nil {
				e := toolErr(CodeInvalidRequest, "file_id: %v", err)
				s.audit(ctx, "update_file", filepath.Base(p["path"]), "", nil, e)
				return nil, e
			}
			f, err := s.upload(ctx, p["path"], "", "", id, p["convert_to"])
			s.audit(ctx, "update_file", filepath.Base(p["path"]), id, f, err)
			if err != nil {
				return nil, err
			}
			return f, nil
		},
	)
}

// stringParams reads optional string parameters. A missing or null value is
// "". Any other type is an invalid_request, since dropping it would change what
// the call does: a numeric folder_id would otherwise put the file in the Drive
// root.
func stringParams(params map[string]any, keys ...string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		switch v := params[k].(type) {
		case nil:
			out[k] = ""
		case string:
			out[k] = v
		default:
			return nil, toolErr(CodeInvalidRequest, "%s: want a string, got %T", k, v)
		}
	}
	return out, nil
}

// upload opens the local file once, inside the upload root, checks the handle
// it holds, and streams that handle to gogoogle with the read capped at the size
// limit: an upload when updateID is empty, an update otherwise. The path is
// never opened a second time, so a link swapped in after the check cannot
// redirect the read, and a file that grows is cut off at the limit.
func (s *Skill) upload(ctx context.Context, path, name, folder, updateID, convert string) (*driveutil.File, error) {
	if !s.write.enabled {
		return nil, toolErr(CodeWriteDisabled, "write tools are disabled; start the server with --enable-write")
	}
	to, err := parseConvert(convert)
	if err != nil {
		return nil, err
	}
	file, err := s.open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			s.logger.Warn("close upload file", "error", cerr)
		}
	}()
	content := oscfs.CapReader(file, s.write.maxBytes)

	var f *driveutil.File
	if updateID == "" {
		f, err = s.svc.Upload(ctx, driveutil.UploadRequest{Path: path, Content: content, Name: name, FolderID: folder, ConvertTo: to})
	} else {
		f, err = s.svc.UpdateContent(ctx, updateID, path, content, to)
	}
	if err != nil {
		return nil, driveError(err)
	}
	return f, nil
}

func parseConvert(s string) (driveutil.ConvertTo, error) {
	switch c := driveutil.ConvertTo(s); c {
	case driveutil.ConvertNone, driveutil.ConvertDoc, driveutil.ConvertSheet, driveutil.ConvertSlides:
		return c, nil
	}
	return "", toolErr(CodeUnsupportedConversion, "convert_to %q: want doc, sheet, or slides", s)
}

// open opens path inside the upload root and returns the handle only if it is
// a regular file within the size limit. The open is confined by the operating
// system, so a link that leads out of the root is refused at the moment of the
// open, and the checks are made on the handle, not on the path.
func (s *Skill) open(path string) (*os.File, error) {
	if path == "" {
		return nil, toolErr(CodeInvalidRequest, "path is required")
	}
	root := s.write.realRoot
	if root == "" { // Init has not run, so resolve now
		r, err := oscfs.RealDir(s.write.root)
		if err != nil {
			return nil, toolErr(CodeInvalidRequest, "upload root: %v", err)
		}
		root = r
	}
	f, err := s.openFile(root, path, s.write.maxBytes)
	switch {
	case err == nil:
		return f, nil
	case errors.Is(err, oscfs.ErrOutsideRoot):
		return nil, toolErr(CodePathOutsideRoot, "%q is outside the upload root", path)
	case errors.Is(err, fs.ErrNotExist):
		return nil, toolErr(CodeNotFound, "no file %q under the upload root", path)
	case errors.Is(err, oscfs.ErrNotRegular):
		return nil, toolErr(CodeInvalidRequest, "%q is not a regular file", path)
	case errors.Is(err, oscfs.ErrTooLarge):
		return nil, toolErr(CodeFileTooLarge, "%q is larger than the limit of %d bytes", path, s.write.maxBytes)
	default:
		return nil, toolErr(CodeInvalidRequest, "path %q: cannot be read", path)
	}
}

// driveError maps a gogoogle error to a ToolError.
func driveError(err error) *ToolError {
	switch {
	case errors.Is(err, oscfs.ErrTooLarge):
		// The file passed the size check and then grew while it was sent.
		return toolErr(CodeFileTooLarge, "the file grew past the size limit while it was being sent")
	case errors.Is(err, driveutil.ErrNotFound):
		return toolErr(CodeNotFound, "Drive file not found, or not shared with this server's credentials")
	case errors.Is(err, driveutil.ErrPermissionDenied):
		return toolErr(CodePermissionDenied, "Drive denied the request; check the credentials' access and that the drive.file scope was granted")
	case errors.Is(err, driveutil.ErrUnsupportedConversion):
		return toolErr(CodeUnsupportedConversion, "%v", err)
	default:
		return toolErr(CodeAPIError, "Drive request failed: %v", err)
	}
}

// audit records a write call: who made it, which tool, the file name, and the
// resulting file ID or the failure code. Credentials and paths on the server's
// disk are not logged.
func (s *Skill) audit(ctx context.Context, tool, name, target string, f *driveutil.File, err error) {
	attrs := []any{"tool", tool, "name", name}
	if target != "" {
		attrs = append(attrs, "target", target)
	}
	if info := oauth2.GetTokenInfoContext(ctx); info != nil {
		attrs = append(attrs, "sub", info.Subject, "client_id", info.ClientID, "scope", info.Scope)
	} else {
		attrs = append(attrs, "sub", "local")
	}
	if err != nil {
		code := CodeAPIError
		var te *ToolError
		if errors.As(err, &te) {
			code = te.Code
		}
		s.logger.Warn("write_call_failed", append(attrs, "code", code)...)
		return
	}
	s.logger.Info("write_call", append(attrs, "file_id", f.ID)...)
}
