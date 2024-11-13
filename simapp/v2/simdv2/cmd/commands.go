package cmd

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"cosmossdk.io/client/v2/offchain"
	coreserver "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	runtimev2 "cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/api/rest"
	telemetry "cosmossdk.io/server/v2/api/telemetry/prometheus"
	"cosmossdk.io/server/v2/cometbft"
	serverstore "cosmossdk.io/server/v2/store"
	"cosmossdk.io/simapp/v2"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
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

// CommandDependencies is a struct that contains all the dependencies needed to initialize the root command.
// an alternative design could fetch these even later from the command context
type CommandDependencies[T transaction.Tx] struct {
	GlobalConfig  coreserver.ConfigMap
	TxConfig      client.TxConfig
	ModuleManager *runtimev2.MM[T]
	SimApp        *simapp.SimApp[T]
	Consensus     serverv2.ServerComponent[T]
}

func InitRootCmd[T transaction.Tx](
	rootCmd *cobra.Command,
	logger log.Logger,
	deps CommandDependencies[T],
) (serverv2.ConfigWriter, error) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(deps.ModuleManager),
		genesisCommand(deps.ModuleManager, deps.SimApp),
		NewTestnetCmd(deps.ModuleManager),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		// add keybase, auxiliary RPC, query, genesis, and tx child commands
		queryCommand(),
		txCommand(),
		keys.Commands(),
		offchain.OffChain(),
	)

	// build CLI skeleton for initial config parsing or a client application invocation
	if deps.SimApp == nil {
		if deps.Consensus == nil {
			deps.Consensus = cometbft.NewWithConfigOptions[T](initCometConfig())
		}
		return serverv2.AddCommands[T](
			rootCmd,
			logger,
			io.NopCloser(nil),
			deps.GlobalConfig,
			initServerConfig(),
			deps.Consensus,
			&grpc.Server[T]{},
			&serverstore.Server[T]{},
			&telemetry.Server[T]{},
			&rest.Server[T]{},
		)
	}

	// build full app!
	simApp := deps.SimApp
	grpcServer, err := grpc.New[T](logger, simApp.InterfaceRegistry(), simApp.QueryHandlers(), simApp.Query, deps.GlobalConfig)
	if err != nil {
		return nil, err
	}
	// store component (not a server)
	storeComponent, err := serverstore.New[T](simApp.Store(), deps.GlobalConfig)
	if err != nil {
		return nil, err
	}
	restServer, err := rest.New[T](logger, simApp.App.AppManager, deps.GlobalConfig)
	if err != nil {
		return nil, err
	}

	// consensus component
	if deps.Consensus == nil {
		deps.Consensus, err = cometbft.New(
			logger,
			simApp.Name(),
			simApp.Store(),
			simApp.App.AppManager,
			simApp.App.QueryHandlers(),
			simApp.App.SchemaDecoderResolver(),
			&genericTxDecoder[T]{deps.TxConfig},
			deps.GlobalConfig,
			initCometOptions[T](),
		)
		if err != nil {
			return nil, err
		}
	}
	telemetryServer, err := telemetry.New[T](deps.GlobalConfig, logger)
	if err != nil {
		return nil, err
	}

	// wire server commands
	return serverv2.AddCommands[T](
		rootCmd,
		logger,
		simApp,
		deps.GlobalConfig,
		initServerConfig(),
		deps.Consensus,
		grpcServer,
		storeComponent,
		telemetryServer,
		restServer,
	)
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

func RootCommandPersistentPreRun(clientCtx client.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// set the default command outputs
		cmd.SetOut(cmd.OutOrStdout())
		cmd.SetErr(cmd.ErrOrStderr())

		clientCtx = clientCtx.WithCmdContext(cmd.Context())
		clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}

		customClientTemplate, customClientConfig := initClientConfig()
		clientCtx, err = config.CreateClientConfig(
			clientCtx, customClientTemplate, customClientConfig)
		if err != nil {
			return err
		}

		if err = client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
			return err
		}

		return nil
	}
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
