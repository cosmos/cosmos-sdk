package migration

import (
	"go/ast"

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

	// reuse the shared import alias builder
	importAliases := buildImportAliases(node)

	// walk the AST and find all selector expressions to replace
	ast.Inspect(node, func(n ast.Node) bool {
		selectorExpr, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		for _, replacement := range typeReplacements {
			alias, exists := importAliases[replacement.ImportPath]
			if !exists {
				continue
			}

			if ident.Name == alias && selectorExpr.Sel.Name == replacement.OldType {
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
