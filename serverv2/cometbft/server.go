package cometbft

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	abciserver "github.com/cometbft/cometbft/abci/server"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	cometservice "github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmttypes "github.com/cometbft/cometbft/types"
	cometlog "github.com/cosmos/cosmos-sdk/serverv2/cometbft/log"
	"github.com/cosmos/cosmos-sdk/serverv2/cometbft/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

type Config struct {
	Transport  string
	Addr       string
	Standalone bool

	CmtConfig *cmtcfg.Config
}

type CometBFTServer struct {
	Node   *node.Node
	app    abci.Application
	logger log.Logger

	config    Config
	service   cometservice.Service
	cleanupFn func()
}

func NewCometBFTServer(logger log.Logger, app types.ProtoApp, cfg Config) *CometBFTServer {
	logger = logger.With("module", "cometbft-server")

	return &CometBFTServer{
		logger: logger,
		app:    NewCometABCIWrapper(app, logger),
		config: cfg,
	}
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

	nodeKey, err := p2p.LoadOrGenNodeKey(s.config.CmtConfig.NodeKeyFile())
	if err != nil {
		return err
	}

	s.Node, err = node.NewNodeWithContext(
		ctx,
		s.config.CmtConfig,
		pvm.LoadOrGenFilePV(s.config.CmtConfig.PrivValidatorKeyFile(), s.config.CmtConfig.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(s.app),
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

func (s *CometBFTServer) Stop() error {
	defer s.cleanupFn()
	if s.service != nil {
		return s.service.Stop()
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
