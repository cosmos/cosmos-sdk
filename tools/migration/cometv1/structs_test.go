package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestUpdateStructs_ContextVariations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string // the expected *new* type name to confirm replacement worked
	}{
		{
			name: "Used in variable declaration",
			code: `
				package test
				import abci "github.com/cometbft/cometbft/abci/types"
				var req abci.RequestInitChain
			`,
			expected: "InitChainRequest",
		},
		{
			name: "Used in function parameter",
			code: `
				package test
				import abci "github.com/cometbft/cometbft/abci/types"
				func Handle(req abci.ResponseEcho) {}
			`,
			expected: "EchoResponse",
		},
		{
			name: "Used as function return type",
			code: `
				package test
				import abci "github.com/cometbft/cometbft/abci/types"
				func NewRequest() abci.RequestFlush {
					return abci.RequestFlush{}
				}
			`,
			expected: "FlushRequest",
		},
		{
			name: "Used in struct field",
			code: `
				package test
				import abci "github.com/cometbft/cometbft/abci/types"
				type MyHandler struct {
					Info abci.ResponseInfo
				}
			`,
			expected: "InfoResponse",
		},
		{
			name: "Used in type conversion / composite literal",
			code: `
				package test
				import abci "github.com/cometbft/cometbft/abci/types"
				func Convert(x interface{}) {
					_ = x.(abci.RequestCommit)
				}
			`,
			expected: "CommitRequest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			modified, err := updateStructs(node, typeReplacements)
			if err != nil {
				t.Fatalf("updateStructs error: %v", err)
			}
			if !modified {
				t.Fatal("Expected modification, but none occurred")
			}

			// check that expected type appears in the updated AST
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
