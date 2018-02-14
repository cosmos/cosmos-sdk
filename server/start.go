package server

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/abci/server"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

const (
	flagWithTendermint = "with-tendermint"
	flagAddress        = "address"
)

// StartCmd runs the service passed in, either
// stand-alone, or in-process with tendermint
func StartCmd(app *baseapp.BaseApp) *cobra.Command {
	start := startCmd{
		app: app,
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
	// do this in main:
	// rootDir := viper.GetString(cli.HomeFlag)
	// node.Logger = ....
	app *baseapp.BaseApp
}

func (s startCmd) run(cmd *cobra.Command, args []string) error {
	logger := s.app.Logger
	if !viper.GetBool(flagWithTendermint) {
		logger.Info("Starting ABCI without Tendermint")
		return s.startStandAlone()
	}
	logger.Info("Starting ABCI with Tendermint")
	return s.startInProcess()
}

func (s startCmd) startStandAlone() error {
	logger := s.app.Logger
	// Start the ABCI listener
	addr := viper.GetString(flagAddress)
	svr, err := server.NewServer(addr, "socket", s.app)
	if err != nil {
		return errors.Errorf("Error creating listener: %v\n", err)
	}
	svr.SetLogger(logger.With("module", "abci-server"))
	svr.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
	return nil
}

func (s startCmd) startInProcess() error {
	logger := s.app.Logger
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}

	// Create & start tendermint node
	n, err := node.NewNode(cfg,
		types.LoadOrGenPrivValidatorFS(cfg.PrivValidatorFile()),
		proxy.NewLocalClientCreator(s.app),
		node.DefaultGenesisDocProviderFunc(cfg),
		node.DefaultDBProvider,
		logger.With("module", "node"))
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
