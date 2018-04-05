package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/server"
)

// rootCmd is the entry point for this binary
var (
	context = server.NewDefaultContext()
	rootCmd = &cobra.Command{
		Use:               "gaiad",
		Short:             "Gaia Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(context),
	}
)

// TODO: distinguish from basecoin
func generateApp(rootDir string, logger log.Logger) (abci.Application, error) {
	dataDir := filepath.Join(rootDir, "data")
	dbMain, err := dbm.NewGoLevelDB("gaia", dataDir)
	if err != nil {
		return nil, err
	}
	dbAcc, err := dbm.NewGoLevelDB("gaia-acc", dataDir)
	if err != nil {
		return nil, err
	}
	dbIBC, err := dbm.NewGoLevelDB("gaia-ibc", dataDir)
	if err != nil {
		return nil, err
	}
	dbStaking, err := dbm.NewGoLevelDB("gaia-staking", dataDir)
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
	server.AddCommands(rootCmd, server.DefaultGenAppState, generateApp, context)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "GA", os.ExpandEnv("$HOME/.gaiad"))
	executor.Execute()
}
