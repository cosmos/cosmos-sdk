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

	corectx "cosmossdk.io/core/context"
	"cosmossdk.io/core/log"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	cometlog "cosmossdk.io/server/v2/cometbft/log"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/snapshots"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*CometBFTServer[transaction.Tx])(nil)
	_ serverv2.HasCLICommands                  = (*CometBFTServer[transaction.Tx])(nil)
	_ serverv2.HasStartFlags                   = (*CometBFTServer[transaction.Tx])(nil)
)

type CometBFTServer[T transaction.Tx] struct {
	Node      *node.Node
	Consensus *Consensus[T]

	appName, version string
	initTxCodec      transaction.Codec[T]
	logger           log.Logger
	config           Config
	options          ServerOptions[T]
	cfgOptions       []CfgOption
}

func New[T transaction.Tx](appName, version string, txCodec transaction.Codec[T], options ServerOptions[T], cfgOptions ...CfgOption) *CometBFTServer[T] {
	return &CometBFTServer[T]{
		appName:     appName,
		version:     version,
		initTxCodec: txCodec,
		options:     options,
		cfgOptions:  cfgOptions,
	}
}

func (s *CometBFTServer[T]) Init(appI serverv2.AppI[T], v *viper.Viper, logger log.Logger) error {
	s.config = Config{ConfigTomlConfig: GetConfigTomlFromViper(v), ConsensusAuthority: appI.GetConsensusAuthority()}
	s.logger = logger.With(log.ModuleKey, s.Name())

	// create consensus
	store := appI.GetStore().(types.Store)
	consensus := NewConsensus[T](s.appName, s.version, appI.GetAppManager(), s.options.Mempool, appI.GetGRPCQueryDecoders(), store, s.config, s.initTxCodec, s.logger)

	consensus.prepareProposalHandler = s.options.PrepareProposalHandler
	consensus.processProposalHandler = s.options.ProcessProposalHandler
	consensus.verifyVoteExt = s.options.VerifyVoteExtensionHandler
	consensus.extendVote = s.options.ExtendVoteHandler

	// TODO: set these; what is the appropriate presence of the Store interface here?
	var ss snapshots.StorageSnapshotter
	var sc snapshots.CommitSnapshotter

	snapshotStore, err := GetSnapshotStore(s.config.ConfigTomlConfig.RootDir)
	if err != nil {
		return err
	}

	sm := snapshots.NewManager(snapshotStore, s.options.SnapshotOptions, sc, ss, nil, s.logger)
	consensus.SetSnapshotManager(sm)

	s.Consensus = consensus
	return nil
}

func (s *CometBFTServer[T]) Name() string {
	return "comet"
}

func (s *CometBFTServer[T]) Start(ctx context.Context) error {
	viper := ctx.Value(corectx.ViperContextKey).(*viper.Viper)
	cometConfig := GetConfigTomlFromViper(viper)

	wrappedLogger := cometlog.CometLoggerWrapper{Logger: s.logger}
	if s.config.AppTomlConfig.Standalone {
		svr, err := abciserver.NewServer(s.config.AppTomlConfig.Address, s.config.AppTomlConfig.Transport, s.Consensus)
		if err != nil {
			return fmt.Errorf("error creating listener: %w", err)
		}

		svr.SetLogger(wrappedLogger)

		return svr.Start()
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(cometConfig.NodeKeyFile())
	if err != nil {
		return err
	}

	s.Node, err = node.NewNode(
		ctx,
		cometConfig,
		pvm.LoadOrGenFilePV(cometConfig.PrivValidatorKeyFile(), cometConfig.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewConsensusSyncLocalClientCreator(s.Consensus),
		getGenDocProvider(cometConfig),
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(cometConfig.Instrumentation),
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
	flags := pflag.NewFlagSet("cometbft", pflag.ExitOnError)
	// TODO: add flags to overwrite app.toml / config.toml values
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
		cfg := &Config{AppTomlConfig: DefaultConfig()}
		// overwrite the default config with the provided options
		for _, opt := range s.cfgOptions {
			opt(cfg)
		}

		return cfg.AppTomlConfig
	}

	return s.config.AppTomlConfig
}

// WriteDefaultConfigAt writes the default cometbft config.toml
func (s *CometBFTServer[T]) WriteDefaultConfigAt(configPath string) error {
	cfg := &Config{ConfigTomlConfig: cmtcfg.DefaultConfig()}
	for _, opt := range s.cfgOptions {
		opt(cfg)
	}

	cmtcfg.WriteConfigFile(filepath.Join(configPath, "config.toml"), cfg.ConfigTomlConfig)
	return nil
}
