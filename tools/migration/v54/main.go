package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	migration "github.com/cosmos/cosmos-sdk/tools/migrate"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
		log.Debug().Msgf("checking dir %q", dir)
		if _, err := os.Stat(dir); err != nil {
			panic(fmt.Sprintf("error with directory: %q: %v\n", dir, err))
		}
	}

	log.Debug().Msgf("starting v53 -> v54 migration in dir %q", dir)

	args := migration.MigrateArgs{
		// --- go.mod operations ---
		GoModRemoval:      removals,
		GoModAddition:     additions,
		GoModReplacements: replacements,
		GoModUpdates:      moduleUpdates,

		// --- AST: import rewrites ---
		ImportUpdates:  importReplacements,
		ImportWarnings: importWarnings,

		// --- AST: type/struct changes ---
		TypeUpdates:        typeReplacements,
		FieldRemovals:      fieldRemovals,
		FieldModifications: fieldModifications,

		// --- AST: function arg changes ---
		ArgUpdates:     callUpdates,
		ArgSurgeries:   argSurgeries,
		CallArgEdits:   callArgEdits,
		ComplexUpdates: complexUpdates,

		// --- AST: statement/block removal ---
		StatementRemovals: statementRemovals,
		MapEntryRemovals:  mapEntryRemovals,

		// --- Text-level replacements (post-AST) ---
		TextReplacements: textReplacements,

		// --- File operations ---
		FileRemovals: fileRemovals,
	}
	if err := migration.Migrate(dir, args); err != nil {
		panic(fmt.Sprintf("error migrating: %v\n", err))
	}

	log.Info().Msg("migration complete — run `goimports -w . && go mod tidy` to finalize")
}
