package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "gaiad",
		Short:             "Gaia Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, app.GaiaAppInit(),
		server.ConstructAppCreator(newApp, "gaia"),
		server.ConstructAppExporter(exportAppStateAndTMValidators, "gaia"))

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "GA", app.DefaultNodeHome)
	executor.Execute()
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewGaiaApp(logger, db)
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	gapp := app.NewGaiaApp(logger, db)
	return gapp.ExportAppStateAndValidators()
}
