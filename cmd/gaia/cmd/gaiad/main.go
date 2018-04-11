package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
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

func generateApp(rootDir string, logger log.Logger) (abci.Application, error) {
	dataDir := filepath.Join(rootDir, "data")
	db, err := dbm.NewGoLevelDB("gaia", dataDir)
	if err != nil {
		return nil, err
	}
	bapp := app.NewBasecoinApp(logger, db)
	//dbAcc, err := dbm.NewGoLevelDB("gaia-acc", dataDir)
	//if err != nil {
	//return nil, err
	//}
	//dbIBC, err := dbm.NewGoLevelDB("gaia-ibc", dataDir)
	//if err != nil {
	//return nil, err
	//}
	//dbStake, err := dbm.NewGoLevelDB("gaia-stake", dataDir)
	//if err != nil {
	//return nil, err
	//}
	//dbs := map[string]dbm.DB{
	//"main":  dbMain,
	//"acc":   dbAcc,
	//"ibc":   dbIBC,
	//"stake": dbStake,
	//}
	//bapp := app.NewGaiaApp(logger, dbs)
	return bapp, nil
}

func main() {
	server.AddCommands(rootCmd, app.DefaultGenAppState, generateApp, context)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.gaiad")
	executor := cli.PrepareBaseCmd(rootCmd, "GA", rootDir)
	executor.Execute()
}
