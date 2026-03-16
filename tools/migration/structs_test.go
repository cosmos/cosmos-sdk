package migration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestUpdateStructs_ContextVariations(t *testing.T) {
	testReplacements := []TypeReplacement{
		{
			ImportPath: "github.com/cosmos/cosmos-sdk/store/types",
			OldType:    "CommitMultiStore",
			NewType:    "RootStore",
		},
	}

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "Used in variable declaration",
			code: `
				package test
				import storetypes "github.com/cosmos/cosmos-sdk/store/types"
				var s storetypes.CommitMultiStore
			`,
			expected: "RootStore",
		},
		{
			name: "Used in function parameter",
			code: `
				package test
				import storetypes "github.com/cosmos/cosmos-sdk/store/types"
				func Init(s storetypes.CommitMultiStore) {}
			`,
			expected: "RootStore",
		},
		{
			name: "Used as function return type",
			code: `
				package test
				import storetypes "github.com/cosmos/cosmos-sdk/store/types"
				func GetStore() storetypes.CommitMultiStore {
					return nil
				}
			`,
			expected: "RootStore",
		},
		{
			name: "Used in struct field",
			code: `
				package test
				import storetypes "github.com/cosmos/cosmos-sdk/store/types"
				type App struct {
					Store storetypes.CommitMultiStore
				}
			`,
			expected: "RootStore",
		},
		{
			name: "Used in type assertion",
			code: `
				package test
				import storetypes "github.com/cosmos/cosmos-sdk/store/types"
				func Convert(x interface{}) {
					_ = x.(storetypes.CommitMultiStore)
				}
			`,
			expected: "RootStore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			modified, err := updateStructs(node, testReplacements)
			if err != nil {
				t.Fatalf("updateStructs error: %v", err)
			}
			if !modified {
				t.Fatal("Expected modification, but none occurred")
			}

			var found bool
			ast.Inspect(node, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if ok && sel.Sel.Name == tt.expected {
					found = true
					return false
				}
				return true
			})

			if !found {
				t.Errorf("Expected updated type %q not found in AST", tt.expected)
			}
		})
	}
}
