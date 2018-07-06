package server

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/abci/server"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/node"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
)

const (
	flagWithTendermint = "with-tendermint"
	flagAddress        = "address"
)

// StartCmd runs the service passed in, either
// stand-alone, or in-process with tendermint
func StartCmd(ctx *sdk.ServerContext, appCreator AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !viper.GetBool(flagWithTendermint) {
				ctx.Logger.Info("Starting ABCI without Tendermint")
				return startStandAlone(ctx, appCreator)
			}
			ctx.Logger.Info("Starting ABCI with Tendermint")
			_, err := startInProcess(ctx, appCreator)
			return err
		},
	}

	// basic flags for abci app
	cmd.Flags().Bool(flagWithTendermint, true, "run abci app embedded in-process with tendermint")
	cmd.Flags().String(flagAddress, "tcp://0.0.0.0:26658", "Listen address")

	// AddNodeFlags adds support for all tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

func startStandAlone(ctx *sdk.ServerContext, appCreator AppCreator) error {
	// Generate the app in the proper dir
	addr := viper.GetString(flagAddress)
	home := viper.GetString("home")
	app, err := appCreator(home, ctx)
	if err != nil {
		return err
	}

	svr, err := server.NewServer(addr, "socket", app)
	if err != nil {
		return errors.Errorf("error creating listener: %v\n", err)
	}
	svr.SetLogger(ctx.Logger.With("module", "abci-server"))
	err = svr.Start()
	if err != nil {
		cmn.Exit(err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		err = svr.Stop()
		if err != nil {
			cmn.Exit(err.Error())
		}
	})
	return nil
}

func startInProcess(ctx *sdk.ServerContext, appCreator AppCreator) (*node.Node, error) {
	cfg := ctx.Config
	home := cfg.RootDir
	app, err := appCreator(home, ctx)
	if err != nil {
		return nil, err
	}

	// Create & start tendermint node
	n, err := node.NewNode(cfg,
		pvm.LoadOrGenFilePV(cfg.PrivValidatorFile()),
		proxy.NewLocalClientCreator(app),
		node.DefaultGenesisDocProviderFunc(cfg),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider,
		ctx.Logger.With("module", "node"))
	if err != nil {
		return nil, err
	}

	err = n.Start()
	if err != nil {
		return nil, err
	}

	// Trap signal, run forever.
	n.RunForever()
	return n, nil
}
