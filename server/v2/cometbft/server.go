package cometbft

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	abciserver "github.com/cometbft/cometbft/abci/server"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	addresscodec "cosmossdk.io/core/address"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/schema/indexer"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/appmanager"
	cometlog "cosmossdk.io/server/v2/cometbft/log"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/oe"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/snapshots"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const ServerName = "comet"

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*CometBFTServer[transaction.Tx])(nil)
	_ serverv2.HasCLICommands                  = (*CometBFTServer[transaction.Tx])(nil)
	_ serverv2.HasStartFlags                   = (*CometBFTServer[transaction.Tx])(nil)
)

type CometBFTServer[T transaction.Tx] struct {
	Node      *node.Node
	Consensus abci.Application

	logger        log.Logger
	serverOptions ServerOptions[T]
	config        Config
	cfgOptions    []CfgOption

	app     appmanager.AppManager[T]
	txCodec transaction.Codec[T]
	store   types.Store
}

// AppCodecs contains all codecs that the CometBFT server requires
// provided by the application. They are extracted in struct to not be API
// breaking once amino is completely deprecated or new codecs should be added.
type AppCodecs[T transaction.Tx] struct {
	TxCodec transaction.Codec[T]

	// The following codecs are only required for the gRPC services
	AppCodec              codec.Codec
	LegacyAmino           registry.AminoRegistrar
	ConsensusAddressCodec addresscodec.Codec
}

