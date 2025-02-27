package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"cosmossdk.io/client/v2/offchain"
	coreserver "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	runtimev2 "cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	grpcserver "cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/api/grpcgateway"
	"cosmossdk.io/server/v2/api/rest"
	"cosmossdk.io/server/v2/api/swagger"
	"cosmossdk.io/server/v2/api/telemetry"
	"cosmossdk.io/server/v2/cometbft"
	serverstore "cosmossdk.io/server/v2/store"
	"cosmossdk.io/simapp/v2"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/docs"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdktelemetry "github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/version"
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
	ClientContext client.Context
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
		version.NewVersionCommand(),
	)

	// build CLI skeleton for initial config parsing or a client application invocation
	if deps.SimApp == nil {
		return serverv2.AddCommands[T](
			rootCmd,
			logger,
			io.NopCloser(nil),
			deps.GlobalConfig,
			initServerConfig(),
			cometbft.NewWithConfigOptions[T](initCometConfig()),
			&grpcserver.Server[T]{},
			&serverstore.Server[T]{},
			&telemetry.Server[T]{},
			&rest.Server[T]{},
			&grpcgateway.Server[T]{},
			&swagger.Server[T]{},
		)
	}

	// build full app!
	simApp := deps.SimApp

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
	consensusServer, err := cometbft.New(
		logger,
		simApp.Name(),
		simApp.Store(),
		simApp.App.AppManager,
		cometbft.AppCodecs[T]{
			AppCodec:              simApp.AppCodec(),
			TxCodec:               &client.DefaultTxDecoder[T]{TxConfig: deps.TxConfig},
			LegacyAmino:           deps.ClientContext.LegacyAmino,
			ConsensusAddressCodec: deps.ClientContext.ConsensusAddressCodec,
		},
		simApp.App.QueryHandlers(),
		simApp.App.SchemaDecoderResolver(),
		initCometOptions[T](),
		deps.GlobalConfig,
	)
	if err != nil {
		return nil, err
	}

	telemetryServer, err := telemetry.New[T](logger, sdktelemetry.EnableTelemetry, deps.GlobalConfig)
	if err != nil {
		return nil, err
	}

	swaggerServer, err := swagger.New[T](logger, docs.GetSwaggerFS(), deps.GlobalConfig)
	if err != nil {
		return nil, err
	}

	grpcServer, err := grpcserver.New[T](
		logger,
		simApp.InterfaceRegistry(),
		simApp.QueryHandlers(),
		simApp.Query,
		deps.GlobalConfig,
		grpcserver.WithExtraGRPCHandlers[T](
			consensusServer.GRPCServiceRegistrar(
				deps.ClientContext,
				deps.GlobalConfig,
			),
		),
	)
	if err != nil {
		return nil, err
	}

	grpcgatewayServer, err := grpcgateway.New[T](
		logger,
		deps.GlobalConfig,
		simApp.InterfaceRegistry(),
		simApp.App.AppManager,
	)
	if err != nil {
		return nil, err
	}
	registerGRPCGatewayRoutes(deps.ClientContext, grpcgatewayServer)

	// wire server commands
	return serverv2.AddCommands[T](
		rootCmd,
		logger,
		simApp,
		deps.GlobalConfig,
		initServerConfig(),
		consensusServer,
		grpcServer,
		storeComponent,
		telemetryServer,
		restServer,
		grpcgatewayServer,
		swaggerServer,
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

// registerGRPCGatewayRoutes registers the gRPC gateway routes for all modules and other components
func registerGRPCGatewayRoutes[T transaction.Tx](
	clientContext client.Context,
	server *grpcgateway.Server[T],
) {
	// those are the extra services that the CometBFT server implements (server/v2/cometbft/grpc.go)
	cmtservice.RegisterGRPCGatewayRoutes(clientContext, server.GRPCGatewayRouter)
	_ = nodeservice.RegisterServiceHandlerClient(context.Background(), server.GRPCGatewayRouter, nodeservice.NewServiceClient(clientContext))
	_ = txtypes.RegisterServiceHandlerClient(context.Background(), server.GRPCGatewayRouter, txtypes.NewServiceClient(clientContext))
}
