package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/server"
)

func main() {
	rootCmd := MakeGaiaD()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// MakeD returns the
func MakeGaiaD() *cobra.Command {
	cdc := app.MakeCodec()

	// Read in the configuration file for the sdk
	client.SetSDKConfig()

	// Get new server context
	ctx := server.NewDefaultContext()

	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "gaiad",
		Short:             "Gaia Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}
	rootCmd.AddCommand(gaiaInit.InitCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.AddGenesisAccountCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.GenTxCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.CollectGenTxsCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.TestnetFilesCmd(ctx, cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	return client.PrepareDMainCmd(rootCmd, "GA", app.DefaultNodeHome)
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewGaiaApp(logger, db, traceStore,
		baseapp.SetPruning(viper.GetString("pruning")),
		baseapp.SetMinimumFees(viper.GetString("minimum_fees")),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	gApp := app.NewGaiaApp(logger, db, traceStore)
	if height != -1 {
		err := gApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
	}
	return gApp.ExportAppStateAndValidators(forZeroHeight)
}
