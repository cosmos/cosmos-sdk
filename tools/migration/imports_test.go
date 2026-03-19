package migration

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestUpdateImports(t *testing.T) {
	importReplacements := []ImportReplacement{
		{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},
		{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant", AllPackages: true},
		{Old: "cosmossdk.io/x/evidence", New: "github.com/cosmos/cosmos-sdk/x/evidence", AllPackages: true},
		{Old: "cosmossdk.io/systemtests", New: "github.com/cosmos/cosmos-sdk/testutil/systemtests", AllPackages: true},
	}

	tests := []struct {
		name            string
		input           string
		doNotWantImport string
		wantImport      string
		wantAlias       string
		wantMod         bool
	}{
		{
			name: "vanity URL sub-package is updated",
			input: `package main
					import (
						"fmt"
						"cosmossdk.io/x/upgrade/types"
					)`,
			doNotWantImport: `"cosmossdk.io/x/upgrade/types"`,
			wantImport:      `"github.com/cosmos/cosmos-sdk/x/upgrade/types"`,
			wantMod:         true,
		},
		{
			name: "vanity URL root package is updated",
			input: `package main
					import (
						"cosmossdk.io/x/feegrant"
					)`,
			doNotWantImport: `"cosmossdk.io/x/feegrant"`,
			wantImport:      `"github.com/cosmos/cosmos-sdk/x/feegrant"`,
			wantMod:         true,
		},
		{
			name:       "no replacement needed",
			input:      `package main; import "fmt"`,
			wantImport: `"fmt"`,
			wantMod:    false,
		},
		{
			name: "already-migrated import is not double-replaced",
			input: `package main
					import "github.com/cosmos/cosmos-sdk/x/evidence/types"`,
			wantImport: `"github.com/cosmos/cosmos-sdk/x/evidence/types"`,
			wantMod:    false,
		},
		{
			name: "vanity systemtests package is updated",
			input: `package main
					import "cosmossdk.io/systemtests/assert"`,
			doNotWantImport: `"cosmossdk.io/systemtests/assert"`,
			wantImport:      `"github.com/cosmos/cosmos-sdk/testutil/systemtests/assert"`,
			wantMod:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			modified, err := updateImports(node, importReplacements)
			if err != nil {
				t.Fatalf("updateImports returned error: %v", err)
			}
			if modified != tt.wantMod {
				t.Errorf("expected modified=%v, got %v", tt.wantMod, modified)
			}

			importsToMaybeAlias := make(map[string]string)
			for _, imp := range node.Imports {
				if imp.Name == nil {
					importsToMaybeAlias[imp.Path.Value] = ""
				} else {
					importsToMaybeAlias[imp.Path.Value] = imp.Name.Name
				}
			}
			if tt.doNotWantImport != "" {
				if _, ok := importsToMaybeAlias[tt.doNotWantImport]; ok {
					t.Errorf("expected import %q to not be in the AST", tt.doNotWantImport)
				}
			}
			if _, ok := importsToMaybeAlias[tt.wantImport]; !ok {
				t.Errorf("expected import %q to be in the AST", tt.wantImport)
			}
			if tt.wantAlias != "" && importsToMaybeAlias[tt.wantImport] != tt.wantAlias {
				t.Errorf("expected alias %q for import %q, got %q", tt.wantAlias, tt.wantImport, importsToMaybeAlias[tt.wantImport])
			}
		})
	}
}

func TestUpdateImportsExcept(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		replacements    []ImportReplacement
		doNotWantImport string
		wantImport      string
		wantAlias       string
		wantMod         bool
	}{
		{
			name: "Except skips exact exception match",
			input: `package main
import "github.com/cometbft/cometbft/api"
`,
			replacements: []ImportReplacement{
				{Old: "github.com/cometbft/cometbft", New: "github.com/cometbft/cometbft/v2", AllPackages: true, Except: []string{"api"}},
			},
			wantImport: `"github.com/cometbft/cometbft/api"`,
			wantMod:    false,
		},
		{
			name: "Except skips sub-package of exception",
			input: `package main
import "github.com/cometbft/cometbft/api/v1"
`,
			replacements: []ImportReplacement{
				{Old: "github.com/cometbft/cometbft", New: "github.com/cometbft/cometbft/v2", AllPackages: true, Except: []string{"api"}},
			},
			wantImport: `"github.com/cometbft/cometbft/api/v1"`,
			wantMod:    false,
		},
		{
			name: "Except allows non-exception packages to be replaced",
			input: `package main
import "github.com/cometbft/cometbft/types"
`,
			replacements: []ImportReplacement{
				{Old: "github.com/cometbft/cometbft", New: "github.com/cometbft/cometbft/v2", AllPackages: true, Except: []string{"api"}},
			},
			doNotWantImport: `"github.com/cometbft/cometbft/types"`,
			wantImport:      `"github.com/cometbft/cometbft/v2/types"`,
			wantMod:         true,
		},
		{
			name: "exact match replacement adds defensive alias",
			input: `package main
import "cosmossdk.io/x/feegrant"
`,
			replacements: []ImportReplacement{
				{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant"},
			},
			wantImport: `"github.com/cosmos/cosmos-sdk/x/feegrant"`,
			wantAlias:  "feegrant",
			wantMod:    true,
		},
		{
			name: "exact match doesn't override existing alias",
			input: `package main
import fg "cosmossdk.io/x/feegrant"
`,
			replacements: []ImportReplacement{
				{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant"},
			},
			wantImport: `"github.com/cosmos/cosmos-sdk/x/feegrant"`,
			wantAlias:  "fg",
			wantMod:    true,
		},
		{
			name: "multiple Except values",
			input: `package main
import (
	"github.com/cometbft/cometbft/api"
	"github.com/cometbft/cometbft/proto"
	"github.com/cometbft/cometbft/types"
)
`,
			replacements: []ImportReplacement{
				{Old: "github.com/cometbft/cometbft", New: "github.com/cometbft/cometbft/v2", AllPackages: true, Except: []string{"api", "proto"}},
			},
			wantImport: `"github.com/cometbft/cometbft/v2/types"`,
			wantMod:    true,
		},
		{
			name:         "empty replacements returns unmodified",
			input:        `package main; import "fmt"`,
			replacements: []ImportReplacement{},
			wantImport:   `"fmt"`,
			wantMod:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			modified, err := updateImports(node, tt.replacements)
			if err != nil {
				t.Fatalf("updateImports returned error: %v", err)
			}
			if modified != tt.wantMod {
				t.Errorf("expected modified=%v, got %v", tt.wantMod, modified)
			}

			importsToMaybeAlias := make(map[string]string)
			for _, imp := range node.Imports {
				if imp.Name == nil {
					importsToMaybeAlias[imp.Path.Value] = ""
				} else {
					importsToMaybeAlias[imp.Path.Value] = imp.Name.Name
				}
			}
			if tt.doNotWantImport != "" {
				if _, ok := importsToMaybeAlias[tt.doNotWantImport]; ok {
					t.Errorf("expected import %q to not be in the AST", tt.doNotWantImport)
				}
			}
			if tt.wantImport != "" {
				if _, ok := importsToMaybeAlias[tt.wantImport]; !ok {
					t.Errorf("expected import %q to be in the AST", tt.wantImport)
				}
			}
			if tt.wantAlias != "" && importsToMaybeAlias[tt.wantImport] != tt.wantAlias {
				t.Errorf("expected alias %q for import %q, got %q", tt.wantAlias, tt.wantImport, importsToMaybeAlias[tt.wantImport])
			}
		})
	}
}
