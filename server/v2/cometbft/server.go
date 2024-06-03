package cometbft

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	abciserver "github.com/cometbft/cometbft/abci/server"
	cmtcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cosmossdk.io/core/log"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/handlers"
	cometlog "cosmossdk.io/server/v2/cometbft/log"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/snapshots"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const (
	flagWithComet     = "with-comet"
	flagAddress       = "address"
	flagTransport     = "transport"
	flagTraceStore    = "trace-store"
	flagCPUProfile    = "cpu-profile"
	FlagMinGasPrices  = "minimum-gas-prices"
	FlagQueryGasLimit = "query-gas-limit"
	FlagHaltHeight    = "halt-height"
	FlagHaltTime      = "halt-time"
	FlagTrace         = "trace"
)

var _ serverv2.ServerModule = (*CometBFTServer[transaction.Tx])(nil)

type CometBFTServer[T transaction.Tx] struct {
	Node   *node.Node
	App    *Consensus[T]
	logger log.Logger

	config    Config
	cleanupFn func()
}

// App is an interface that represents an application in the CometBFT server.
// It provides methods to access the app manager, logger, and store.
type App[T transaction.Tx] interface {
	GetApp() *appmanager.AppManager[T]
	GetLogger() log.Logger
	GetStore() types.Store
}

func NewCometBFTServer[T transaction.Tx](
	app *appmanager.AppManager[T],
	store types.Store,
	logger log.Logger,
	cfg Config,
	txCodec transaction.Codec[T],
) *CometBFTServer[T] {
	logger = logger.With("module", "cometbft-server")

	// create noop mempool
	mempool := mempool.NoOpMempool[T]{}

	// create consensus
	consensus := NewConsensus[T](app, mempool, store, cfg, txCodec, logger)

	consensus.SetPrepareProposalHandler(handlers.NoOpPrepareProposal[T]())
	consensus.SetProcessProposalHandler(handlers.NoOpProcessProposal[T]())
	consensus.SetVerifyVoteExtension(handlers.NoOpVerifyVoteExtensionHandler())
	consensus.SetExtendVoteExtension(handlers.NoOpExtendVote())

	// TODO: set these; what is the appropriate presence of the Store interface here?
	var ss snapshots.StorageSnapshotter
	var sc snapshots.CommitSnapshotter

	snapshotStore, err := GetSnapshotStore(cfg.CmtConfig.RootDir)
	if err != nil {
		panic(err)
	}

	sm := snapshots.NewManager(snapshotStore, snapshots.SnapshotOptions{}, sc, ss, nil, logger) // TODO: set options somehow
	consensus.SetSnapshotManager(sm)

	return &CometBFTServer[T]{
		logger: logger,
		App:    consensus,
		config: cfg,
	}
}

func (s *CometBFTServer[T]) Name() string {
	return "cometbft"
}

func (s *CometBFTServer[T]) Start(ctx context.Context) error {
	wrappedLogger := cometlog.CometLoggerWrapper{Logger: s.logger}
	if s.config.Standalone {
		svr, err := abciserver.NewServer(s.config.Addr, s.config.Transport, s.App)
		if err != nil {
			return fmt.Errorf("error creating listener: %w", err)
		}

		svr.SetLogger(wrappedLogger)

		return svr.Start()
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(s.config.CmtConfig.NodeKeyFile())
	if err != nil {
		return err
	}

	s.Node, err = node.NewNode(
		ctx,
		s.config.CmtConfig,
		pvm.LoadOrGenFilePV(s.config.CmtConfig.PrivValidatorKeyFile(), s.config.CmtConfig.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(s.App),
		getGenDocProvider(s.config.CmtConfig),
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(s.config.CmtConfig.Instrumentation),
		wrappedLogger,
	)
	if err != nil {
		return err
	}

	s.cleanupFn = func() {
		if s.Node != nil && s.Node.IsRunning() {
			_ = s.Node.Stop()
		}
	}

	return s.Node.Start()
}

func (s *CometBFTServer[T]) Stop(_ context.Context) error {
	defer s.cleanupFn()
	if s.Node != nil {
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

func (s *CometBFTServer[T]) StartCmdFlags() pflag.FlagSet {
	flags := *pflag.NewFlagSet("cometbft", pflag.ExitOnError)
	flags.Bool(flagWithComet, true, "Run abci app embedded in-process with CometBFT")
	flags.String(flagAddress, "tcp://127.0.0.1:26658", "Listen address")
	flags.String(flagTransport, "socket", "Transport protocol: socket, grpc")
	flags.String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	flags.String(FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)")
	flags.Uint64(FlagQueryGasLimit, 0, "Maximum gas a Rest/Grpc query can consume. Blank and 0 imply unbounded.")
	flags.Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	flags.Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	flags.String(flagCPUProfile, "", "Enable CPU profiling and write to the provided file")
	flags.Bool(FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
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
			s.QueryBlockCmd(),
			s.QueryBlocksCmd(),
			s.QueryBlockResultsCmd(),
			cmtcmd.ResetAllCmd,
			cmtcmd.ResetStateCmd,
		},
	}
}
