package cometbft

import (
	"context"
	"fmt"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	cometlog "cosmossdk.io/server/v2/cometbft/log"
	"cosmossdk.io/server/v2/cometbft/types"
	abciserver "github.com/cometbft/cometbft/abci/server"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmttypes "github.com/cometbft/cometbft/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

type CometBFTServer struct {
	Node   *node.Node
	app    abci.Application
	logger log.Logger

	config    *cmtcfg.Config
	cleanupFn func()
}

func NewCometBFTServer(
	logger log.Logger,
	app *runtime.App,
	cfg *cmtcfg.Config,
	voteExtHandler types.VoteExtensionsHandler,
) *CometBFTServer {
	logger = logger.With("module", "cometbft-server")
	return &CometBFTServer{
		logger: logger,
		app:    NewConsensus[transaction.Tx](app.AppManager, nil, nil, cfg),
		config: cfg,
	}
}

func (s *CometBFTServer) Name() string {
	return "cometbft"
}

func (s *CometBFTServer) Start(ctx context.Context) error {
	wrappedLogger := cometlog.CometLoggerWrapper{Logger: s.logger}
	if s.config.Standalone {
		svr, err := abciserver.NewServer(s.config.Addr, s.config.Transport, s.app)
		if err != nil {
			return fmt.Errorf("error creating listener: %w", err)
		}

		svr.SetLogger(wrappedLogger)

		return svr.Start()
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(s.config.NodeKeyFile())
	if err != nil {
		return err
	}

	s.Node, err = node.NewNodeWithContext(
		ctx,
		s.config,
		pvm.LoadOrGenFilePV(s.config.PrivValidatorKeyFile(), s.config.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(s.app),
		getGenDocProvider(s.config),
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(s.config.Instrumentation),
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

func (s *CometBFTServer) Stop(_ context.Context) error {
	defer s.cleanupFn()
	if s.Node != nil {
		return s.Node.Stop()
	}
	return nil
}

// returns a function which returns the genesis doc from the genesis file.
func getGenDocProvider(cfg *cmtcfg.Config) func() (*cmttypes.GenesisDoc, error) {
	return func() (*cmttypes.GenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(cfg.GenesisFile())
		if err != nil {
			return nil, err
		}

		return appGenesis.ToGenesisDoc()
	}
}
