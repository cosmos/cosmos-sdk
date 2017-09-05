package proxy

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Run proxy server, verifying tendermint rpc",
	Long: `This node will run a secure proxy to a tendermint rpc server.

All calls that can be tracked back to a block header by a proof
will be verified before passing them back to the caller. Other that
that it will present the same interface as a full tendermint node,
just with added trust and running locally.`,
	RunE:         commands.RequireInit(runProxy),
	SilenceUsage: true,
}

const (
	bindFlag = "serve"
)

func init() {
	RootCmd.Flags().String(bindFlag, ":8888", "Serve the proxy on the given port")
}

// TODO: pass in a proper logger
var logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))

func init() {
	logger = logger.With("module", "main")
	logger = log.NewFilter(logger, log.AllowInfo())
}

func runProxy(cmd *cobra.Command, args []string) error {
	// First, connect a client
	node := commands.GetNode()
	bind := viper.GetString(bindFlag)
	cert, err := commands.GetCertifier()
	if err != nil {
		return err
	}
	sc := client.SecureClient(node, cert)

	err = client.StartProxy(sc, bind, logger)
	if err != nil {
		return err
	}

	cmn.TrapSignal(func() {
		// TODO: close up shop
	})

	return nil
}
