package migration

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

func TestUpdateFunctionCalls(t *testing.T) {
	testUpdates := []FunctionArgUpdate{
		{
			ImportPath:  "github.com/cosmos/cosmos-sdk/x/bank/keeper",
			FuncName:    "NewKeeper",
			OldArgCount: 6,
			NewArgCount: 5,
		},
	}

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
				import "github.com/cosmos/cosmos-sdk/x/bank/keeper"
				func main() {
					keeper.NewKeeper("a", "b", "c", "d", "e", "f")
				}
			`,
			wantModified:   true,
			wantCallString: `keeper.NewKeeper("a", "b", "c", "d", "e")`,
		},
		{
			name: "function with explicit alias is updated",
			input: `
				package main
				import bk "github.com/cosmos/cosmos-sdk/x/bank/keeper"
				func main() {
					bk.NewKeeper("a", "b", "c", "d", "e", "f")
				}
			`,
			wantModified:   true,
			wantCallString: `bk.NewKeeper("a", "b", "c", "d", "e")`,
		},
		{
			name: "function with wrong arg count is not updated",
			input: `
				package main
				import "github.com/cosmos/cosmos-sdk/x/bank/keeper"
				func main() {
					keeper.NewKeeper("a")
				}
			`,
			wantModified:   false,
			wantCallString: `keeper.NewKeeper("a")`,
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

			modified, err := updateFunctionCalls(node, testUpdates)
			if err != nil {
				t.Fatalf("updateFunctionCalls returned error: %v", err)
			}

			if modified != tt.wantModified {
				t.Errorf("expected modified=%v, got %v", tt.wantModified, modified)
			}

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
