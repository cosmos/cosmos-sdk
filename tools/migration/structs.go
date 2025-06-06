package migration

import (
	"go/ast"
	"strings"

	"github.com/rs/zerolog/log"
)

type TypeReplacement struct {
	ImportPath string // import path for the package containing the type
	OldType    string // old type name (without package prefix)
	NewType    string // new type name (without package prefix)
}

// updateStructs finds and replaces all references to the specified struct types.
func updateStructs(node *ast.File, typeReplacements []TypeReplacement) (bool, error) {
	modified := false
	// first, build a map of import paths to their aliases in this file
	importAliases := make(map[string]string) // maps import path to its alias
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")

		// determine the package alias
		var alias string
		if imp.Name != nil {
			// explicit alias
			alias = imp.Name.Name
		} else {
			// default alias is the last part of the import path
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}

		importAliases[importPath] = alias
	}

	// now walk the AST and find all selector expressions to replace
	ast.Inspect(node, func(n ast.Node) bool {
		// check if this is a selector expression (e.g., abci.RequestInitChain)
		selectorExpr, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// get the identifier (package name/alias) - e.g., "abci" part in abci.RequestInitChain
		ident, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		for _, replacement := range typeReplacements {
			alias, exists := importAliases[replacement.ImportPath]
			if !exists {
				// this file doesn't import the package we're interested in
				continue
			}

			// check if this selector matches our target type
			if ident.Name == alias && selectorExpr.Sel.Name == replacement.OldType {
				// we found a match, replace the type name
				selectorExpr.Sel.Name = replacement.NewType
				modified = true
				log.Debug().Msgf("Replaced %s.%s with %s.%s",
					ident.Name, replacement.OldType,
					ident.Name, replacement.NewType)
			}
		}

		return true
	})

	return modified, nil
}
