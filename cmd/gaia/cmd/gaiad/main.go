package main

import (
	"encoding/json"
	"io"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "gaiad",
		Short:             "Gaia Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}
	rootCmd.AddCommand(gaiaInit.InitCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.CollectGenTxsCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.TestnetFilesCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.GenTxCmd(ctx, cdc))
	rootCmd.AddCommand(gaiaInit.AddGenesisAccountCmd(ctx, cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "GA", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewGaiaApp(logger, db, traceStore, true,
		baseapp.SetPruning(viper.GetString("pruning")),
		baseapp.SetMinimumFees(viper.GetString("minimum_fees")),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	gApp := app.NewGaiaApp(logger, db, traceStore, false)
	if height != -1 {
		err := gApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
	}
	return gApp.ExportAppStateAndValidators(forZeroHeight)
}
