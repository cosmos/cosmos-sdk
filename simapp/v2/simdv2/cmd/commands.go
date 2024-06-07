package cmd

import (
	"errors"
	"fmt"
	"io"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/client/v2/offchain"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	runtimev2 "cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/cometbft"
	"cosmossdk.io/simapp/v2"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	authcmd "cosmossdk.io/x/auth/client/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

var _ transaction.Codec[transaction.Tx] = &temporaryTxDecoder{}

type temporaryTxDecoder struct {
	txConfig client.TxConfig
}

// Decode implements transaction.Codec.
func (t *temporaryTxDecoder) Decode(bz []byte) (transaction.Tx, error) {
	return t.txConfig.TxDecoder()(bz)
}

// DecodeJSON implements transaction.Codec.
func (t *temporaryTxDecoder) DecodeJSON(bz []byte) (transaction.Tx, error) {
	return t.txConfig.TxJSONDecoder()(bz)
}

func newApp(
	viper *viper.Viper,
	logger log.Logger,
) serverv2.App[transaction.Tx] {
	sa := simapp.NewSimApp(logger, viper)
	return serverv2.App[transaction.Tx]{Application: sa, Store: sa.GetStore()}
}

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	_ codectypes.InterfaceRegistry,
	_ codec.Codec,
	moduleManager *runtimev2.MM,
	v1moduleManager *module.Manager,
) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(moduleManager),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		// pruning.Cmd(newApp),
		// snapshot.Cmd(newApp),
	)

	// Add empty server struct here for writing default config
	err := serverv2.AddCommands(
		rootCmd,
		newApp,
		log.NewNopLogger(),
		cometbft.New(&temporaryTxDecoder{txConfig}),
		grpc.NewGRPCServer(),
	)
	if err != nil {
		panic(fmt.Sprintf("Add cmd, %v", err))
	}

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		genesisCommand(txConfig, v1moduleManager, appExport),
		queryCommand(),
		txCommand(),
		keys.Commands(),
		offchain.OffChain(),
	)
}

// genesisCommand builds genesis-related `simd genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(
	txConfig client.TxConfig,
	moduleManager *module.Manager,
	appExport servertypes.AppExporter,
	cmds ...*cobra.Command,
) *cobra.Command {
	cmd := genutilcli.Commands(txConfig, moduleManager, appExport)

	for _, subCmd := range cmds {
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.QueryEventForTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

// appExport creates a new simapp (optionally at a given height) and exports state.
func appExport(
	logger log.Logger,
	_ dbm.DB,
	_ io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// this check is necessary as we use the flag in x/upgrade.
	// we can exit more gracefully by checking the flag here.
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(server.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	var simApp *simapp.SimApp
	if height != -1 {
		simApp = simapp.NewSimApp(logger, appOpts)

		if err := simApp.LoadHeight(uint64(height)); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		simApp = simapp.NewSimApp(logger, appOpts)
	}

	return simApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
