package migration

import (
	"go/ast"
	"strings"

	"github.com/rs/zerolog/log"
)

// FunctionArgUpdate describes a function whose argument list has changed between versions.
// Currently only supports truncation (removing trailing arguments).
type FunctionArgUpdate struct {
	ImportPath  string // import path for the package containing the function
	FuncName    string // function name to update
	OldArgCount int    // old number of arguments
	NewArgCount int    // new number of arguments
}

// updateFunctionCalls finds and updates function calls that need argument changes.
// Currently, this only handles cases where function arguments were truncated.
// For example, http.New(s string, b int) -> http.New(s string)
func updateFunctionCalls(node *ast.File, functionUpdates []FunctionArgUpdate) (bool, error) {
	modified := false

	// build a map of import paths to their aliases in this file
	importAliases := buildImportAliases(node)

	// walk the AST to find function calls
	ast.Inspect(node, func(n ast.Node) bool {
		// look for function calls
		callExpr, ok := n.(*ast.CallExpr)
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

		// for each function update we want to apply
		for _, update := range functionUpdates {
			// check all possible aliases for this import path
			alias, exists := importAliases[update.ImportPath]
			if !exists {
				// this file doesn't import the package we're interested in
				continue
			}

			// if this is our target function and has the right number of arguments
			if pkgIdent.Name == alias &&
				selectorExpr.Sel.Name == update.FuncName &&
				len(callExpr.Args) == update.OldArgCount {

				if update.NewArgCount > update.OldArgCount {
					log.Warn().Msgf("NewArgCount (%d) > OldArgCount (%d) for %s.%s, skipping",
						update.NewArgCount, update.OldArgCount, pkgIdent.Name, update.FuncName)
					break
				}

				// truncate the argument list to the new count
				callExpr.Args = callExpr.Args[:update.NewArgCount]
				modified = true

				log.Debug().Msgf("Updated %s.%s() call to use %d arguments instead of %d",
					pkgIdent.Name, update.FuncName, update.NewArgCount, update.OldArgCount)

				break
			}
		}

		return true
	})

	return modified, nil
}

// buildImportAliases returns a map from import path to the alias used in the file.
// Extracted to avoid duplication across calls.go, structs.go, complex_functions.go.
func buildImportAliases(node *ast.File) map[string]string {
	importAliases := make(map[string]string)
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		if imp.Name != nil {
			importAliases[importPath] = imp.Name.Name
		} else {
			parts := strings.Split(importPath, "/")
			importAliases[importPath] = parts[len(parts)-1]
		}
	}
	return importAliases
}
