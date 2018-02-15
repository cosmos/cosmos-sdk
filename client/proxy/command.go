package proxy

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// TODO: unify these with general client commands,
// maybe reogranize location
const (
	flagListenAddr = "laddr"
	flagNode       = "node"
	flagChainID    = "chain-id"
)

// ProxyCmd starts a transparent proxy for tendermint rpc, verifying with
// lite-client proofs
func ProxyCmd(logger log.Logger) *cobra.Command {
	cmd := proxyCmd{logger}
	proxy := &cobra.Command{
		Use:   "proxy",
		Short: "Run lite-client proxy server, verifying tendermint rpc",
		Long: `This node will run a secure proxy to a tendermint rpc server.

All calls that can be tracked back to a block header by a proof
will be verified before passing them back to the caller. Other that
that it will present the same interface as a full tendermint node,
just with added trust and running locally.`,
		RunE:         cmd.run,
		SilenceUsage: true,
	}
	proxy.Flags().String(flagListenAddr, ":8888", "Serve the proxy on the given port")
	proxy.Flags().String(flagNode, "localhost:46657", "Connect to a Tendermint node at this address")
	proxy.Flags().String(flagChainID, "tendermint", "Specify the Tendermint chain ID")
	return proxy
}

type proxyCmd struct {
	logger log.Logger
}

func (p proxyCmd) run(cmd *cobra.Command, args []string) error {
	// TODO: these two are generic for all client commands
	home := viper.GetString("home")
	chainID := viper.GetString(flagChainID)
	// This as well possibly
	nodeAddr := viper.GetString(flagNode)
	// This is the only specific to proxy command
	listenAddr := viper.GetString(flagListenAddr)

	// First, connect a client
	node := rpcclient.NewHTTP(nodeAddr, "/websocket")

	cert, err := GetCertifier(chainID, home, nodeAddr)
	if err != nil {
		return err
	}
	sc := SecureClient(node, cert)

	err = StartProxy(sc, listenAddr, p.logger)
	if err != nil {
		return err
	}

	cmn.TrapSignal(func() {
		// TODO: close up shop
	})

	return nil
}
