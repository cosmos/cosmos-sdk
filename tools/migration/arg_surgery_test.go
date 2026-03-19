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

func TestUpdateArgSurgeryAST(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		surgeries    []ArgSurgeryWithAST
		wantModified bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name: "remove arg and append synthesized call wrapping it",
			input: `package main
import govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
func f() {
	govkeeper.NewKeeper(cdc, storeService, acctKeeper, bankKeeper, stakingKeeper, distrKeeper, router, config, authority)
}`,
			surgeries: []ArgSurgeryWithAST{
				{
					ImportPath:  "github.com/cosmos/cosmos-sdk/x/gov/keeper",
					FuncName:    "NewKeeper",
					OldArgCount: -1,
					Transform: func(pkgAlias string, args []ast.Expr) []ast.Expr {
						if len(args) < 9 {
							return args
						}
						stakingKeeper := args[4]
						newArgs := make([]ast.Expr, 0, 9)
						newArgs = append(newArgs, args[0:4]...)
						newArgs = append(newArgs, args[5:9]...)
						newArgs = append(newArgs, &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: pkgAlias},
								Sel: &ast.Ident{Name: "NewDefaultCalculateVoteResultsAndVotingPower"},
							},
							Args: []ast.Expr{stakingKeeper},
						})
						return newArgs
					},
				},
			},
			wantModified: true,
			// stakingKeeper (pos 4) should be removed from main args
			wantMissing: []string{"stakingKeeper, distrKeeper"},
			// should appear wrapped in the new function call
			wantContains: []string{
				"NewDefaultCalculateVoteResultsAndVotingPower",
				"stakingKeeper",
				"bankKeeper, distrKeeper", // bankKeeper is now followed by distrKeeper (stakingKeeper removed)
			},
		},
		{
			name: "no match for wrong function name",
			input: `package main
import govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
func f() {
	govkeeper.OtherFunc(a, b, c)
}`,
			surgeries: []ArgSurgeryWithAST{
				{
					ImportPath:  "github.com/cosmos/cosmos-sdk/x/gov/keeper",
					FuncName:    "NewKeeper",
					OldArgCount: -1,
					Transform:   func(_ string, args []ast.Expr) []ast.Expr { return args },
				},
			},
			wantModified: false,
		},
		{
			name: "no match when import not present",
			input: `package main
import "fmt"
func f() {
	fmt.Println("hello")
}`,
			surgeries: []ArgSurgeryWithAST{
				{
					ImportPath:  "github.com/cosmos/cosmos-sdk/x/gov/keeper",
					FuncName:    "NewKeeper",
					OldArgCount: -1,
					Transform:   func(_ string, args []ast.Expr) []ast.Expr { return args },
				},
			},
			wantModified: false,
		},
		{
			name: "matches with explicit alias",
			input: `package main
import gk "github.com/cosmos/cosmos-sdk/x/gov/keeper"
func f() {
	gk.NewKeeper(a, b)
}`,
			surgeries: []ArgSurgeryWithAST{
				{
					ImportPath:  "github.com/cosmos/cosmos-sdk/x/gov/keeper",
					FuncName:    "NewKeeper",
					OldArgCount: -1,
					Transform: func(_ string, args []ast.Expr) []ast.Expr {
						// Just reverse args for testing
						return []ast.Expr{args[1], args[0]}
					},
				},
			},
			wantModified: true,
			wantContains: []string{"gk.NewKeeper(b, a)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			modified, err := updateArgSurgeryAST(node, tt.surgeries)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if modified != tt.wantModified {
				t.Errorf("modified = %v, want %v", modified, tt.wantModified)
			}

			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, node); err != nil {
				t.Fatalf("print error: %v", err)
			}
			output := buf.String()
			normalizedOutput := strings.Join(strings.Fields(output), " ")

			for _, s := range tt.wantContains {
				normalizedWant := strings.Join(strings.Fields(s), " ")
				if !strings.Contains(normalizedOutput, normalizedWant) {
					t.Errorf("output should contain %q, got:\n%s", s, output)
				}
			}
			for _, s := range tt.wantMissing {
				normalizedWant := strings.Join(strings.Fields(s), " ")
				if strings.Contains(normalizedOutput, normalizedWant) {
					t.Errorf("output should NOT contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestUpdateArgSurgery(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		surgeries    []ArgSurgery
		wantModified bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name: "remove arg by position",
			input: `package main
import keeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
func f() {
	keeper.NewKeeper(a, b, c, d)
}`,
			surgeries: []ArgSurgery{
				{
					ImportPath:         "github.com/cosmos/cosmos-sdk/x/bank/keeper",
					FuncName:           "NewKeeper",
					OldArgCount:        4,
					RemoveArgPositions: []int{1}, // remove b
				},
			},
			wantModified: true,
			wantContains: []string{"a, c, d"},
			wantMissing:  []string{"a, b, c"},
		},
		{
			name: "$ARG{N} placeholder — bare reference to removed arg",
			input: `package main
import keeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
func f() {
	keeper.NewKeeper(a, b, c, d)
}`,
			surgeries: []ArgSurgery{
				{
					ImportPath:         "github.com/cosmos/cosmos-sdk/x/bank/keeper",
					FuncName:           "NewKeeper",
					OldArgCount:        4,
					RemoveArgPositions: []int{2}, // remove c
					AppendArgs:         []string{"$ARG{2}"},
				},
			},
			wantModified: true,
			// c should be removed from its original position and appended at the end
			wantContains: []string{"a, b, d, c"},
			wantMissing:  []string{"a, b, c, d"},
		},
		{
			name: "$ARG{N} placeholder — wrap removed arg in function call",
			input: `package main
import govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
func f() {
	govkeeper.NewKeeper(cdc, storeService, acctKeeper, bankKeeper, stakingKeeper, distrKeeper, router, config, authority)
}`,
			surgeries: []ArgSurgery{
				{
					ImportPath:         "github.com/cosmos/cosmos-sdk/x/gov/keeper",
					FuncName:           "NewKeeper",
					OldArgCount:        9,
					RemoveArgPositions: []int{4}, // remove stakingKeeper
					AppendArgs:         []string{"govkeeper.NewDefaultCalculateVoteResultsAndVotingPower($ARG{4})"},
				},
			},
			wantModified: true,
			wantContains: []string{
				"NewDefaultCalculateVoteResultsAndVotingPower",
				"stakingKeeper",
				"bankKeeper, distrKeeper", // stakingKeeper removed from middle
			},
			wantMissing: []string{"bankKeeper, stakingKeeper"},
		},
		{
			name: "$ARG{N} placeholder — no placeholders behaves as before",
			input: `package main
import keeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
func f() {
	keeper.NewKeeper(a, b, c)
}`,
			surgeries: []ArgSurgery{
				{
					ImportPath:         "github.com/cosmos/cosmos-sdk/x/bank/keeper",
					FuncName:           "NewKeeper",
					OldArgCount:        3,
					RemoveArgPositions: []int{2}, // remove c
					AppendArgs:         []string{"nil"},
				},
			},
			wantModified: true,
			wantContains: []string{"a, b, nil"},
			wantMissing:  []string{"a, b, c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			modified, err := updateArgSurgery(node, tt.surgeries)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if modified != tt.wantModified {
				t.Errorf("modified = %v, want %v", modified, tt.wantModified)
			}

			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, node); err != nil {
				t.Fatalf("print error: %v", err)
			}
			output := buf.String()

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
