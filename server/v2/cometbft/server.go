package cometbft

// import (
// 	"context"
// 	"fmt"

// 	"cosmossdk.io/log"
// 	"cosmossdk.io/server/v2/appmanager"
// 	cometlog "cosmossdk.io/server/v2/cometbft/log"
// 	"cosmossdk.io/server/v2/cometbft/types"
// 	"cosmossdk.io/server/v2/core/transaction"
// 	"cosmossdk.io/store/v2/snapshots"
// 	abciserver "github.com/cometbft/cometbft/abci/server"
// 	abci "github.com/cometbft/cometbft/abci/types"
// 	cmtcfg "github.com/cometbft/cometbft/config"
// 	"github.com/cometbft/cometbft/node"
// 	"github.com/cometbft/cometbft/p2p"
// 	pvm "github.com/cometbft/cometbft/privval"
// 	"github.com/cometbft/cometbft/proxy"
// 	// genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
// )

// type CometBFTServer struct {
// 	Node   *node.Node
// 	app    abci.Application
// 	logger log.Logger

// 	config    Config
// 	cleanupFn func()
// }

// func NewCometBFTServer[T transaction.Tx](logger log.Logger, app appmanager.AppManager[T], cfg Config, voteExtHandler types.VoteExtensionsHandler) *CometBFTServer {
// 	logger = logger.With("module", "cometbft-server")
// 	return &CometBFTServer{
// 		logger: logger,
// 		app:    NewConsensus[T](app, nil, nil, cfg),
// 		config: cfg,
// 	}
// }

// func (s *CometBFTServer) Start(ctx context.Context) error {
// 	wrappedLogger := cometlog.CometLoggerWrapper{Logger: s.logger}
// 	if s.config.Standalone {
// 		svr, err := abciserver.NewServer(s.config.Addr, s.config.Transport, s.app)
// 		if err != nil {
// 			return fmt.Errorf("error creating listener: %w", err)
// 		}

// 		svr.SetLogger(wrappedLogger)

// 		return svr.Start()
// 	}

// 	nodeKey, err := p2p.LoadOrGenNodeKey(s.config.CmtConfig.NodeKeyFile())
// 	if err != nil {
// 		return err
// 	}

// 	s.Node, err = node.NewNode(
// 		ctx,
// 		s.config.CmtConfig,
// 		pvm.LoadOrGenFilePV(s.config.CmtConfig.PrivValidatorKeyFile(), s.config.CmtConfig.PrivValidatorStateFile()),
// 		nodeKey,
// 		proxy.NewLocalClientCreator(s.app),
// 		getGenDocProvider(s.config.CmtConfig),
// 		cmtcfg.DefaultDBProvider,
// 		node.DefaultMetricsProvider(s.config.CmtConfig.Instrumentation),
// 		wrappedLogger,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	s.cleanupFn = func() {
// 		if s.Node != nil && s.Node.IsRunning() {
// 			_ = s.Node.Stop()
// 		}
// 	}

// 	return s.Node.Start()
// }

// func (s *CometBFTServer) Stop() error {
// 	defer s.cleanupFn()
// 	if s.Node != nil {
// 		return s.Node.Stop()
// 	}
// 	return nil
// }

// // returns a function which returns the genesis doc from the genesis file.
// func getGenDocProvider(cfg *cmtcfg.Config) func() (node.ChecksummedGenesisDoc, error) {
// 	return func() (node.ChecksummedGenesisDoc, error) {
// 		// TODO: re-add this after fixing deps
// 		// appGenesis, err := genutiltypes.AppGenesisFromFile(cfg.GenesisFile())
// 		// if err != nil {
// 		// 	return nil, err
// 		// }

// 		// return appGenesis.ToGenesisDoc()
// 		return node.ChecksummedGenesisDoc{}, nil
// 	}
// }
