package proxy

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// ProxyCmd starts a transparent proxy for tendermint rpc, verifying with
// lite-client proofs
var ProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Run lite-client proxy server, verifying tendermint rpc",
	Long: `This node will run a secure proxy to a tendermint rpc server.

All calls that can be tracked back to a block header by a proof
will be verified before passing them back to the caller. Other that
that it will present the same interface as a full tendermint node,
just with added trust and running locally.`,
	RunE:         runProxy,
	SilenceUsage: true,
}

const (
	flagListenAddr = "laddr"
	flagNode       = "node"
	flagChainID    = "chain-id"
)

func init() {
	LiteCmd.Flags().String(flagListenAddr, ":8888", "Serve the proxy on the given port")
	LiteCmd.Flags().String(flagNode, "localhost:46657", "Connect to a Tendermint node at this address")
	LiteCmd.Flags().String(flagChainID, "tendermint", "Specify the Tendermint chain ID")
}

func runProxy(cmd *cobra.Command, args []string) error {
	// TODO: these two are generic for all client commands
	home := viper.GetString("home")
	chainID := viper.GetString(flagChainID)
	// This as well possibly
	nodeAddr := viper.GetString(flagNode)
	// This is the only specific to proxy command
	laddr := viper.GetString(flagListenAddr)

	// First, connect a client
	node := rpcclient.NewHTTP(nodeAddr, "/websocket")

	cert, err := proxy.GetCertifier(chainID, home, nodeAddr)
	if err != nil {
		return err
	}
	sc := proxy.SecureClient(node, cert)

	err = proxy.StartProxy(sc, listenAddr, logger)
	if err != nil {
		return err
	}

	cmn.TrapSignal(func() {
		// TODO: close up shop
	})

	return nil
}