func New[T transaction.Tx](
	logger log.Logger,
	appName string,
	store types.Store,
	app appmanager.AppManager[T],
	appCodecs AppCodecs[T],
	queryHandlers map[string]appmodulev2.Handler,
	decoderResolver decoding.DecoderResolver,
	serverOptions ServerOptions[T],
	cfg server.ConfigMap,
	cfgOptions ...CfgOption,
) (*CometBFTServer[T], error) {
	srv := &CometBFTServer[T]{
		serverOptions: serverOptions,
		cfgOptions:    cfgOptions,
		app:           app,
		txCodec:       appCodecs.TxCodec,
		store:         store,
	}
	srv.logger = logger.With(log.ModuleKey, srv.Name())

	home, _ := cfg[serverv2.FlagHome].(string)

	// get configs (app.toml + config.toml) from viper
	appTomlConfig := srv.Config().(*AppTomlConfig)
	configTomlConfig := cmtcfg.DefaultConfig().SetRoot(home)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, srv.Name(), &appTomlConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if err := serverv2.UnmarshalSubConfig(cfg, "", &configTomlConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	srv.config = Config{
		ConfigTomlConfig: configTomlConfig,
		AppTomlConfig:    appTomlConfig,
	}

	chainID, _ := cfg[FlagChainID].(string)
	if chainID == "" {
		// fallback to genesis chain-id
		reader, err := os.Open(srv.config.ConfigTomlConfig.GenesisFile())
		if err != nil {
			return nil, fmt.Errorf("failed to open genesis file: %w", err)
		}
		defer reader.Close()

		chainID, err = genutiltypes.ParseChainIDFromGenesis(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chain-id from genesis file: %w", err)
		}
	}

	indexedABCIEvents := make(map[string]struct{}, len(srv.config.AppTomlConfig.IndexABCIEvents))
	for _, e := range srv.config.AppTomlConfig.IndexABCIEvents {
		indexedABCIEvents[e] = struct{}{}
	}

	sc := store.GetStateCommitment().(snapshots.CommitSnapshotter)

	snapshotStore, err := GetSnapshotStore(srv.config.ConfigTomlConfig.RootDir)
	if err != nil {
		return nil, err
	}

	// initialize the indexer
	var listener *appdata.Listener
	if indexerCfg := srv.config.AppTomlConfig.Indexer; len(indexerCfg.Target) > 0 {
		indexingTarget, err := indexer.StartIndexing(indexer.IndexingOptions{
			Config:       indexerCfg,
			Resolver:     decoderResolver,
			Logger:       logger.With(log.ModuleKey, "indexer"),
			SyncSource:   nil, // TODO: Support catch-up syncs
			AddressCodec: appCodecs.AppCodec.InterfaceRegistry().SigningContext().AddressCodec(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to start indexing: %w", err)
		}

		listener = &indexingTarget.Listener
	}

	// snapshot manager
	snapshotManager := snapshots.NewManager(
		snapshotStore,
		srv.serverOptions.SnapshotOptions(cfg),
		sc,
		nil, // extensions snapshotter registered below
		logger,
	)
	if exts := serverOptions.SnapshotExtensions; len(exts) > 0 {
		if err := snapshotManager.RegisterExtensions(serverOptions.SnapshotExtensions...); err != nil {
			return nil, fmt.Errorf("failed to register snapshot extensions: %w", err)
		}
	}

	c := &consensus[T]{
		appName:                appName,
		version:                getCometBFTServerVersion(),
		app:                    app,
		cfg:                    srv.config,
		store:                  store,
		logger:                 logger,
		appCodecs:              appCodecs,
		listener:               listener,
		snapshotManager:        snapshotManager,
		streamingManager:       srv.serverOptions.StreamingManager,
		mempool:                srv.serverOptions.Mempool(cfg),
		lastCommittedHeight:    atomic.Int64{},
		prepareProposalHandler: srv.serverOptions.PrepareProposalHandler,
		processProposalHandler: srv.serverOptions.ProcessProposalHandler,
		verifyVoteExt:          srv.serverOptions.VerifyVoteExtensionHandler,
		checkTxHandler:         srv.serverOptions.CheckTxHandler,
		extendVote:             srv.serverOptions.ExtendVoteHandler,
		chainID:                chainID,
		indexedABCIEvents:      indexedABCIEvents,
		initialHeight:          0,
		queryHandlersMap:       queryHandlers,
		getProtoRegistry:       sync.OnceValues(gogoproto.MergedRegistry),
		addrPeerFilter:         srv.serverOptions.AddrPeerFilter,
		idPeerFilter:           srv.serverOptions.IdPeerFilter,
		cfgMap:                 cfg,
	}

	c.optimisticExec = oe.NewOptimisticExecution(
		logger,
		c.internalFinalizeBlock,
	)

	srv.Consensus = c

	return srv, nil
}

// NewWithConfigOptions creates a new CometBFT server with the provided config options.
// It is *not* a fully functional server (since it has been created without dependencies)
// The returned server should only be used to get and set configuration.
func NewWithConfigOptions[T transaction.Tx](opts ...CfgOption) *CometBFTServer[T] {
	return &CometBFTServer[T]{
		cfgOptions: opts,
	}
}

func (s *CometBFTServer[T]) Name() string {
	return ServerName
}

func (s *CometBFTServer[T]) Start(ctx context.Context) error {
	wrappedLogger := cometlog.CometLoggerWrapper{Logger: s.logger}
	if s.config.AppTomlConfig.Standalone {
		svr, err := abciserver.NewServer(s.config.AppTomlConfig.Address, s.config.AppTomlConfig.Transport, s.Consensus)
		if err != nil {
			return fmt.Errorf("error creating listener: %w", err)
		}

		svr.SetLogger(wrappedLogger)

		return svr.Start()
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(s.config.ConfigTomlConfig.NodeKeyFile())
	if err != nil {
		return err
	}

	pv, err := pvm.LoadOrGenFilePV(
		s.config.ConfigTomlConfig.PrivValidatorKeyFile(),
		s.config.ConfigTomlConfig.PrivValidatorStateFile(),
		s.serverOptions.KeygenF,
	)
	if err != nil {
		return err
	}

	s.Node, err = node.NewNode(
		ctx,
		s.config.ConfigTomlConfig,
		pv,
		nodeKey,
		proxy.NewConsensusSyncLocalClientCreator(s.Consensus),
		getGenDocProvider(s.config.ConfigTomlConfig),
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(s.config.ConfigTomlConfig.Instrumentation),
		wrappedLogger,
	)
	if err != nil {
		return err
	}

	s.logger.Info("starting consensus server")
	return s.Node.Start()
}

func (s *CometBFTServer[T]) Stop(context.Context) error {
	if s.Node != nil && s.Node.IsRunning() {
		s.logger.Info("stopping consensus server")
		return s.Node.Stop()
	}

	return nil
}

// returns a function which returns the genesis doc from the genesis file.
func getGenDocProvider(cfg *cmtcfg.Config) func() (node.ChecksummedGenesisDoc, error) {
	return func() (node.ChecksummedGenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(cfg.GenesisFile())
		if err != nil {
			return node.ChecksummedGenesisDoc{
				Sha256Checksum: []byte{},
			}, err
		}

		gen, err := appGenesis.ToGenesisDoc()
		if err != nil {
			return node.ChecksummedGenesisDoc{
				Sha256Checksum: []byte{},
			}, err
		}
		genbz, err := gen.AppState.MarshalJSON()
		if err != nil {
			return node.ChecksummedGenesisDoc{
				Sha256Checksum: []byte{},
			}, err
		}

		bz, err := json.Marshal(genbz)
		if err != nil {
			return node.ChecksummedGenesisDoc{
				Sha256Checksum: []byte{},
			}, err
		}
		sum := sha256.Sum256(bz)

		return node.ChecksummedGenesisDoc{
			GenesisDoc:     gen,
			Sha256Checksum: sum[:],
		}, nil
	}
}

func (s *CometBFTServer[T]) StartCmdFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(s.Name(), pflag.ExitOnError)

	flags.String(FlagAddress, "tcp://127.0.0.1:26658", "Listen address")
	flags.String(FlagTransport, "socket", "Transport protocol: socket, grpc")
	flags.Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	flags.Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	flags.Bool(FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	flags.Bool(Standalone, false, "Run app without CometBFT")
	flags.Int(FlagMempoolMaxTxs, mempool.DefaultMaxTx, "Sets MaxTx value for the app-side mempool")

	// add comet flags, we use an empty command to avoid duplicating CometBFT's AddNodeFlags.
	// we can then merge the flag sets.
	emptyCmd := &cobra.Command{}
	cmtcmd.AddNodeFlags(emptyCmd)
	flags.AddFlagSet(emptyCmd.Flags())

	return flags
}

func (s *CometBFTServer[T]) CLICommands() serverv2.CLIConfig {
	return serverv2.CLIConfig{
		Commands: []*cobra.Command{
			StatusCommand(),
			ShowNodeIDCmd(),
			ShowValidatorCmd(),
			ShowAddressCmd(),
			VersionCmd(),
			s.BootstrapStateCmd(),
			cmtcmd.ResetAllCmd,
			cmtcmd.ResetStateCmd,
		},
		Queries: []*cobra.Command{
			QueryBlockCmd(),
			QueryBlocksCmd(),
			QueryBlockResultsCmd(),
		},
	}
}

// CometBFT is a special server, it has config in config.toml and app.toml

// Config returns the (app.toml) server configuration.
func (s *CometBFTServer[T]) Config() any {
	if s.config.AppTomlConfig == nil || s.config.AppTomlConfig.Address == "" {
		cfg := &Config{AppTomlConfig: DefaultAppTomlConfig()}
		// overwrite the default config with the provided options
		for _, opt := range s.cfgOptions {
			opt(cfg)
		}

		return cfg.AppTomlConfig
	}

	return s.config.AppTomlConfig
}

// WriteCustomConfigAt writes the default cometbft config.toml
func (s *CometBFTServer[T]) WriteCustomConfigAt(configPath string) error {
	cfg := &Config{ConfigTomlConfig: cmtcfg.DefaultConfig()}
	for _, opt := range s.cfgOptions {
		opt(cfg)
	}

	cmtcfg.WriteConfigFile(filepath.Join(configPath, "config.toml"), cfg.ConfigTomlConfig)
	return nil
}

// gRPCServiceRegistrar returns a function that registers the CometBFT gRPC service
// Those services are defined for backward compatibility.
// Eventually, they will be removed in favor of the new gRPC services.
func (s *CometBFTServer[T]) GRPCServiceRegistrar(
	clientCtx client.Context,
	cfg server.ConfigMap,
) func(srv *grpc.Server) error {
	return gRPCServiceRegistrar[T](clientCtx, cfg, s.Config().(*AppTomlConfig), s.txCodec, s.Consensus, s.app)
}
