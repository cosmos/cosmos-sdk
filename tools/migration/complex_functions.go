package migration

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/ast/astutil"
)

type ComplexFunctionReplacement struct {
	ImportPath      string      // import path for the package containing the function
	FuncName        string      // function name to update
	RequiredImports []string    // new imports required for the replacement
	ReplacementFunc ReplaceFunc // function that generates replacement nodes
}

// ReplaceFunc is a function that takes an *ast.CallExpr and returns a slice of replacement statements
type ReplaceFunc func(call *ast.CallExpr) []ast.Stmt

// updateComplexFunctions handles replacing function calls with multiple statements.
// For example, cmtos.Exit("foo") -> fmt.Println("foo"); os.Exit(1)
func updateComplexFunctions(fset *token.FileSet, node *ast.File, complexReplacements []ComplexFunctionReplacement) (bool, error) {
	modified := false

	// map import paths to their aliases in this file
	importAliases := make(map[string]string) // maps import path to its alias/name
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")

		if imp.Name != nil {
			// explicit alias
			importAliases[importPath] = imp.Name.Name
		} else {
			// default name is the last part of the import path
			parts := strings.Split(importPath, "/")
			importAliases[importPath] = parts[len(parts)-1]
		}
	}

	// create a map to store replacements we need to make
	replacements := make(map[ast.Stmt][]ast.Stmt)

	// find all statements that need to be replaced
	ast.Inspect(node, func(n ast.Node) bool {
		// we're only interested in statement-level nodes that might contain function calls
		switch stmt := n.(type) {
		case *ast.ExprStmt:
			// check if the expression is a call expression
			callExpr, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				return true
			}

			// check if it's a selector expression (package.Function)
			selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			// get the package identifier
			pkgIdent, ok := selectorExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			// for each complex replacement
			for _, replacement := range complexReplacements {
				// check if this import is used in the file
				alias, exists := importAliases[replacement.ImportPath]
				if !exists {
					continue
				}

				// check that we have a selector function from the package with the alias we expect it in.
				if pkgIdent.Name == alias && selectorExpr.Sel.Name == replacement.FuncName {
					log.Debug().Msgf("Marked %s.%s() call for replacement",
						pkgIdent.Name, replacement.FuncName)
					modified = true
					// generate replacement statements
					replacementStmts := replacement.ReplacementFunc(callExpr)
					replacements[stmt] = replacementStmts

					// update imports
					for _, imp := range replacement.RequiredImports {
						astutil.AddImport(fset, node, imp)
					}
				}
			}
		}
		return true
	})

	// apply statement replacements by finding their parent block statements
	if modified {
		// add statements
		ast.Inspect(node, func(n ast.Node) bool {
			blockStmt, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			// check if any statements in this block need to be replaced
			var newList []ast.Stmt
			for _, stmt := range blockStmt.List {
				if replacement, ok := replacements[stmt]; ok {
					// replace the statement with our new statements
					newList = append(newList, replacement...)
					log.Debug().Msg("Replaced statement with multiple statements")
				} else {
					// keep the original statement
					newList = append(newList, stmt)
				}
			}

			// update the block's statement list
			if len(newList) != len(blockStmt.List) {
				blockStmt.List = newList
			}

			return true
		})
	}

	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		if slices.ContainsFunc(complexReplacements, func(c ComplexFunctionReplacement) bool {
			return importPath == c.ImportPath
		}) {
			if imp.Name != nil {
				astutil.DeleteNamedImport(fset, node, imp.Name.Name, importPath)
			} else {
				astutil.DeleteImport(fset, node, importPath)
			}
		}
	}

	return modified, nil
}
