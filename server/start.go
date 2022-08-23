package server

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/abci/server"
	tcmd "github.com/tendermint/tendermint/cmd/cometbft/commands"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/rpc/client/local"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	crgserver "github.com/cosmos/cosmos-sdk/server/rosetta/lib/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Tendermint full-node start flags
	flagWithTendermint     = "with-tendermint"
	flagAddress            = "address"
	flagTransport          = "transport"
	flagTraceStore         = "trace-store"
	flagCPUProfile         = "cpu-profile"
	FlagMinGasPrices       = "minimum-gas-prices"
	FlagQueryGasLimit      = "query-gas-limit"
	FlagHaltHeight         = "halt-height"
	FlagHaltTime           = "halt-time"
	FlagInterBlockCache    = "inter-block-cache"
	FlagUnsafeSkipUpgrades = "unsafe-skip-upgrades"
	FlagTrace              = "trace"
	FlagInvCheckPeriod     = "inv-check-period"

	FlagPruning             = "pruning"
	FlagPruningKeepRecent   = "pruning-keep-recent"
	FlagPruningKeepEvery    = "pruning-keep-every"
	FlagPruningInterval     = "pruning-interval"
	FlagIndexEvents         = "index-events"
	FlagMinRetainBlocks     = "min-retain-blocks"
	FlagIAVLCacheSize       = "iavl-cache-size"
	FlagDisableIAVLFastNode = "iavl-disable-fastnode"

	// state sync-related flags
	FlagStateSyncSnapshotInterval   = "state-sync.snapshot-interval"
	FlagStateSyncSnapshotKeepRecent = "state-sync.snapshot-keep-recent"

	// gRPC-related flags
	flagGRPCOnly       = "grpc-only"
	flagGRPCEnable     = "grpc.enable"
	flagGRPCAddress    = "grpc.address"
	flagGRPCWebEnable  = "grpc-web.enable"
	flagGRPCWebAddress = "grpc-web.address"
)

// StartCmd runs the service passed in, either stand-alone or in-process with
// CometBFT.
func StartCmd[T types.Application](appCreator types.AppCreator[T]) *cobra.Command {
	return StartCmdWithOptions(appCreator, StartCmdOptions[T]{})
}

