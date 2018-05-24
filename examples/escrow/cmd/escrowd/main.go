package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/escrow/app"
	"github.com/cosmos/cosmos-sdk/server"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "escrowd",
		Short:             "Escrow Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, server.DefaultAppInit,
		server.ConstructAppCreator(newApp, "escrow"),
		server.ConstructAppExporter(exportAppState, "escrow"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.escrowd")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	executor.Execute()
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewEscrowApp(logger, db)
}

func exportAppState(logger log.Logger, db dbm.DB) (json.RawMessage, error) {
	bapp := app.NewEscrowApp(logger, db)
	return bapp.ExportAppStateJSON()
}
