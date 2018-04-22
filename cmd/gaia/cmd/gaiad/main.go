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

func generateApp(rootDir string, logger log.Logger) (abci.Application, error) {
	dataDir := filepath.Join(rootDir, "data")
	db, err := dbm.NewGoLevelDB("gaia", dataDir)
	if err != nil {
		return nil, err
	}
	bapp := app.NewGaiaApp(logger, db)
	return bapp, nil
}

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()
	rootCmd := &cobra.Command{
		Use:               "gaiad",
		Short:             "Gaia Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, app.GaiaGenAppState, generateApp)

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.gaiad")
	executor := cli.PrepareBaseCmd(rootCmd, "GA", rootDir)
	executor.Execute()
}
