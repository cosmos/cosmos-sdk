package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/simpleGov/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/wire"
)

// SimpleGovAppInit initial parameters
var SimpleGovAppInit = server.AppInit{
	AppGenState: SimpleGovAppGenState,
	AppGenTx:    server.SimpleAppGenTx,
}

// SimpleGovAppGenState sets up the app_state and appends the simpleGov app state
func SimpleGovAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {
	appState, err = server.SimpleAppGenState(cdc, appGenTxs)
	if err != nil {
		return
	}
	return
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewSimpleGovApp(logger, db)
}

func exportAppState(logger log.Logger, db dbm.DB) (json.RawMessage, error) {
	dapp := app.NewSimpleGovApp(logger, db)
	return dapp.ExportAppStateJSON()
}

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "simplegovd",
		Short:             "Simple Governance Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, SimpleGovAppInit,
		server.ConstructAppCreator(newApp, "simplegov"),
		server.ConstructAppExporter(exportAppState, "simplegov"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.simplegovd")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	executor.Execute()
}
