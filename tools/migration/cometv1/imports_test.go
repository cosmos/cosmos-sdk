package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestUpdateImports(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		doNotWantImport string
		wantImport      string
		wantAlias       string
		wantMod         bool
	}{
		{
			name:            "import path is updated and alias is added",
			input:           `package main; import "github.com/cometbft/cometbft/proto/tendermint/types"`,
			doNotWantImport: `"github.com/cometbft/cometbft/proto/tendermint/types"`,
			wantImport:      `"github.com/cometbft/cometbft/api/cometbft/types/v1"`,
			wantAlias:       "types",
			wantMod:         true,
		},
		{
			name:            "alias stays intact",
			input:           `package main; import blah "github.com/cometbft/cometbft/proto/tendermint/types"`,
			doNotWantImport: `"github.com/cometbft/cometbft/proto/tendermint/types"`,
			wantImport:      `"github.com/cometbft/cometbft/api/cometbft/types/v1"`,
			wantAlias:       "blah",
			wantMod:         true,
		},
		{
			name: "block import with one replacement",
			input: `package main
					import (
						"fmt"
						"github.com/cometbft/cometbft/proto/tendermint/crypto"
					)`,
			doNotWantImport: `"github.com/cometbft/cometbft/proto/tendermint/crypto"`,
			wantImport:      `"github.com/cometbft/cometbft/api/cometbft/crypto/v1"`,
			wantAlias:       "crypto",
			wantMod:         true,
		},
		{
			name:       "no replacement needed",
			input:      `package main; import "fmt"`,
			wantImport: `"fmt"`,
			wantMod:    false,
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

			// importsToMaybeAlias is a mapping of imports to an alias if it has one.
			importsToMaybeAlias := make(map[string]string)
			for _, imp := range node.Imports {
				if imp.Name == nil {
					importsToMaybeAlias[imp.Path.Value] = ""
				} else {
					importsToMaybeAlias[imp.Path.Value] = imp.Name.Name
				}
			}
			if _, ok := importsToMaybeAlias[tt.doNotWantImport]; ok {
				t.Errorf("expected import %q to not be in the AST", tt.input)
			}
			if _, ok := importsToMaybeAlias[tt.wantImport]; !ok {
				t.Errorf("expected import %q to be in the AST", tt.wantImport)
			}
			if importsToMaybeAlias[tt.wantImport] != tt.wantAlias {
				t.Errorf("expected alias %q for import %q, got %q", tt.wantAlias, tt.wantImport, importsToMaybeAlias[tt.wantImport])
			}
		})
	}
}
