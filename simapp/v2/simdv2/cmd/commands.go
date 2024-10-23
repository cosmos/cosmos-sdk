package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/client/v2/offchain"
	coreserver "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	runtimev2 "cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2/cli"
)

func newApp[T transaction.Tx](logger log.Logger, viper *viper.Viper) serverv2.AppI[T] {
	viper.Set(serverv2.FlagHome, simapp.DefaultNodeHome)
	return serverv2.AppI[T](simapp.NewSimApp[T](logger, viper))
}

type configWriter interface {
	WriteConfig(filename string) error
}

func initRootCmd[T transaction.Tx](
	rootCmd *cobra.Command,
	logger log.Logger,
	globalAppConfig coreserver.ConfigMap,
	txConfig client.TxConfig,
	moduleManager *runtimev2.MM[T],
	app *simapp.SimApp[T],
) (configWriter, error) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(moduleManager),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		NewTestnetCmd(moduleManager),
	)

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		genesisCommand(moduleManager, app),
		queryCommand(),
		txCommand(),
		keys.Commands(),
		offchain.OffChain(),
	)

	// wire server commands
	//if _, err := serverv2.AddCommands(
	//	rootCmd,
	//	newApp,
	//	initServerConfig(),
	//	cometbft.New(
	//		&genericTxDecoder[T]{txConfig},
	//		initCometOptions[T](),
	//		initCometConfig(),
	//	),
	//	grpc.New[T](),
	//	serverstore.New[T](),
	//	telemetry.New[T](),
	//	rest.New[T](),
	//); err != nil {
	//	panic(err)
	//}

	return nil, nil
}

// genesisCommand builds genesis-related `simd genesis` command.
func genesisCommand[T transaction.Tx](
	moduleManager *runtimev2.MM[T],
	app *simapp.SimApp[T],
) *cobra.Command {
	var genTxValidator func([]transaction.Msg) error
	if moduleManager != nil {
		genTxValidator = moduleManager.Modules()[genutiltypes.ModuleName].(genutil.AppModule).GenTxValidator()
	}
	cmd := v2.Commands(
		genTxValidator,
		moduleManager,
		app,
	)

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
