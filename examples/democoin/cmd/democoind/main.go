package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/examples/democoin/app"
	"github.com/cosmos/cosmos-sdk/server"
)

// init parameters
var CoolAppInit = server.AppInit{
	AppGenState: CoolAppGenState,
	AppGenTx:    server.SimpleAppGenTx,
}

// coolGenAppParams sets up the app_state and appends the cool app state
func CoolAppGenState(cdc *codec.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {
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
        "difficulty": "1",
        "count": "0"
      }`)

	appState, err = server.InsertKeyJSON(cdc, appState, key, value)
	return
}

func newApp(logger log.Logger, db dbm.DB, _ io.Writer) abci.Application {
	return app.NewDemocoinApp(logger, db)
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB, _ io.Writer) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	dapp := app.NewDemocoinApp(logger, db)
	return dapp.ExportAppStateAndValidators()
}

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "democoind",
		Short:             "Democoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(gaiaInit.InitCmd(ctx, cdc, CoolAppInit))
	rootCmd.AddCommand(gaiaInit.TestnetFilesCmd(ctx, cdc, CoolAppInit))

	server.AddCommands(ctx, cdc, rootCmd, CoolAppInit,
		newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.democoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}
