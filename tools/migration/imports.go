package migration

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/rs/zerolog/log"
)

type ImportReplacement struct {
	// Old is the old import path
	Old string
	// New is the new import path.
	New string
	// AllPackages notes whether all packages from Old should be replaced to New.
	// For example, cosmossdk.io/x/upgrade/<ANY_SUB_PACKAGE> should be changed to github.com/cosmos/cosmos-sdk/x/upgrade/<ANY_SUB_PACKAGE>
	AllPackages bool
	// Except defines packages that should be ignored for replacements.
	// For example:
	// github.com/cometbft/cometbft recently upgraded to cometbft/v2
	// however; we also have github.com/cometbft/cometbft/api which is its own go.mod, and should not be replaced.
	Except []string
}

func updateImports(node *ast.File, replacements []ImportReplacement) (bool, error) {
	modified := false
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		for _, replacement := range replacements {
			for _, exception := range replacement.Except {
				if importPath == replacement.Old+"/"+exception ||
					strings.HasPrefix(importPath, replacement.Old+"/"+exception+"/") {
					hasException = true
					break
				}
			}
			} else if importPath == replacement.Old {
				log.Debug().Msgf("updated import %s to %s", importPath, replacement.New)
				imp.Path.Value = fmt.Sprintf(`"%s"`, replacement.New)
				// Defensively set an alias to the last segment of the original import path
				// so that existing code using e.g. `types.Foo` doesn't break when the import
				// path changes to something like `.../v1`.
				if imp.Name == nil {
					paths := strings.Split(importPath, "/")
					imp.Name = &ast.Ident{
						Name: paths[len(paths)-1],
					}
				}
				modified = true
				break
			}
		}
	}
	return modified, nil
}
