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
	_ serverv2.ServerComponent[
		serverv2.AppI[transaction.Tx], transaction.Tx,
	] = (*CometBFTServer[serverv2.AppI[transaction.Tx], transaction.Tx])(nil)
	_ serverv2.HasCLICommands = (*CometBFTServer[serverv2.AppI[transaction.Tx], transaction.Tx])(nil)
	_ serverv2.HasStartFlags  = (*CometBFTServer[serverv2.AppI[transaction.Tx], transaction.Tx])(nil)
)

type CometBFTServer[AppT serverv2.AppI[T], T transaction.Tx] struct {
	Node      *node.Node
	Consensus Consensus[T]

	initTxCodec transaction.Codec[T]
	Logger      log.Logger
	Config      Config
	Options     ServerOptions[T]
}

func New[AppT serverv2.AppI[T], T transaction.Tx](txCodec transaction.Codec[T], options ServerOptions[T]) *CometBFTServer[AppT, T] {
	return &CometBFTServer[AppT, T]{
		initTxCodec: txCodec,
		Options:     options,
	}
}

func (s *CometBFTServer[AppT, T]) Init(appI AppT, v *viper.Viper, logger log.Logger) error {
	s.Config = Config{CmtConfig: GetConfigFromViper(v), ConsensusAuthority: appI.GetConsensusAuthority()}
	s.Logger = logger.With(log.ModuleKey, s.Name())

	// create consensus
	store := appI.GetStore().(types.Store)
	consensus := NewConsensus[T](appI.GetAppManager(), s.Options.Mempool, store, s.Config, s.initTxCodec, s.Logger)

	consensus.prepareProposalHandler = s.Options.PrepareProposalHandler
	consensus.processProposalHandler = s.Options.ProcessProposalHandler
	consensus.verifyVoteExt = s.Options.VerifyVoteExtensionHandler
	consensus.extendVote = s.Options.ExtendVoteHandler

	// TODO: set these; what is the appropriate presence of the Store interface here?
	var ss snapshots.StorageSnapshotter
	var sc snapshots.CommitSnapshotter

	snapshotStore, err := GetSnapshotStore(s.Config.CmtConfig.RootDir)
	if err != nil {
		return err
	}

	sm := snapshots.NewManager(snapshotStore, s.Options.SnapshotOptions, sc, ss, nil, s.Logger)
	consensus.SetSnapshotManager(sm)

	s.Consensus = consensus
	return nil
}

func (s *CometBFTServer[AppT, T]) Name() string {
	return "cometbft-server"
}

func (s *CometBFTServer[AppT, T]) Start(ctx context.Context) error {
	viper := ctx.Value(corectx.ViperContextKey).(*viper.Viper)
	cometConfig := GetConfigFromViper(viper)

	wrappedLogger := cometlog.CometLoggerWrapper{Logger: s.Logger}
	if s.Config.Standalone {
		svr, err := abciserver.NewServer(s.Config.Addr, s.Config.Transport, s.Consensus)
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

func (s *CometBFTServer[AppT, T]) Stop(context.Context) error {
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

func (s *CometBFTServer[AppT, T]) StartCmdFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("cometbft", pflag.ExitOnError)
	flags.Bool(FlagWithComet, true, "Run abci app embedded in-process with CometBFT")
	flags.String(FlagAddress, "tcp://127.0.0.1:26658", "Listen address")
	flags.String(FlagTransport, "socket", "Transport protocol: socket, grpc")
	flags.String(FlagTraceStore, "", "Enable KVStore tracing to an output file")
	flags.String(FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)")
	flags.Uint64(FlagQueryGasLimit, 0, "Maximum gas a Rest/Grpc query can consume. Blank and 0 imply unbounded.")
	flags.Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	flags.Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	flags.String(FlagCPUProfile, "", "Enable CPU profiling and write to the provided file")
	flags.Bool(FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	return flags
}

func (s *CometBFTServer[AppT, T]) CLICommands() serverv2.CLIConfig {
	return serverv2.CLIConfig{
		Commands: []*cobra.Command{
			s.StatusCommand(),
			s.ShowNodeIDCmd(),
			s.ShowValidatorCmd(),
			s.ShowAddressCmd(),
			s.VersionCmd(),
			s.QueryBlockCmd(),
			s.QueryBlocksCmd(),
			s.QueryBlockResultsCmd(),
			cmtcmd.ResetAllCmd,
			cmtcmd.ResetStateCmd,
		},
	}
}

func (s *CometBFTServer[AppT, T]) WriteDefaultConfigAt(configPath string) error {
	cometConfig := cmtcfg.DefaultConfig()
	cmtcfg.WriteConfigFile(filepath.Join(configPath, "config.toml"), cometConfig)
	return nil
}
