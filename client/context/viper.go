package context

import (
	"github.com/spf13/viper"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/core"
)

func NewCoreContextFromViper() core.CoreContext {
	nodeURI := viper.GetString(client.FlagNode)
	var rpc rpcclient.Client
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}
	return core.CoreContext{
		ChainID:         viper.GetString(client.FlagChainID),
		Height:          viper.GetInt64(client.FlagHeight),
		TrustNode:       viper.GetBool(client.FlagTrustNode),
		FromAddressName: viper.GetString(client.FlagName),
		NodeURI:         nodeURI,
		Sequence:        viper.GetInt64(client.FlagSequence),
		Client:          rpc,
	}
}
