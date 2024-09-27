package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/client/v2/offchain"
	corectx "cosmossdk.io/core/context"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	runtimev2 "cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/api/telemetry"
	"cosmossdk.io/server/v2/cometbft"
	"cosmossdk.io/server/v2/store"
	"cosmossdk.io/simapp/v2"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	genutilv2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2/cli"
)

func newApp[T transaction.Tx](logger log.Logger, viper *viper.Viper) serverv2.AppI[T] {
	viper.Set(serverv2.FlagHome, simapp.DefaultNodeHome)
	return serverv2.AppI[T](simapp.NewSimApp[T](logger, viper))
}

func initRootCmd[T transaction.Tx](
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	moduleManager *runtimev2.MM[T],
) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(moduleManager),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		NewTestnetCmd(moduleManager),
	)

	logger, err := serverv2.NewLogger(viper.New(), rootCmd.OutOrStdout())
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		genesisCommand(moduleManager),
		queryCommand(),
		txCommand(),
		keys.Commands(),
		offchain.OffChain(),
	)

	// wire server commands
	if err = serverv2.AddCommands(
		rootCmd,
		newApp,
		logger,
		initServerConfig(),
		cometbft.New(
			&genericTxDecoder[T]{txConfig},
			initCometOptions[T](),
			initCometConfig(),
		),
		grpc.New[T](),
		store.New[T](newApp),
		telemetry.New[T](),
	); err != nil {
		panic(err)
	}
}

// genesisCommand builds genesis-related `simd genesis` command.
func genesisCommand[T transaction.Tx](
	moduleManager *runtimev2.MM[T],
	cmds ...*cobra.Command,
) *cobra.Command {
	cmd := v2.Commands(
		moduleManager.Modules()[genutiltypes.ModuleName].(genutil.AppModule),
		moduleManager,
		appExport[T],
	)

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
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
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
func appExport[T transaction.Tx](
	ctx context.Context,
	height int64,
	jailAllowedAddrs []string,
) (genutilv2.ExportedApp, error) {
	value := ctx.Value(corectx.ViperContextKey)
	viper, ok := value.(*viper.Viper)
	if !ok {
		return genutilv2.ExportedApp{},
			fmt.Errorf("incorrect viper type %T: expected *viper.Viper in context", value)
	}
	value = ctx.Value(corectx.LoggerContextKey)
	logger, ok := value.(log.Logger)
	if !ok {
		return genutilv2.ExportedApp{},
			fmt.Errorf("incorrect logger type %T: expected log.Logger in context", value)
	}

	// overwrite the FlagInvCheckPeriod
	viper.Set(server.FlagInvCheckPeriod, 1)
	viper.Set(serverv2.FlagHome, simapp.DefaultNodeHome)

	var simApp *simapp.SimApp[T]
	if height != -1 {
		simApp = simapp.NewSimApp[T](logger, viper)

		if err := simApp.LoadHeight(uint64(height)); err != nil {
			return genutilv2.ExportedApp{}, err
		}
	} else {
		simApp = simapp.NewSimApp[T](logger, viper)
	}

	return simApp.ExportAppStateAndValidators(jailAllowedAddrs)
}

var _ transaction.Codec[transaction.Tx] = &genericTxDecoder[transaction.Tx]{}

type genericTxDecoder[T transaction.Tx] struct {
	txConfig client.TxConfig
}

// Decode implements transaction.Codec.
func (t *genericTxDecoder[T]) Decode(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}

// DecodeJSON implements transaction.Codec.
func (t *genericTxDecoder[T]) DecodeJSON(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxJSONDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}
