package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/simpleGov/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/wire"
	tmtypes "github.com/tendermint/tendermint/types"
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

func newApp(logger log.Logger, db dbm.DB, storeTracer io.Writer) abci.Application {
	return app.NewSimpleGovApp(logger, db, baseapp.SetPruning(viper.GetString("pruning")))
}

func exportAppState(logger log.Logger, db dbm.DB, storeTracer io.Writer) (json.RawMessage, []tmtypes.GenesisValidator, error) {
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
