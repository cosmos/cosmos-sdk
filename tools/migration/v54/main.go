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

	log.Debug().Msgf("starting migration in dir %q", dir)

	args := migration.MigrateArgs{
		GoModRemoval:   removals,
		GoModUpdates:   moduleUpdates,
		ArgUpdates:     callUpdates,
		ComplexUpdates: complexReplacements,
		ImportUpdates:  importReplacements,
		TypeUpdates:    typeReplacements,
	}
	if err := migration.Migrate(dir, args); err != nil {
		panic(fmt.Sprintf("error migrating: %v\n", err))
	}
}
