package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/cmd/cosmos-sdk-cli/app"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/wire"
)

// init parameters
var MyAwesomeProjectAppInit = server.AppInit{
	AppGenState: MyAwesomeProjectAppGenState,
	AppGenTx:    server.SimpleAppGenTx,
}

// GenAppParams sets up the app_state, append any other app specific components.
func MyAwesomeProjectAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {
	appState, err = server.SimpleAppGenState(cdc, appGenTxs)
	if err != nil {
		return
	}

	return
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewMyAwesomeProjectApp(logger, db)
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	dapp := app.NewMyAwesomeProjectApp(logger, db)
	return dapp.ExportAppStateAndValidators()
}

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "myawesomeprojectd",
		Short:             "MyAwesomeProject Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, MyAwesomeProjectAppInit,
		server.ConstructAppCreator(newApp, "myawesomeproject"),
		server.ConstructAppExporter(exportAppStateAndTMValidators, "myawesomeproject"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.myawesomeprojectd")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}
