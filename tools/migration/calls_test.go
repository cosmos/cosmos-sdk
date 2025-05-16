package migration

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

// test data
var (
	functionUpdates = []FunctionArgUpdate{
		{
			ImportPath:  "github.com/cometbft/cometbft/rpc/client/http",
			FuncName:    "New",
			OldArgCount: 2,
			NewArgCount: 1,
		},
	}
)

func TestUpdateFunctionCalls(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantModified   bool
		wantCallString string
	}{
		{
			name: "function with implicit alias and old arg count is updated",
			input: `
				package main
				import "github.com/cometbft/cometbft/rpc/client/http"
				func main() {
					http.New("arg1", "arg2")
				}
			`,
			wantModified:   true,
			wantCallString: `http.New("arg1")`,
		},
		{
			name: "function with explicit alias is updated",
			input: `
				package main
				import cmt "github.com/cometbft/cometbft/rpc/client/http"
				func main() {
					cmt.New("arg1", "arg2")
				}
			`,
			wantModified:   true,
			wantCallString: `cmt.New("arg1")`,
		},
		{
			name: "function with wrong arg count is not updated",
			input: `
				package main
				import "github.com/cometbft/cometbft/rpc/client/http"
				func main() {
					http.New("arg1")
				}
			`,
			wantModified:   false,
			wantCallString: `http.New("arg1")`,
		},
		{
			name: "unrelated function call is not modified",
			input: `
				package main
				import "fmt"
				func main() {
					fmt.Println("hello")
				}
			`,
			wantModified:   false,
			wantCallString: `fmt.Println("hello")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			modified, err := updateFunctionCalls(node, functionUpdates)
			if err != nil {
				t.Fatalf("updateFunctionCalls returned error: %v", err)
			}

			if modified != tt.wantModified {
				t.Errorf("expected modified=%v, got %v", tt.wantModified, modified)
			}

			// print resulting AST back to source to check changes
			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, node); err != nil {
				t.Fatalf("failed to print AST: %v", err)
			}
			output := buf.String()

			if !strings.Contains(output, tt.wantCallString) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.wantCallString, output)
			}
		})
	}
}
