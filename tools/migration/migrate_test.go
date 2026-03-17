package migration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckImportWarnings(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		warnings     []ImportWarning
		wantWarnings int
		wantChanged  bool
		// wantImport verifies the import is still present after warning (no removal)
		wantImport string
		// wantMissingImport verifies the import was removed (AlsoRemove)
		wantMissingImport string
	}{
		{
			name: "warning without removal",
			input: `package main
import "cosmossdk.io/x/circuit"
`,
			warnings: []ImportWarning{
				{ImportPrefix: "cosmossdk.io/x/circuit", Message: "circuit module removed"},
			},
			wantWarnings: 1,
			wantChanged:  false,
			wantImport:   "cosmossdk.io/x/circuit",
		},
		{
			name: "warning with AlsoRemove",
			input: `package main
import "cosmossdk.io/x/group"
`,
			warnings: []ImportWarning{
				{ImportPrefix: "cosmossdk.io/x/group", Message: "group module removed", AlsoRemove: true},
			},
			wantWarnings:      1,
			wantChanged:       true,
			wantMissingImport: "cosmossdk.io/x/group",
		},
		{
			name: "warning matches prefix — sub-package import",
			input: `package main
import "cosmossdk.io/x/nft/keeper"
`,
			warnings: []ImportWarning{
				{ImportPrefix: "cosmossdk.io/x/nft", Message: "nft module removed", AlsoRemove: true},
			},
			wantWarnings:      1,
			wantChanged:       true,
			wantMissingImport: "cosmossdk.io/x/nft/keeper",
		},
		{
			name: "no match — different import",
			input: `package main
import "fmt"
`,
			warnings: []ImportWarning{
				{ImportPrefix: "cosmossdk.io/x/circuit", Message: "circuit removed"},
			},
			wantWarnings: 0,
			wantChanged:  false,
		},
		{
			name: "multiple warnings on same file",
			input: `package main
import (
	"cosmossdk.io/x/group"
	"cosmossdk.io/x/nft/types"
)
`,
			warnings: []ImportWarning{
				{ImportPrefix: "cosmossdk.io/x/group", Message: "group removed", AlsoRemove: true},
				{ImportPrefix: "cosmossdk.io/x/nft", Message: "nft removed", AlsoRemove: true},
			},
			wantWarnings: 2,
			wantChanged:  true,
		},
		{
			name: "named import with AlsoRemove",
			input: `package main
import groupkeeper "cosmossdk.io/x/group/keeper"
`,
			warnings: []ImportWarning{
				{ImportPrefix: "cosmossdk.io/x/group", Message: "group removed", AlsoRemove: true},
			},
			wantWarnings:      1,
			wantChanged:       true,
			wantMissingImport: "cosmossdk.io/x/group/keeper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test.go", tt.input, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			warnings, changed := checkImportWarnings("test.go", fset, node, tt.warnings)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("got %d warnings, want %d", len(warnings), tt.wantWarnings)
			}
			if changed != tt.wantChanged {
				t.Errorf("changed = %v, want %v", changed, tt.wantChanged)
			}

			if tt.wantImport != "" {
				found := false
				for _, imp := range node.Imports {
					if strings.Trim(imp.Path.Value, `"`) == tt.wantImport {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected import %q to still be present", tt.wantImport)
				}
			}

			if tt.wantMissingImport != "" {
				for _, imp := range node.Imports {
					if strings.Trim(imp.Path.Value, `"`) == tt.wantMissingImport {
						t.Errorf("expected import %q to be removed, but it's still present", tt.wantMissingImport)
					}
				}
			}
		})
	}
}

