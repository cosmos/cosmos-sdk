package migration

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

// replaceDeprecatedExit converts deprecated.Exit("message") to:
// fmt.Println("message")
// os.Exit(1)
func replaceDeprecatedExit(call *ast.CallExpr) []ast.Stmt {
	args := call.Args

	printlnCall := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "fmt"},
				Sel: &ast.Ident{Name: "Println"},
			},
			Args: args,
		},
	}

	exitCall := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "os"},
				Sel: &ast.Ident{Name: "Exit"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.INT,
					Value: "1",
				},
			},
		},
	}

	return []ast.Stmt{printlnCall, exitCall}
}

func TestUpdateComplexFunctions(t *testing.T) {
	testReplacements := []ComplexFunctionReplacement{
		{
			ImportPath:      "github.com/example/deprecated",
			FuncName:        "Exit",
			RequiredImports: []string{"fmt", "os"},
			ReplacementFunc: replaceDeprecatedExit,
		},
	}

	tests := []struct {
		name              string
		input             string
		wantModified      bool
		wantImports       []string
		wantCallFragments []string
		notWantFragments  []string
	}{
		{
			name: "replace deprecated.Exit with fmt.Println + os.Exit",
			input: `
				package main
				import dep "github.com/example/deprecated"
				func main() {
					dep.Exit("goodbye")
				}
			`,
			wantModified: true,
			wantImports:  []string{`"fmt"`, `"os"`},
			wantCallFragments: []string{
				"fmt.Println(\"goodbye\")",
				"os.Exit(1)",
			},
			notWantFragments: []string{
				"dep.Exit",
			},
		},
		{
			name: "unrelated function is not modified",
			input: `
				package main
				import "fmt"
				func main() {
					fmt.Println("hello")
				}
			`,
			wantModified: false,
			wantCallFragments: []string{
				"fmt.Println(\"hello\")",
			},
			notWantFragments: []string{
				"os.Exit(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			modified, err := updateComplexFunctions(fset, node, testReplacements)
			if err != nil {
				t.Fatalf("updateComplexFunctions returned error: %v", err)
			}

			if modified != tt.wantModified {
				t.Errorf("expected modified = %v, got %v", tt.wantModified, modified)
			}

			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, node); err != nil {
				t.Fatalf("failed to print AST: %v", err)
			}
			output := buf.String()

			for _, frag := range tt.wantCallFragments {
				if !strings.Contains(output, frag) {
					t.Errorf("expected output to contain %q, but it did not.\nOutput:\n%s", frag, output)
				}
			}

			for _, frag := range tt.notWantFragments {
				if strings.Contains(output, frag) {
					t.Errorf("expected output to NOT contain %q, but it did.\nOutput:\n%s", frag, output)
				}
			}

			for _, imp := range tt.wantImports {
				if !strings.Contains(output, imp) {
					t.Errorf("expected import %q to be present, but it wasn't", imp)
				}
			}
		})
	}
}