// StartCmdWithOptions runs the service passed in, either stand-alone or in-process with
// CometBFT.
func StartCmdWithOptions[T types.Application](appCreator types.AppCreator[T], opts StartCmdOptions[T]) *cobra.Command {
	if opts.DBOpener == nil {
		opts.DBOpener = OpenDB
	}

	if opts.StartCommandHandler == nil {
		opts.StartCommandHandler = start
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with CometBFT in or out of process. By
default, the application will run with CometBFT in process.

Pruning options can be provided via the '--pruning' flag or alternatively with '--pruning-keep-recent', and
'pruning-interval' together.

For '--pruning' the options are as follows:

default: the last 362880 states are kept, pruning at 10 block intervals
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: all saved states will be deleted, storing only the current and previous state; pruning at 10 block intervals
custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.

The node may be started in a 'query only' mode where only the gRPC and JSON HTTP
API services are enabled via the 'grpc-only' flag. In this mode, Tendermint is
bypassed and can be used when legacy queries are needed after an on-chain upgrade
is performed. Note, when enabled, gRPC will also be automatically enabled.
`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := GetServerContextFromCmd(cmd)

			// Bind flags to the Context's Viper so the app construction can set
			// options accordingly.
			serverCtx.Viper.BindPFlags(cmd.Flags())

The node may be started in a 'query only' mode where only the gRPC and JSON HTTP
API services are enabled via the 'grpc-only' flag. In this mode, CometBFT is
bypassed and can be used when legacy queries are needed after an on-chain upgrade
is performed. Note, when enabled, gRPC will also be automatically enabled.
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			_, err := GetPruningOptionsFromFlags(serverCtx.Viper)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			withCMT, _ := cmd.Flags().GetBool(flagWithComet)
			if !withCMT {
				serverCtx.Logger.Info("starting ABCI without CometBFT")
			}

			// amino is needed here for backwards compatibility of REST routes
			err = startInProcess(serverCtx, clientCtx, appCreator)
			errCode, ok := err.(ErrorCode)
			if !ok {
				return err
			}

			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().Bool(flagWithTendermint, true, "Run abci app embedded in-process with tendermint")
	cmd.Flags().String(flagAddress, "tcp://0.0.0.0:26658", "Listen address")
	cmd.Flags().String(flagTransport, "socket", "Transport protocol: socket, grpc")
	cmd.Flags().String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)")
	cmd.Flags().IntSlice(FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(flagCPUProfile, "", "Enable CPU profiling and write to the provided file")
	cmd.Flags().Bool(FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().String(FlagPruning, storetypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningKeepEvery, 0, "Offset heights to keep on disk after 'keep-every' (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint(FlagInvCheckPeriod, 0, "Assert registered invariants every N blocks")
	cmd.Flags().Uint64(FlagMinRetainBlocks, 0, "Minimum block height offset during ABCI commit to prune Tendermint blocks")

	cmd.Flags().Bool(flagGRPCOnly, false, "Start the node in gRPC query only mode (no Tendermint process is started)")
	cmd.Flags().Bool(flagGRPCEnable, true, "Define if the gRPC server should be enabled")
	cmd.Flags().String(flagGRPCAddress, config.DefaultGRPCAddress, "the gRPC server address to listen on")

	cmd.Flags().Bool(flagGRPCWebEnable, true, "Define if the gRPC-Web server should be enabled. (Note: gRPC must also be enabled.)")
	cmd.Flags().String(flagGRPCWebAddress, config.DefaultGRPCWebAddress, "The gRPC-Web server address to listen on")

	cmd.Flags().Uint64(FlagStateSyncSnapshotInterval, 0, "State sync snapshot interval")
	cmd.Flags().Uint32(FlagStateSyncSnapshotKeepRecent, 2, "State sync snapshot to keep")

	cmd.Flags().Bool(FlagDisableIAVLFastNode, true, "Disable fast node for IAVL tree")

	// add support for all Tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

func start[T types.Application](svrCtx *Context, clientCtx client.Context, appCreator types.AppCreator[T], withCmt bool, opts StartCmdOptions[T]) error {
	svrCfg, err := getAndValidateConfig(svrCtx)
	if err != nil {
		return err
	}

	app, appCleanupFn, err := startApp[T](svrCtx, appCreator, opts)
	if err != nil {
		return err
	}
	defer appCleanupFn()

	metrics, err := startTelemetry(svrCfg)
	if err != nil {
		return err
	}

	app := appCreator(ctx.Logger, db, traceWriter, ctx.Viper)

	config, err := serverconfig.GetConfig(ctx.Viper)
	if err != nil {
		return err
	}

	_, err = startTelemetry(config)
	if err != nil {
		return err
	}

	if !withCmt {
		return startStandAlone[T](svrCtx, svrCfg, clientCtx, app, metrics, opts)
	}
	return startInProcess[T](svrCtx, svrCfg, clientCtx, app, metrics, opts)
}

func startStandAlone[T types.Application](svrCtx *Context, svrCfg serverconfig.Config, clientCtx client.Context, app T, metrics *telemetry.Metrics, opts StartCmdOptions[T]) error {
	addr := svrCtx.Viper.GetString(flagAddress)
	transport := svrCtx.Viper.GetString(flagTransport)

	cmtApp := NewCometABCIWrapper(app)
	svr, err := server.NewServer(addr, transport, cmtApp)
	if err != nil {
		return fmt.Errorf("error creating listener: %w", err)
	}

	svr.SetLogger(servercmtlog.CometLoggerWrapper{Logger: svrCtx.Logger.With("module", "abci-server")})

	g, ctx := getCtx(svrCtx, false)

	// Add the tx service to the gRPC router. We only need to register this
	// service if API or gRPC is enabled, and avoid doing so in the general
	// case, because it spawns a new local CometBFT RPC client.
	if svrCfg.API.Enable || svrCfg.GRPC.Enable {
		// create tendermint client
		// assumes the rpc listen address is where tendermint has its rpc server
		rpcclient, err := rpchttp.New(svrCtx.Config.RPC.ListenAddress)
		if err != nil {
			return err
		}
		// re-assign for making the client available below
		// do not use := to avoid shadowing clientCtx
		clientCtx = clientCtx.WithClient(rpcclient)

		// use the provided clientCtx to register the services
		app.RegisterTxService(clientCtx)
		app.RegisterTendermintService(clientCtx)
		app.RegisterNodeService(clientCtx, svrCfg)
	}

func startInProcess(ctx *Context, clientCtx client.Context, appCreator types.AppCreator) error {
	cfg := ctx.Config
	home := cfg.RootDir
	var cpuProfileCleanup func()

	err = startAPIServer(ctx, g, svrCfg, clientCtx, svrCtx, app, svrCtx.Config.RootDir, grpcSrv, metrics)
	if err != nil {
		return err
	}

	if opts.PostSetupStandalone != nil {
		if err := opts.PostSetupStandalone(app, svrCtx, clientCtx, ctx, g); err != nil {
			return err
		}
	}

	g.Go(func() error {
		if err := svr.Start(); err != nil {
			svrCtx.Logger.Error("failed to start out-of-process ABCI server", "err", err)
			return err
		}

		// Wait for the calling process to be canceled or close the provided context,
		// so we can gracefully stop the ABCI server.
		<-ctx.Done()
		svrCtx.Logger.Info("stopping the ABCI server...")
		return svr.Stop()
	})

	return g.Wait()
}

func startInProcess[T types.Application](svrCtx *Context, svrCfg serverconfig.Config, clientCtx client.Context, app T,
	metrics *telemetry.Metrics, opts StartCmdOptions[T],
) error {
	cmtCfg := svrCtx.Config
	gRPCOnly := svrCtx.Viper.GetBool(flagGRPCOnly)

	g, ctx := getCtx(svrCtx, true)

	if gRPCOnly {
		// TODO: Generalize logic so that gRPC only is really in startStandAlone
		svrCtx.Logger.Info("starting node in gRPC only mode; CometBFT is disabled")
		svrCfg.GRPC.Enable = true
	} else {
		svrCtx.Logger.Info("starting node with ABCI CometBFT in-process")
		tmNode, cleanupFn, err := startCmtNode(ctx, cmtCfg, app, svrCtx)
		if err != nil {
			return err
		}
		defer cleanupFn()

		// Add the tx service to the gRPC router. We only need to register this
		// service if API or gRPC is enabled, and avoid doing so in the general
		// case, because it spawns a new local CometBFT RPC client.
		if svrCfg.API.Enable || svrCfg.GRPC.Enable {
			// Re-assign for making the client available below do not use := to avoid
			// shadowing the clientCtx variable.
			clientCtx = clientCtx.WithClient(local.New(tmNode))

			app.RegisterTxService(clientCtx)
			app.RegisterTendermintService(clientCtx)
			app.RegisterNodeService(clientCtx, svrCfg)
		}
	}

	grpcSrv, clientCtx, err := startGrpcServer(ctx, g, svrCfg.GRPC, clientCtx, svrCtx, app)
	if err != nil {
		return err
	}

	err = startAPIServer(ctx, g, svrCfg, clientCtx, svrCtx, app, cmtCfg.RootDir, grpcSrv, metrics)
	if err != nil {
		return err
	}

	config, err := serverconfig.GetConfig(ctx.Viper)
	if err != nil {
		return err
	}

	if err := config.ValidateBasic(); err != nil {
		return err
	}

	if err := config.ValidateBasic(); err != nil {
		ctx.Logger.Error("WARNING: The minimum-gas-prices config in app.toml is set to the empty string. " +
			"This defaults to 0 in the current version, but will error in the next version " +
			"(SDK v0.45). Please explicitly put the desired minimum-gas-prices in your app.toml.")
	}

	// wait for signal capture and gracefully return
	// we are guaranteed to be waiting for the "ListenForQuitSignals" goroutine.
	return g.Wait()
}

// TODO: Move nodeKey into being created within the function.
func startCmtNode(
	ctx context.Context,
	cfg *cmtcfg.Config,
	app types.Application,
	svrCtx *Context,
) (tmNode *node.Node, cleanupFn func(), err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, cleanupFn, err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, cleanupFn, err
	}

	genDocProvider := node.DefaultGenesisDocProviderFunc(cfg)

	var (
		tmNode   *node.Node
		gRPCOnly = ctx.Viper.GetBool(flagGRPCOnly)
	)

	if gRPCOnly {
		ctx.Logger.Info("starting node in gRPC only mode; Tendermint is disabled")
		config.GRPC.Enable = true
	} else {
		ctx.Logger.Info("starting node with ABCI Tendermint in-process")

		tmNode, err = node.NewNode(
			cfg,
			pvm.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile()),
			nodeKey,
			proxy.NewLocalClientCreator(app),
			genDocProvider,
			node.DefaultDBProvider,
			node.DefaultMetricsProvider(cfg.Instrumentation),
			ctx.Logger,
		)
		if err != nil {
			return err
		}
		if err := tmNode.Start(); err != nil {
			return err
		}
	}

	// Add the tx service to the gRPC router. We only need to register this
	// service if API or gRPC is enabled, and avoid doing so in the general
	// case, because it spawns a new local tendermint RPC client.
	if (config.API.Enable || config.GRPC.Enable) && tmNode != nil {
		clientCtx = clientCtx.WithClient(local.New(tmNode))

		app.RegisterTxService(clientCtx)
		app.RegisterTendermintService(clientCtx)

		if a, ok := app.(types.ApplicationQueryService); ok {
			a.RegisterNodeService(clientCtx)
		}
	}
	return config, nil
}

	metrics, err := startTelemetry(config)
	if err != nil {
		return err
	}

	var apiSrv *api.Server
	if config.API.Enable {
		genDoc, err := genDocProvider()
		if err != nil {
			return defaultGenesisDoc, err
		}

		clientCtx := clientCtx.WithHomeDir(home).WithChainID(genDoc.ChainID)

		genbz, err := gen.AppState.MarshalJSON()
		if err != nil {
			return defaultGenesisDoc, err
		}

		apiSrv = api.New(clientCtx, ctx.Logger.With("module", "api-server"))
		app.RegisterAPIRoutes(apiSrv, config.API)
		if config.Telemetry.Enabled {
			apiSrv.SetTelemetry(metrics)
		}
		errCh := make(chan error)

		go func() {
			if err := apiSrv.Start(config); err != nil {
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			return err

		case <-time.After(types.ServerStartTime): // assume server started successfully
		}
	}

	return traceWriter, cleanup, nil
}

func startGrpcServer(
	ctx context.Context,
	g *errgroup.Group,
	config serverconfig.GRPCConfig,
	clientCtx client.Context,
	svrCtx *Context,
	app types.Application,
) (*grpc.Server, client.Context, error) {
	if !config.Enable {
		// return grpcServer as nil if gRPC is disabled
		return nil, clientCtx, nil
	}
	_, _, err := net.SplitHostPort(config.Address)
	if err != nil {
		return nil, clientCtx, err
	}

	maxSendMsgSize := config.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = serverconfig.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := config.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = serverconfig.DefaultGRPCMaxRecvMsgSize
	}

	// if gRPC is enabled, configure gRPC client for gRPC gateway
	grpcClient, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize),
		),
	)

	if config.GRPC.Enable {
		grpcSrv, err = servergrpc.StartGRPCServer(clientCtx, app, config.GRPC.Address)
		if err != nil {
			return err
		}

		if config.GRPCWeb.Enable {
			grpcWebSrv, err = servergrpc.StartGRPCWeb(grpcSrv, config)
			if err != nil {
				ctx.Logger.Error("failed to start grpc-web http server: ", err)
				return err
			}
		}

		defer func() {
			svrCtx.Logger.Info("stopping CPU profiler", "profile", cpuProfile)
			pprof.StopCPUProfile()

			if err := f.Close(); err != nil {
				svrCtx.Logger.Info("failed to close cpu-profile file", "profile", cpuProfile, "err", err.Error())
			}
		}()
	}

	// At this point it is safe to block the process if we're in gRPC only mode as
	// we do not need to start Rosetta or handle any Tendermint related processes.
	if gRPCOnly {
		// wait for signal capture and gracefully return
		return WaitForQuitSignals()
	}

	var rosettaSrv crgserver.Server
	if config.Rosetta.Enable {
		offlineMode := config.Rosetta.Offline

		// If GRPC is not enabled rosetta cannot work in online mode, so it works in
		// offline mode.
		if !config.GRPC.Enable {
			offlineMode = true
		}

		conf := &rosetta.Config{
			Blockchain:        config.Rosetta.Blockchain,
			Network:           config.Rosetta.Network,
			TendermintRPC:     ctx.Config.RPC.ListenAddress,
			GRPCEndpoint:      config.GRPC.Address,
			Addr:              config.Rosetta.Address,
			Retries:           config.Rosetta.Retries,
			Offline:           offlineMode,
			Codec:             clientCtx.Codec.(*codec.ProtoCodec),
			InterfaceRegistry: clientCtx.InterfaceRegistry,
		}

	telemetry.SetGaugeWithLabels([]string{"server", "info"}, 1, ls)
}

func getCtx(svrCtx *Context, block bool) (*errgroup.Group, context.Context) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	ListenForQuitSignals(g, block, cancelFn, svrCtx.Logger)
	return g, ctx
}

func startApp[T types.Application](svrCtx *Context, appCreator types.AppCreator[T], opts StartCmdOptions[T]) (app T, cleanupFn func(), err error) {
	traceWriter, traceCleanupFn, err := SetupTraceWriter(svrCtx.Logger, svrCtx.Viper.GetString(flagTraceStore))
	if err != nil {
		return app, traceCleanupFn, err
	}

	home := svrCtx.Config.RootDir
	db, err := opts.DBOpener(home, GetAppDBBackend(svrCtx.Viper))
	if err != nil {
		return app, traceCleanupFn, err
	}

	if isTestnet, ok := svrCtx.Viper.Get(KeyIsTestnet).(bool); ok && isTestnet {
		var appPtr *T
		appPtr, err = testnetify[T](svrCtx, appCreator, db, traceWriter)
		if err != nil {
			return app, traceCleanupFn, err
		}

		errCh := make(chan error)
		go func() {
			if err := rosettaSrv.Start(); err != nil {
				errCh <- err
			}

			return err

		case <-time.After(types.ServerStartTime): // assume server started successfully
		}
	}

	addStartNodeFlags(cmd, opts)
	cmd.Flags().String(KeyTriggerTestnetUpgrade, "", "If set (example: \"v21\"), triggers the v21 upgrade handler to run on the first block of the testnet")
	cmd.Flags().Bool("skip-confirmation", false, "Skip the confirmation prompt")
	return cmd
}

// testnetify modifies both state and blockStore, allowing the provided operator address and local validator key to control the network
// that the state in the data folder represents. The chainID of the local genesis file is modified to match the provided chainID.
func testnetify[T types.Application](ctx *Context, testnetAppCreator types.AppCreator[T], db corestore.KVStoreWithBatch, traceWriter io.WriteCloser) (*T, error) {
	config := ctx.Config

	newChainID, ok := ctx.Viper.Get(KeyNewChainID).(string)
	if !ok {
		return nil, fmt.Errorf("expected string for key %s", KeyNewChainID)
	}

	// Modify app genesis chain ID and save to genesis file.
	genFilePath := config.GenesisFile()
	appGen, err := genutiltypes.AppGenesisFromFile(genFilePath)
	if err != nil {
		return nil, err
	}
	appGen.ChainID = newChainID
	if err := appGen.ValidateAndComplete(); err != nil {
		return nil, err
	}
	if err := appGen.SaveAs(genFilePath); err != nil {
		return nil, err
	}

	// Load the comet genesis doc provider.
	genDocProvider := node.DefaultGenesisDocProviderFunc(config)

	// Initialize blockStore and stateDB.
	blockStoreDB, err := cmtcfg.DefaultDBProvider(&cmtcfg.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return nil, err
	}
	blockStore := store.NewBlockStore(blockStoreDB)

	stateDB, err := cmtcfg.DefaultDBProvider(&cmtcfg.DBContext{ID: "state", Config: config})
	if err != nil {
		return nil, err
	}

	defer blockStore.Close()
	defer stateDB.Close()

	privValidator, err := pvm.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), func() (cmtcrypto.PrivKey, error) {
		return cmted25519.GenPrivKey(), nil
	}) // TODO: make this modular
	if err != nil {
		return nil, err
	}
	userPubKey, err := privValidator.GetPubKey()
	if err != nil {
		return nil, err
	}
	validatorAddress := userPubKey.Address()

	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: config.Storage.DiscardABCIResponses,
	})

	state, genDoc, err := node.LoadStateFromDBOrGenesisDocProvider(stateDB, genDocProvider, "")
	if err != nil {
		return nil, err
	}

	ctx.Viper.Set(KeyNewValAddr, validatorAddress)
	ctx.Viper.Set(KeyUserPubKey, userPubKey)
	testnetApp := testnetAppCreator(ctx.Logger, db, traceWriter, ctx.Viper)

	// We need to create a temporary proxyApp to get the initial state of the application.
	// Depending on how the node was stopped, the application height can differ from the blockStore height.
	// This height difference changes how we go about modifying the state.
	cmtApp := NewCometABCIWrapper(testnetApp)
	_, context := getCtx(ctx, true)
	clientCreator := proxy.NewLocalClientCreator(cmtApp)
	metrics := node.DefaultMetricsProvider(cmtcfg.DefaultConfig().Instrumentation)
	_, _, _, _, _, proxyMetrics, _, _ := metrics(genDoc.ChainID) //nolint: dogsled // function from comet
	proxyApp := proxy.NewAppConns(clientCreator, proxyMetrics)
	if err := proxyApp.Start(); err != nil {
		return nil, fmt.Errorf("error starting proxy app connections: %w", err)
	}
	res, err := proxyApp.Query().Info(context, proxy.InfoRequest)
	if err != nil {
		return nil, fmt.Errorf("error calling Info: %w", err)
	}
	err = proxyApp.Stop()
	if err != nil {
		return nil, err
	}
	appHash := res.LastBlockAppHash
	appHeight := res.LastBlockHeight

	var block *cmttypes.Block
	switch {
	case appHeight == blockStore.Height():
		block, _ = blockStore.LoadBlock(blockStore.Height())
		// If the state's last blockstore height does not match the app and blockstore height, we likely stopped with the halt height flag.
		if state.LastBlockHeight != appHeight {
			state.LastBlockHeight = appHeight
			block.AppHash = appHash
			state.AppHash = appHash
		} else {
			// Node was likely stopped via SIGTERM, delete the next block's seen commit
			err := blockStoreDB.Delete([]byte(fmt.Sprintf("SC:%v", blockStore.Height()+1)))
			if err != nil {
				return nil, err
			}
		}
	case blockStore.Height() > state.LastBlockHeight:
		// This state usually occurs when we gracefully stop the node.
		err = blockStore.DeleteLatestBlock()
		if err != nil {
			return nil, err
		}
		block, _ = blockStore.LoadBlock(blockStore.Height())
	default:
		// If there is any other state, we just load the block
		block, _ = blockStore.LoadBlock(blockStore.Height())
	}

	block.ChainID = newChainID
	state.ChainID = newChainID

	block.LastBlockID = state.LastBlockID
	block.LastCommit.BlockID = state.LastBlockID

	// Create a vote from our validator
	vote := cmttypes.Vote{
		Type:             cmtproto.PrecommitType,
		Height:           state.LastBlockHeight,
		Round:            0,
		BlockID:          state.LastBlockID,
		Timestamp:        time.Now(),
		ValidatorAddress: validatorAddress,
		ValidatorIndex:   0,
		Signature:        []byte{},
	}

	// Sign the vote, and copy the proto changes from the act of signing to the vote itself
	voteProto := vote.ToProto()
	err = privValidator.SignVote(newChainID, voteProto, false)
	if err != nil {
		return nil, err
	}
	vote.Signature = voteProto.Signature
	vote.Timestamp = voteProto.Timestamp

	// Modify the block's lastCommit to be signed only by our validator
	block.LastCommit.Signatures[0].ValidatorAddress = validatorAddress
	block.LastCommit.Signatures[0].Signature = vote.Signature
	block.LastCommit.Signatures = []cmttypes.CommitSig{block.LastCommit.Signatures[0]}

	// Load the seenCommit of the lastBlockHeight and modify it to be signed from our validator
	seenCommit := blockStore.LoadSeenCommit(state.LastBlockHeight)
	seenCommit.BlockID = state.LastBlockID
	seenCommit.Round = vote.Round
	seenCommit.Signatures[0].Signature = vote.Signature
	seenCommit.Signatures[0].ValidatorAddress = validatorAddress
	seenCommit.Signatures[0].Timestamp = vote.Timestamp
	seenCommit.Signatures = []cmttypes.CommitSig{seenCommit.Signatures[0]}
	err = blockStore.SaveSeenCommit(state.LastBlockHeight, seenCommit)
	if err != nil {
		return nil, err
	}

	// Create ValidatorSet struct containing just our validator.
	newVal := &cmttypes.Validator{
		Address:     validatorAddress,
		PubKey:      userPubKey,
		VotingPower: 900000000000000,
	}
	newValSet := &cmttypes.ValidatorSet{
		Validators: []*cmttypes.Validator{newVal},
		Proposer:   newVal,
	}

	// Replace all valSets in state to be the valSet with just our validator.
	state.Validators = newValSet
	state.LastValidators = newValSet
	state.NextValidators = newValSet
	state.LastHeightValidatorsChanged = blockStore.Height()

	err = stateStore.Save(state)
	if err != nil {
		return nil, err
	}

	// Create a ValidatorsInfo struct to store in stateDB.
	valSet, err := state.Validators.ToProto()
	if err != nil {
		return nil, err
	}
	valInfo := &cmtstate.ValidatorsInfo{
		ValidatorSet:      valSet,
		LastHeightChanged: state.LastBlockHeight,
	}
	buf, err := valInfo.Marshal()
	if err != nil {
		return nil, err
	}

	// Modify Validators stateDB entry.
	err = stateDB.Set([]byte(fmt.Sprintf("validatorsKey:%v", blockStore.Height())), buf)
	if err != nil {
		return nil, err
	}

	// Modify LastValidators stateDB entry.
	err = stateDB.Set([]byte(fmt.Sprintf("validatorsKey:%v", blockStore.Height()-1)), buf)
	if err != nil {
		return nil, err
	}

	// Modify NextValidators stateDB entry.
	err = stateDB.Set([]byte(fmt.Sprintf("validatorsKey:%v", blockStore.Height()+1)), buf)
	if err != nil {
		return nil, err
	}

	// Since we modified the chainID, we set the new genesisDoc in the stateDB.
	b, err := cmtjson.Marshal(genDoc)
	if err != nil {
		return nil, err
	}
	if err := stateDB.SetSync([]byte("genesisDoc"), b); err != nil {
		return nil, err
	}

	return &testnetApp, err
}

// addStartNodeFlags should be added to any CLI commands that start the network.
func addStartNodeFlags[T types.Application](cmd *cobra.Command, opts StartCmdOptions[T]) {
	cmd.Flags().Bool(flagWithComet, true, "Run abci app embedded in-process with CometBFT")
	cmd.Flags().String(flagAddress, "tcp://127.0.0.1:26658", "Listen address")
	cmd.Flags().String(flagTransport, "socket", "Transport protocol: socket, grpc")
	cmd.Flags().String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)")
	cmd.Flags().Uint64(FlagQueryGasLimit, 0, "Maximum gas a Rest/Grpc query can consume. Blank and 0 imply unbounded.")
	cmd.Flags().IntSlice(FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(flagCPUProfile, "", "Enable CPU profiling and write to the provided file")
	cmd.Flags().Bool(FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().String(FlagPruning, pruningtypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint(FlagInvCheckPeriod, 0, "Assert registered invariants every N blocks")
	cmd.Flags().Uint64(FlagMinRetainBlocks, 0, "Minimum block height offset during ABCI commit to prune CometBFT blocks")
	cmd.Flags().Bool(FlagAPIEnable, false, "Define if the API server should be enabled")
	cmd.Flags().Bool(FlagAPISwagger, false, "Define if swagger documentation should automatically be registered (Note: the API must also be enabled)")
	cmd.Flags().String(FlagAPIAddress, serverconfig.DefaultAPIAddress, "the API server address to listen on")
	cmd.Flags().Uint(FlagAPIMaxOpenConnections, 1000, "Define the number of maximum open connections")
	cmd.Flags().Uint(FlagRPCReadTimeout, 10, "Define the CometBFT RPC read timeout (in seconds)")
	cmd.Flags().Uint(FlagRPCWriteTimeout, 0, "Define the CometBFT RPC write timeout (in seconds)")
	cmd.Flags().Uint(FlagRPCMaxBodyBytes, 1000000, "Define the CometBFT maximum request body (in bytes)")
	cmd.Flags().Bool(FlagAPIEnableUnsafeCORS, false, "Define if CORS should be enabled (unsafe - use it at your own risk)")
	cmd.Flags().Bool(flagGRPCOnly, false, "Start the node in gRPC query only mode (no CometBFT process is started)")
	cmd.Flags().Bool(flagGRPCEnable, true, "Define if the gRPC server should be enabled")
	cmd.Flags().String(flagGRPCAddress, serverconfig.DefaultGRPCAddress, "the gRPC server address to listen on")
	cmd.Flags().Uint64(FlagStateSyncSnapshotInterval, 0, "State sync snapshot interval")
	cmd.Flags().Uint32(FlagStateSyncSnapshotKeepRecent, 2, "State sync snapshot to keep")
	cmd.Flags().Bool(FlagDisableIAVLFastNode, false, "Disable fast node for IAVL tree")
	cmd.Flags().Int(FlagMempoolMaxTxs, mempool.DefaultMaxTx, "Sets MaxTx value for the app-side mempool")
	cmd.Flags().Duration(FlagShutdownGrace, 0*time.Second, "On Shutdown, duration to wait for resource clean up")

	// support old flags name for backwards compatibility
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == "with-tendermint" {
			name = flagWithComet
		}

		return pflag.NormalizedName(name)
	})

	// wait for signal capture and gracefully return
	return WaitForQuitSignals()
}

func startTelemetry(cfg config.Config) (*telemetry.Metrics, error) {
	if !cfg.Telemetry.Enabled {
		return nil, nil
	}
	return telemetry.New(cfg.Telemetry)
}
