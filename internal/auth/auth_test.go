package auth

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/api/drive/v3"
)

// INV-1: by default the server asks Google only for read-only scopes.
func TestDefaultScopesAreReadOnly(t *testing.T) {
	for _, s := range Scopes() {
		if !strings.HasSuffix(s, ".readonly") {
			t.Errorf("default scope %q is not read-only", s)
		}
	}
}

// Write mode adds drive.file and nothing else: never the full drive scope.
func TestWriteScopesAddOnlyDriveFile(t *testing.T) {
	read, write := Scopes(), ScopesWrite()
	var extra []string
	for _, s := range write {
		if !slices.Contains(read, s) {
			extra = append(extra, s)
		}
	}
	if len(extra) != 1 || extra[0] != drive.DriveFileScope {
		t.Errorf("write mode adds %v, want only %s", extra, drive.DriveFileScope)
	}
	if slices.Contains(write, drive.DriveScope) {
		t.Error("write mode requests the full drive scope")
	}
	for _, s := range read {
		if !slices.Contains(write, s) {
			t.Errorf("write mode drops the default scope %q", s)
		}
	}
}
