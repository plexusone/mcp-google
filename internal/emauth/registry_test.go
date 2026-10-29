package emauth_test

import (
	"sort"
	"testing"

	"github.com/plexusone/mcp-google/internal/emauth"
	"github.com/plexusone/mcp-google/skills/docs"
	"github.com/plexusone/mcp-google/skills/sheets"
	"github.com/plexusone/mcp-google/skills/slides"
	"github.com/plexusone/omniskill/skill"
)

// registeredToolNames returns the names of every tool the server registers.
// Skills are built with a nil HTTP client and never initialized: Tools()
// only describes the tools and does not touch the Google API.
func registeredToolNames(t *testing.T) []string {
	t.Helper()
	skills := []skill.Skill{
		docs.New(nil),
		sheets.New(nil),
		slides.New(nil),
	}
	seen := map[string]string{}
	var names []string
	for _, s := range skills {
		for _, tool := range s.Tools() {
			name := tool.Name()
			if prev, dup := seen[name]; dup {
				t.Fatalf("tool %q registered by both %s and %s skills", name, prev, s.Name())
			}
			seen[name] = s.Name()
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		t.Fatal("no registered tools found")
	}
	sort.Strings(names)
	return names
}

// TestToolScopeRegistry enforces that every registered tool has exactly one
// required policy scope, and that the authorization map names no tool the
// server does not register.
func TestToolScopeRegistry(t *testing.T) {
	names := registeredToolNames(t)
	scopes := emauth.ToolScopes()

	registered := make(map[string]bool, len(names))
	for _, name := range names {
		registered[name] = true
		scope, ok := scopes[name]
		if !ok {
			t.Errorf("registered tool %q has no entry in the authorization map", name)
			continue
		}
		if scope == "" {
			t.Errorf("registered tool %q maps to an empty scope", name)
		}
	}

	for name := range scopes {
		if !registered[name] {
			t.Errorf("authorization map has entry %q for a tool that is not registered", name)
		}
	}
}
