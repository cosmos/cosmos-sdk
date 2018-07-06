package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
)

func main() {
	cdc := app.MakeCodec()
	ctx := sdk.NewDefaultServerContext()
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
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func newApp(ctx *sdk.ServerContext, db dbm.DB) abci.Application {
	return app.NewGaiaApp(ctx, db)
}

func exportAppStateAndTMValidators(ctx *sdk.ServerContext, db dbm.DB) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	gapp := app.NewGaiaApp(ctx, db)
	return gapp.ExportAppStateAndValidators()
}