func TestUpdateGoModules(t *testing.T) {
	tests := []struct {
		name               string
		gomod              string
		updates            GoModUpdate
		removals           GoModRemoval
		replacements       []GoModReplacement
		additions          GoModAddition
		stripLocalReplaces bool
		wantContains       []string
		wantMissing        []string
	}{
		{
			name: "update module version",
			gomod: `module example.com/app

go 1.22

require (
	github.com/cosmos/cosmos-sdk v0.53.0
	github.com/cometbft/cometbft v0.38.0
)
`,
			updates: GoModUpdate{
				"github.com/cosmos/cosmos-sdk": "v0.54.0",
			},
			wantContains: []string{"github.com/cosmos/cosmos-sdk v0.54.0"},
			wantMissing:  []string{"v0.53.0"},
		},
		{
			name: "remove module requirement",
			gomod: `module example.com/app

go 1.22

require (
	cosmossdk.io/x/circuit v0.1.0
	github.com/cosmos/cosmos-sdk v0.53.0
)
`,
			removals:     GoModRemoval{"cosmossdk.io/x/circuit"},
			wantMissing:  []string{"cosmossdk.io/x/circuit"},
			wantContains: []string{"github.com/cosmos/cosmos-sdk"},
		},
		{
			name: "add module requirement",
			gomod: `module example.com/app

go 1.22

require (
	github.com/cosmos/cosmos-sdk v0.53.0
)
`,
			additions: GoModAddition{
				"cosmossdk.io/x/epochs": "v0.1.0",
			},
			wantContains: []string{"cosmossdk.io/x/epochs v0.1.0"},
		},
		{
			name: "add replace directive",
			gomod: `module example.com/app

go 1.22

require (
	github.com/cosmos/cosmos-sdk v0.53.0
)
`,
			replacements: []GoModReplacement{
				{Module: "github.com/cosmos/cosmos-sdk", Replacement: "github.com/cosmos/cosmos-sdk", Version: "v0.54.0-beta.1"},
			},
			wantContains: []string{"replace github.com/cosmos/cosmos-sdk"},
		},
		{
			name: "strip local path replaces",
			gomod: `module example.com/app

go 1.22

require (
	github.com/cosmos/cosmos-sdk v0.53.0
	cosmossdk.io/x/bank v0.1.0
)

replace (
	github.com/cosmos/cosmos-sdk => ../cosmos-sdk
	cosmossdk.io/x/bank => ./x/bank
)
`,
			stripLocalReplaces: true,
			wantMissing:        []string{"=> ../cosmos-sdk", "=> ./x/bank"},
		},
		{
			name: "no changes — empty operations",
			gomod: `module example.com/app

go 1.22

require github.com/cosmos/cosmos-sdk v0.53.0
`,
			wantContains: []string{"github.com/cosmos/cosmos-sdk v0.53.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			gomodPath := filepath.Join(tmpDir, "go.mod")
			if err := os.WriteFile(gomodPath, []byte(tt.gomod), 0o600); err != nil {
				t.Fatalf("write go.mod: %v", err)
			}

			if tt.updates == nil {
				tt.updates = GoModUpdate{}
			}
			if tt.additions == nil {
				tt.additions = GoModAddition{}
			}

			err := updateGoModules(
				[]string{gomodPath},
				tt.updates,
				tt.removals,
				tt.replacements,
				tt.additions,
				tt.stripLocalReplaces,
			)
			if err != nil {
				t.Fatalf("updateGoModules error: %v", err)
			}

			result, err := os.ReadFile(gomodPath)
			if err != nil {
				t.Fatalf("read result: %v", err)
			}
			output := string(result)

			for _, s := range tt.wantContains {
				if !strings.Contains(output, s) {
					t.Errorf("output should contain %q, got:\n%s", s, output)
				}
			}
			for _, s := range tt.wantMissing {
				if strings.Contains(output, s) {
					t.Errorf("output should NOT contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestMigrate(t *testing.T) {
	// End-to-end test: create a mini project dir with a Go file and go.mod,
	// run Migrate, and verify the output.
	tmpDir := t.TempDir()

	// Create go.mod
	gomod := `module example.com/app

go 1.22

require (
	cosmossdk.io/x/upgrade v0.1.0
	cosmossdk.io/x/circuit v0.1.0
)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// Create a Go source file with imports to rewrite
	goSrc := `package main

import (
	"fmt"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

func main() {
	fmt.Println(upgradetypes.ModuleName)
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goSrc), 0o600); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	// Create a file that should be removed
	anteGo := `package main
import circuitante "cosmossdk.io/x/circuit/ante"
func init() { _ = circuitante.NewCircuitBreakerDecorator }
`
	if err := os.WriteFile(filepath.Join(tmpDir, "ante.go"), []byte(anteGo), 0o600); err != nil {
		t.Fatalf("write ante.go: %v", err)
	}

	// Create a file for text replacements
	appGo := `package main
func setup() {
	app.BaseApp.GRPCQueryRouter()
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.go"), []byte(appGo), 0o600); err != nil {
		t.Fatalf("write app.go: %v", err)
	}

	args := MigrateArgs{
		ImportUpdates: []ImportReplacement{
			{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},
		},
		GoModRemoval: GoModRemoval{"cosmossdk.io/x/circuit"},
		GoModUpdates: GoModUpdate{
			"cosmossdk.io/x/upgrade": "v0.2.0",
		},
		FileRemovals: []FileRemoval{
			{FileName: "ante.go", ContainsMustMatch: "circuitante"},
		},
		TextReplacements: []TextReplacement{
			{Old: "app.BaseApp.GRPCQueryRouter()", New: "app.GRPCQueryRouter()"},
		},
	}

	err := Migrate(tmpDir, args)
	if err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	// Verify ante.go was removed
	if _, err := os.Stat(filepath.Join(tmpDir, "ante.go")); !os.IsNotExist(err) {
		t.Error("ante.go should have been removed")
	}

	// Verify main.go import was rewritten
	mainContent, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	mainStr := string(mainContent)
	if !strings.Contains(mainStr, "github.com/cosmos/cosmos-sdk/x/upgrade/types") {
		t.Errorf("main.go should contain rewritten import, got:\n%s", mainStr)
	}
	if strings.Contains(mainStr, "cosmossdk.io/x/upgrade") {
		t.Errorf("main.go should not contain old import, got:\n%s", mainStr)
	}

	// Verify app.go text replacement
	appContent, err := os.ReadFile(filepath.Join(tmpDir, "app.go"))
	if err != nil {
		t.Fatalf("read app.go: %v", err)
	}
	appStr := string(appContent)
	if !strings.Contains(appStr, "app.GRPCQueryRouter()") {
		t.Errorf("app.go should contain rewritten call, got:\n%s", appStr)
	}
	if strings.Contains(appStr, "app.BaseApp.GRPCQueryRouter()") {
		t.Errorf("app.go should not contain old call, got:\n%s", appStr)
	}

	// Verify go.mod was updated
	modContent, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	modStr := string(modContent)
	if strings.Contains(modStr, "cosmossdk.io/x/circuit") {
		t.Errorf("go.mod should not contain removed module, got:\n%s", modStr)
	}
	if !strings.Contains(modStr, "v0.2.0") {
		t.Errorf("go.mod should contain updated version, got:\n%s", modStr)
	}
}

func TestMigrateEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Migrate should succeed even with no files
	err := Migrate(tmpDir, MigrateArgs{})
	if err != nil {
		t.Fatalf("Migrate on empty dir should not error: %v", err)
	}
}

func TestCheckImportWarningsEmpty(t *testing.T) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", `package main; import "fmt"`, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Empty warnings list should return nothing
	warnings, changed := checkImportWarnings("test.go", fset, node, nil)
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(warnings))
	}
	if changed {
		t.Error("expected changed=false for empty warnings list")
	}
}

func TestBuildImportAliases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "named import",
			input: `package main; import gk "github.com/cosmos/cosmos-sdk/x/gov/keeper"`,
			want:  map[string]string{"github.com/cosmos/cosmos-sdk/x/gov/keeper": "gk"},
		},
		{
			name:  "unnamed import uses last segment",
			input: `package main; import "github.com/cosmos/cosmos-sdk/x/gov/keeper"`,
			want:  map[string]string{"github.com/cosmos/cosmos-sdk/x/gov/keeper": "keeper"},
		},
		{
			name:  "blank import",
			input: `package main; import _ "github.com/cosmos/cosmos-sdk/x/gov/keeper"`,
			want:  map[string]string{"github.com/cosmos/cosmos-sdk/x/gov/keeper": "_"},
		},
		{
			name: "multiple imports",
			input: `package main
import (
	"fmt"
	keeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)`,
			want: map[string]string{
				"fmt": "fmt",
				"github.com/cosmos/cosmos-sdk/x/bank/keeper": "keeper",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			got := buildImportAliases(node)
			for path, wantAlias := range tt.want {
				if gotAlias, ok := got[path]; !ok {
					t.Errorf("missing alias for %q", path)
				} else if gotAlias != wantAlias {
					t.Errorf("alias for %q = %q, want %q", path, gotAlias, wantAlias)
				}
			}
		})
	}
}

func TestExprToStringEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"call expression", "foo.Bar()", "foo.Bar(...)"},
		{"unary &", "&x", "&x"},
		{"index expression", "x[0]", "x[0]"},
		{"composite literal with type", "Foo{}", "Foo{...}"},
		{"basic literal int", "42", "42"},
		{"basic literal string", `"hello"`, `"hello"`},
		{"nested selector", "a.B.C", "a.B.C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := exprToString(expr)
			if got != tt.want {
				t.Errorf("exprToString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}

	// Test nil-like case: an expression type we don't handle returns ""
	unknownExpr := &ast.FuncLit{
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{},
	}
	if got := exprToString(unknownExpr); got != "" {
		t.Errorf("exprToString(FuncLit) = %q, want empty string", got)
	}
}

func TestSetExprPos(t *testing.T) {
	// Verify setExprPos doesn't panic on various expression types
	var pos token.Pos = 100

	ident := &ast.Ident{Name: "x"}
	setExprPos(ident, pos)
	if ident.NamePos != pos {
		t.Errorf("ident.NamePos = %v, want %v", ident.NamePos, pos)
	}

	sel := &ast.SelectorExpr{X: &ast.Ident{Name: "pkg"}, Sel: &ast.Ident{Name: "Func"}}
	setExprPos(sel, pos)
	if sel.Sel.NamePos != pos {
		t.Errorf("sel.Sel.NamePos = %v, want %v", sel.Sel.NamePos, pos)
	}

	star := &ast.StarExpr{X: &ast.Ident{Name: "T"}}
	setExprPos(star, pos)
	if star.Star != pos {
		t.Errorf("star.Star = %v, want %v", star.Star, pos)
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "f"}}
	setExprPos(call, pos)
	if call.Lparen != pos {
		t.Errorf("call.Lparen = %v, want %v", call.Lparen, pos)
	}

	lit := &ast.BasicLit{Value: "42"}
	setExprPos(lit, pos)
	if lit.ValuePos != pos {
		t.Errorf("lit.ValuePos = %v, want %v", lit.ValuePos, pos)
	}
}

func TestParseExprSafeErrors(t *testing.T) {
	// Invalid Go expression should return an error
	_, err := parseExprSafe("not a valid go expression !!!")
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}

	// Valid expression should succeed
	expr, err := parseExprSafe("foo.Bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr == nil {
		t.Fatal("expected non-nil expression")
	}
}

func TestMatchesCallPattern(t *testing.T) {
	input := `package main
func f() {
	storetypes.NewKVStoreKeys(a, b)
	app.ModuleManager.SetOrderEndBlockers(x, y)
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", input, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var calls []*ast.CallExpr
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			calls = append(calls, call)
		}
		return true
	})

	if len(calls) < 2 {
		t.Fatalf("expected at least 2 calls, got %d", len(calls))
	}

	// Test FuncPattern match
	if !matchesCallPattern(calls[0], CallArgRemoval{FuncPattern: "storetypes.NewKVStoreKeys"}) {
		t.Error("expected FuncPattern match for storetypes.NewKVStoreKeys")
	}
	if matchesCallPattern(calls[0], CallArgRemoval{FuncPattern: "other.Func"}) {
		t.Error("should not match wrong FuncPattern")
	}

	// Test MethodName match
	if !matchesCallPattern(calls[1], CallArgRemoval{MethodName: "SetOrderEndBlockers"}) {
		t.Error("expected MethodName match for SetOrderEndBlockers")
	}
	if matchesCallPattern(calls[1], CallArgRemoval{MethodName: "WrongMethod"}) {
		t.Error("should not match wrong MethodName")
	}
}

func TestMatchesPrecedingAssign(t *testing.T) {
	input := `package main
func f() {
	groupConfig := group.DefaultConfig()
	x = 42
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", input, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Find the block statement
	var stmts []ast.Stmt
	ast.Inspect(node, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			stmts = block.List
			return false
		}
		return true
	})

	if len(stmts) < 2 {
		t.Fatalf("expected at least 2 statements, got %d", len(stmts))
	}

	// First stmt is `:=` with groupConfig — should match
	if !matchesPrecedingAssign(stmts[0], "groupConfig") {
		t.Error("expected match for groupConfig := ...")
	}

	// Wrong variable name — should not match
	if matchesPrecedingAssign(stmts[0], "otherVar") {
		t.Error("should not match wrong variable name")
	}

	// Second stmt is `=` (not `:=`) — should not match
	if matchesPrecedingAssign(stmts[1], "x") {
		t.Error("should not match regular assignment (not :=)")
	}
}
