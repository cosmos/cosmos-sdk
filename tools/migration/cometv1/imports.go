package main

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/rs/zerolog/log"
)

type ImportReplacement struct {
	Old string
	New string
}

var (
	importReplacements = []ImportReplacement{
		{Old: "github.com/cometbft/cometbft/proto/tendermint/types", New: "github.com/cometbft/cometbft/api/cometbft/types/v1"},
		{Old: "github.com/cometbft/cometbft/proto/tendermint/crypto", New: "github.com/cometbft/cometbft/api/cometbft/crypto/v1"},
		{Old: "github.com/cometbft/cometbft/proto/tendermint/state", New: "github.com/cometbft/cometbft/api/cometbft/state/v1"},
	}
)

func updateImports(node *ast.File, replacements []ImportReplacement) (bool, error) {
	modified := false
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		for _, replacement := range replacements {
			// if we found an old import, we update it to the new one.
			if importPath == replacement.Old {
				log.Debug().Msgf("updated import %s to %s", importPath, replacement.New)
				imp.Path.Value = fmt.Sprintf(`"%s"`, replacement.New)
				// import.Name is the import alias. if it's nil, we defensively change it to the final token of the
				// original import, as that is 99.99% used as the dot selector.
				if imp.Name == nil {
					paths := strings.Split(importPath, "/")
					imp.Name = &ast.Ident{
						Name: paths[len(paths)-1],
					}
				}
				modified = true
			}
		}
	}
	return modified, nil
}
