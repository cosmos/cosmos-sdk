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

	importAliases := buildImportAliases(node)

	// create a map to store replacements we need to make
	replacements := make(map[ast.Stmt][]ast.Stmt)

	// find all statements that need to be replaced
	ast.Inspect(node, func(n ast.Node) bool {
		if stmt, ok := n.(*ast.ExprStmt); ok {
			callExpr, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				return true
			}

			selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkgIdent, ok := selectorExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			for _, replacement := range complexReplacements {
				alias, exists := importAliases[replacement.ImportPath]
				if !exists {
					continue
				}

				if pkgIdent.Name == alias && selectorExpr.Sel.Name == replacement.FuncName {
					log.Debug().Msgf("Marked %s.%s() call for replacement",
						pkgIdent.Name, replacement.FuncName)
					modified = true
					replacementStmts := replacement.ReplacementFunc(callExpr)
					replacements[stmt] = replacementStmts

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
		ast.Inspect(node, func(n ast.Node) bool {
			blockStmt, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			var newList []ast.Stmt
			for _, stmt := range blockStmt.List {
				if replacement, ok := replacements[stmt]; ok {
					newList = append(newList, replacement...)
					log.Debug().Msg("Replaced statement with multiple statements")
				} else {
					newList = append(newList, stmt)
				}
			}

			if len(newList) != len(blockStmt.List) {
				blockStmt.List = newList
			}

			return true
		})

		// clean up old imports that are no longer needed
		// (only when we actually replaced calls — otherwise the import may still be in use)
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
	}

	return modified, nil
}
