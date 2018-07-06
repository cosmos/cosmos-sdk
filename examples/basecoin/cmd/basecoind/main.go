package main

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)

func main() {
	cdc := app.MakeCodec()
	ctx := sdk.NewDefaultServerContext()

	rootCmd := &cobra.Command{
		Use:               "basecoind",
		Short:             "Basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, server.DefaultAppInit,
		server.ConstructAppCreator(newApp, "basecoin"),
		server.ConstructAppExporter(exportAppStateAndTMValidators, "basecoin"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.basecoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)

	err := executor.Execute()
	if err != nil {
		// Note: Handle with #870
		panic(err)
	}
}

func newApp(ctx *sdk.ServerContext, db dbm.DB) abci.Application {
	return app.NewBasecoinApp(ctx, db)
}

func exportAppStateAndTMValidators(ctx *sdk.ServerContext, db dbm.DB) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	bapp := app.NewBasecoinApp(ctx, db)
	return bapp.ExportAppStateAndValidators()
}
