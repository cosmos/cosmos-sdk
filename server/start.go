package server

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/abci/server"
	abci "github.com/tendermint/abci/types"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"
)

const (
	flagWithTendermint = "with-tendermint"
	flagAddress        = "address"
)

// appGenerator lets us lazily initialize app, using home dir
// and other flags (?) to start
type appCreator func(string, log.Logger) (abci.Application, error)

// StartCmd runs the service passed in, either
// stand-alone, or in-process with tendermint
func StartCmd(app appCreator, logger log.Logger) *cobra.Command {
	start := startCmd{
		appCreator: app,
		logger:     logger,
	}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		RunE:  start.run,
	}
	// basic flags for abci app
	cmd.Flags().Bool(flagWithTendermint, true, "run abci app embedded in-process with tendermint")
	cmd.Flags().String(flagAddress, "tcp://0.0.0.0:46658", "Listen address")

	// AddNodeFlags adds support for all
	// tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

type startCmd struct {
	appCreator appCreator
	logger     log.Logger
}

func (s startCmd) run(cmd *cobra.Command, args []string) error {
	if !viper.GetBool(flagWithTendermint) {
		s.logger.Info("Starting ABCI without Tendermint")
		return s.startStandAlone()
	}
	s.logger.Info("Starting ABCI with Tendermint")
	return s.startInProcess()
}

func (s startCmd) startStandAlone() error {
	// Generate the app in the proper dir
	addr := viper.GetString(flagAddress)
	home := viper.GetString("home")
	app, err := s.appCreator(home, s.logger)
	if err != nil {
		return err
	}

	svr, err := server.NewServer(addr, "socket", app)
	if err != nil {
		return errors.Errorf("Error creating listener: %v\n", err)
	}
	svr.SetLogger(s.logger.With("module", "abci-server"))
	svr.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
	return nil
}

func (s startCmd) startInProcess() error {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}

	home := cfg.RootDir
	app, err := s.appCreator(home, s.logger)
	if err != nil {
		return err
	}

	// Create & start tendermint node
	n, err := node.NewNode(cfg,
		types.LoadOrGenPrivValidatorFS(cfg.PrivValidatorFile()),
		proxy.NewLocalClientCreator(app),
		node.DefaultGenesisDocProviderFunc(cfg),
		node.DefaultDBProvider,
		s.logger.With("module", "node"))
	if err != nil {
		return err
	}

	err = n.Start()
	if err != nil {
		return err
	}

	// Trap signal, run forever.
	n.RunForever()
	return nil
}
