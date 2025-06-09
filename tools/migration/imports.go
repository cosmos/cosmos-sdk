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
	// AllPackages notes whether all packages from Old should be replaced to True.
	// For example, cosmossdk.io/x/upgrade/<ANY_SUB_PACKAGE> should be changed to github.com/cosmos/cosmos-sdk/x/upgrade/<ANY_SUB_PACKAGE>
	AllPackages bool
	// Except defines packages that should be ignored for replacements.
	// for example:
	// github.com/cometbft/cometbft recently upgraded to cometbft/v2
	// however; we also have github.com/cometbft/combetbft/api which is its own go.mod, and should not be replaced.
	Except []string
}

func updateImports(node *ast.File, replacements []ImportReplacement) (bool, error) {
	modified := false
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		for _, replacement := range replacements {
			if replacement.AllPackages {
				// importPath = cosmossdk.io/x/upgrade/types
				// replacement = cosmossdk.io/x/upgrade
				hasException := false
				for _, exception := range replacement.Except {
					if strings.Contains(importPath, "/"+exception) {
						hasException = true
						break
					}
				}
				if hasException {
					continue
				}
				// importPath = github.com/cometbft/cometbft/v2/types/foo
				// replacement = github.com/cometbft/cometbft/v2
				// second conditional is to prevent appending duplicate v2's to the same import on multiple migraiton runs.
				if strings.HasPrefix(importPath, replacement.Old) && !strings.HasPrefix(importPath, replacement.New) {
					subPackage := strings.TrimPrefix(importPath, replacement.Old)
					imp.Path.Value = fmt.Sprintf(`"%s%s"`, replacement.New, subPackage)
					modified = true
				}
			} else {
				// if we found an old import, we update it to the new one.
				if importPath == replacement.Old {
					log.Debug().Msgf("updated import %s to %s", importPath, replacement.New)
					imp.Path.Value = fmt.Sprintf(`"%s"`, replacement.New)
					// import.Name is the import alias. if it's nil, we defensively change it to the final token of the
					// original import, as that is 99.99% used as the dot selector.
					// for example, lets say we have:
					// github.com/foo/bar
					// and we update to
					// github.com/foo/bar/v1
					// this would invalidate all code using bar.Whatever
					// so we need to update its import alias to bar.
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
	}
	return modified, nil
}
