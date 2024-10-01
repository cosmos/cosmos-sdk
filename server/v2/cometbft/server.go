package cometbft

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	abciserver "github.com/cometbft/cometbft/abci/server"
	cmtcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	cometlog "cosmossdk.io/server/v2/cometbft/log"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/snapshots"

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
	Consensus *Consensus[T]

	initTxCodec   transaction.Codec[T]
	logger        log.Logger
	serverOptions ServerOptions[T]
	config        Config
	cfgOptions    []CfgOption
}

func New[T transaction.Tx](txCodec transaction.Codec[T], serverOptions ServerOptions[T], cfgOptions ...CfgOption) *CometBFTServer[T] {
	return &CometBFTServer[T]{
		initTxCodec:   txCodec,
		serverOptions: serverOptions,
		cfgOptions:    cfgOptions,
	}
}

func (s *CometBFTServer[T]) Init(appI serverv2.AppI[T], cfg map[string]any, logger log.Logger) error {
	home, _ := cfg[serverv2.FlagHome].(string)

	// get configs (app.toml + config.toml) from viper
	appTomlConfig := s.Config().(*AppTomlConfig)
	configTomlConfig := cmtcfg.DefaultConfig().SetRoot(home)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, s.Name(), &appTomlConfig); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if err := serverv2.UnmarshalSubConfig(cfg, "", &configTomlConfig); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	s.config = Config{
		ConfigTomlConfig: configTomlConfig,
		AppTomlConfig:    appTomlConfig,
	}

	chainID, _ := cfg[FlagChainID].(string)
	if chainID == "" {
		// fallback to genesis chain-id
		reader, err := os.Open(filepath.Join(home, "config", "genesis.json"))
		if err != nil {
			panic(err)
		}
		defer reader.Close()

		chainID, err = genutiltypes.ParseChainIDFromGenesis(reader)
		if err != nil {
			panic(fmt.Errorf("failed to parse chain-id from genesis file: %w", err))
		}
	}

	indexEvents := make(map[string]struct{}, len(s.config.AppTomlConfig.IndexEvents))
	for _, e := range s.config.AppTomlConfig.IndexEvents {
		indexEvents[e] = struct{}{}
	}

	s.logger = logger.With(log.ModuleKey, s.Name())
	store := appI.GetStore().(types.Store)
	consensus := NewConsensus(
		s.logger,
		appI.Name(),
		appI.GetAppManager(),
		s.serverOptions.Mempool(cfg),
		indexEvents,
		appI.GetGPRCMethodsToMessageMap(),
		store,
		s.config,
		s.initTxCodec,
		chainID,
	)
	consensus.prepareProposalHandler = s.serverOptions.PrepareProposalHandler
	consensus.processProposalHandler = s.serverOptions.ProcessProposalHandler
	consensus.verifyVoteExt = s.serverOptions.VerifyVoteExtensionHandler
	consensus.extendVote = s.serverOptions.ExtendVoteHandler
	consensus.addrPeerFilter = s.serverOptions.AddrPeerFilter
	consensus.idPeerFilter = s.serverOptions.IdPeerFilter

	ss := store.GetStateStorage().(snapshots.StorageSnapshotter)
	sc := store.GetStateCommitment().(snapshots.CommitSnapshotter)

	snapshotStore, err := GetSnapshotStore(s.config.ConfigTomlConfig.RootDir)
	if err != nil {
		return err
	}
	consensus.snapshotManager = snapshots.NewManager(snapshotStore, s.serverOptions.SnapshotOptions(cfg), sc, ss, nil, s.logger)

	s.Consensus = consensus

	return nil
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

	return s.Node.Start()
}

func (s *CometBFTServer[T]) Stop(context.Context) error {
	if s.Node != nil && s.Node.IsRunning() {
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
