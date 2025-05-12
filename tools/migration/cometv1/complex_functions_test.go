package main

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

func TestUpdateComplexFunctions(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantModified      bool
		wantImports       []string
		wantCallFragments []string
		notWantFragments  []string
	}{
		{
			name: "replace cmtos.Exit with fmt.Println + os.Exit",
			input: `
				package main
				import cmtos "github.com/cometbft/cometbft/libs/os"
				func main() {
					cmtos.Exit("goodbye")
				}
			`,
			wantModified: true,
			wantImports:  []string{`"fmt"`, `"os"`},
			wantCallFragments: []string{
				"fmt.Println(\"goodbye\")",
				"os.Exit(1)",
			},
			notWantFragments: []string{
				"cmtos.Exit",
			},
		},
		{
			name: "function from wrong package is ignored",
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
		{
			name: "selector with correct name but wrong import alias is ignored",
			input: `
				package main
				import other "github.com/cometbft/cometbft/libs/os"
				func main() {
					notcmtos.Exit("fail")
				}
			`,
			wantModified: false,
			notWantFragments: []string{
				`fmt.Println("fail")`,
				"os.Exit(1)",
			},
		},
		{
			name: "replace function when import has no alias",
			input: `
		package main
		import "github.com/cometbft/cometbft/libs/os"
		func main() {
			os.Exit("see ya")
		}
	`,
			wantModified: true,
			wantImports:  []string{`"fmt"`, `"os"`},
			wantCallFragments: []string{
				"fmt.Println(\"see ya\")",
				"os.Exit(1)",
			},
			notWantFragments: []string{
				"github.com/cometbft/cometbft/libs/os",
				"os.Exit(\"see ya\")", // the original bad call
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

			modified, err := updateComplexFunctions(fset, node)
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
