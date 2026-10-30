package skills

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const apiPrefix = "google.golang.org/api/"

// TestSkillsCallGoogleOnlyThroughGogoogle enforces INV-3: skills hold no
// Google API request logic. A skill may name Google API types in its
// signatures, but it must not build an API client, set client options, or
// finish an API call itself; gogoogle does that.
func TestSkillsCallGoogleOnlyThroughGogoogle(t *testing.T) {
	fset := token.NewFileSet()
	checked := 0
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		checked++
		apiPkgs := map[string]string{} // local name -> import path
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if p == apiPrefix+"option" {
				t.Errorf("%s: imports %s: client options belong in gogoogle", fset.Position(imp.Pos()), p)
			}
			if strings.HasPrefix(p, apiPrefix) {
				name := filepath.Base(p)
				if strings.HasPrefix(name, "v") && len(name) <= 3 { // docs/v1 -> docs
					name = filepath.Base(filepath.Dir(p))
				}
				if imp.Name != nil {
					name = imp.Name.Name
				}
				apiPkgs[name] = p
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if id, ok := sel.X.(*ast.Ident); ok {
				if _, isAPI := apiPkgs[id.Name]; isAPI && strings.HasPrefix(sel.Sel.Name, "New") {
					t.Errorf("%s: calls %s.%s: build services through gogoogle", fset.Position(call.Pos()), id.Name, sel.Sel.Name)
				}
			}
			// X.Do() with no arguments finishes a Google API request.
			if sel.Sel.Name == "Do" && len(call.Args) == 0 {
				t.Errorf("%s: calls .Do(): API requests belong in gogoogle", fset.Position(call.Pos()))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no skill source files were checked")
	}
}
