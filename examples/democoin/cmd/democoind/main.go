package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/examples/democoin/app"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// init parameters
var CoolAppInit = server.AppInit{
	AppGenState: CoolAppGenState,
	AppGenTx:    server.SimpleAppGenTx,
}

// coolGenAppParams sets up the app_state and appends the cool app state
func CoolAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {
	appState, err = server.SimpleAppGenState(cdc, appGenTxs)
	if err != nil {
		return
	}

	key := "cool"
	value := json.RawMessage(`{
        "trend": "ice-cold"
      }`)

	appState, err = server.InsertKeyJSON(cdc, appState, key, value)
	if err != nil {
		return
	}

	key = "pow"
	value = json.RawMessage(`{
        "difficulty": 1,
        "count": 0
      }`)

	appState, err = server.InsertKeyJSON(cdc, appState, key, value)
	return
}

func newApp(ctx *sdk.ServerContext, db dbm.DB) abci.Application {
	return app.NewDemocoinApp(ctx, db)
}

func exportAppStateAndTMValidators(ctx *sdk.ServerContext, db dbm.DB) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	dapp := app.NewDemocoinApp(ctx, db)
	return dapp.ExportAppStateAndValidators()
}

func main() {
	cdc := app.MakeCodec()
	ctx := sdk.NewDefaultServerContext()

	rootCmd := &cobra.Command{
		Use:               "democoind",
		Short:             "Democoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, CoolAppInit,
		server.ConstructAppCreator(newApp, "democoin"),
		server.ConstructAppExporter(exportAppStateAndTMValidators, "democoin"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.democoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}
