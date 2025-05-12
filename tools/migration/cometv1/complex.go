package main

import (
	"go/ast"
	"go/token"

	migration "github.com/cosmos/cosmos-sdk/tools/migrate"
)

var complexReplacements = []migration.ComplexFunctionReplacement{
	{
		ImportPath:      "github.com/cometbft/cometbft/libs/os",
		FuncName:        "Exit",
		RequiredImports: []string{"fmt", "os"},
		ReplacementFunc: replaceLibsOsExit,
	},
}

// replaceLibsOsExit converts cmtos.Exit("message") to:
// fmt.Println("message")
// os.Exit(1)
func replaceLibsOsExit(call *ast.CallExpr) []ast.Stmt {
	// extract the message argument from the original function call
	args := call.Args

	// create fmt.Println(message) statement
	printlnCall := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "fmt"},
				Sel: &ast.Ident{Name: "Println"},
			},
			Args: args,
		},
	}

	// create os.Exit(1) statement
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
