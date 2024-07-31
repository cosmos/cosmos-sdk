package cometbft

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	cometlog "cosmossdk.io/server/v2/cometbft/log"
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

func (s *CometBFTServer[T]) Init(appI serverv2.AppI[T], v *viper.Viper, logger log.Logger) error {
	// get configs (app.toml + config.toml) from viper
	appTomlConfig := s.Config().(*AppTomlConfig)
	if v != nil {
		if err := serverv2.UnmarshalSubConfig(v, s.Name(), &appTomlConfig); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	s.config = Config{
		ConfigTomlConfig: getConfigTomlFromViper(v),
		AppTomlConfig:    appTomlConfig,
	}

	indexEvents := make(map[string]struct{}, len(s.config.AppTomlConfig.IndexEvents))
	for _, e := range s.config.AppTomlConfig.IndexEvents {
		indexEvents[e] = struct{}{}
	}

	s.logger = logger.With(log.ModuleKey, s.Name())
	consensus := NewConsensus(
		s.logger,
		appI.Name(),
		appI.GetConsensusAuthority(),
		appI.GetAppManager(),
		s.serverOptions.Mempool,
		indexEvents,
		appI.GetGRPCQueryDecoders(),
		appI.GetStore().(types.Store),
		s.config,
		s.initTxCodec,
	)
	consensus.prepareProposalHandler = s.serverOptions.PrepareProposalHandler
	consensus.processProposalHandler = s.serverOptions.ProcessProposalHandler
	consensus.verifyVoteExt = s.serverOptions.VerifyVoteExtensionHandler
	consensus.extendVote = s.serverOptions.ExtendVoteHandler
	consensus.addrPeerFilter = s.serverOptions.AddrPeerFilter
	consensus.idPeerFilter = s.serverOptions.IdPeerFilter

	// TODO: set these; what is the appropriate presence of the Store interface here?
	var ss snapshots.StorageSnapshotter
	var sc snapshots.CommitSnapshotter

	snapshotStore, err := GetSnapshotStore(s.config.ConfigTomlConfig.RootDir)
	if err != nil {
		return err
	}
	consensus.snapshotManager = snapshots.NewManager(snapshotStore, s.serverOptions.SnapshotOptions, sc, ss, nil, s.logger)

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

	s.Node, err = node.NewNode(
		ctx,
		s.config.ConfigTomlConfig,
		pvm.LoadOrGenFilePV(s.config.ConfigTomlConfig.PrivValidatorKeyFile(), s.config.ConfigTomlConfig.PrivValidatorStateFile()),
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
			s.StatusCommand(),
			s.ShowNodeIDCmd(),
			s.ShowValidatorCmd(),
			s.ShowAddressCmd(),
			s.VersionCmd(),
			cmtcmd.ResetAllCmd,
			cmtcmd.ResetStateCmd,
		},
		Queries: []*cobra.Command{
			s.QueryBlockCmd(),
			s.QueryBlocksCmd(),
			s.QueryBlockResultsCmd(),
		},
	}
}

// CometBFT is a special server, it has config in config.toml and app.toml

// Config returns the (app.toml) server configuration.
func (s *CometBFTServer[T]) Config() any {
	if s.config.AppTomlConfig == nil || s.config.AppTomlConfig == (&AppTomlConfig{}) {
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
