package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/version"
)

// basecoindCmd is the entry point for this binary
var (
	context = server.NewContext(nil, nil)
	rootCmd = &cobra.Command{
		Use:               "basecoind",
		Short:             "Basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(context),
	}
)

// defaultOptions sets up the app_state for the
// default genesis file. It is a server.GenAppState
// used for `basecoind init`.
func defaultOptions(args ...string) (json.RawMessage, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("Expected 2 args: address and coin denom")
	}
	opts := fmt.Sprintf(`{
      "accounts": [{
        "address": "%s",
        "coins": [
          {
            "denom": "mycoin",
            "amount": 9007199254740992
          }
        ]
      }]
    }`, args)
	return json.RawMessage(opts), nil
}

func generateApp(rootDir string, logger log.Logger) (abci.Application, error) {
	dataDir := filepath.Join(rootDir, "data")
	dbMain, err := dbm.NewGoLevelDB("basecoin", dataDir)
	if err != nil {
		return nil, err
	}
	dbAcc, err := dbm.NewGoLevelDB("basecoin-acc", dataDir)
	if err != nil {
		return nil, err
	}
	dbIBC, err := dbm.NewGoLevelDB("basecoin-ibc", dataDir)
	if err != nil {
		return nil, err
	}
	dbStaking, err := dbm.NewGoLevelDB("basecoin-staking", dataDir)
	if err != nil {
		return nil, err
	}
	dbs := map[string]dbm.DB{
		"main":    dbMain,
		"acc":     dbAcc,
		"ibc":     dbIBC,
		"staking": dbStaking,
	}
	bapp := app.NewBasecoinApp(logger, dbs)
	return bapp, nil
}

func main() {
	rootCmd.AddCommand(
		server.InitCmd(defaultOptions, context),
		server.StartCmd(generateApp, context),
		server.UnsafeResetAllCmd(context),
		server.ShowNodeIDCmd(context),
		server.ShowValidatorCmd(context),
		version.VersionCmd,
	)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.basecoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	executor.Execute()
}
