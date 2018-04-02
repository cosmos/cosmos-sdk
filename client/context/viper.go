package context

import (
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/core"
)

func NewCoreContextFromViper() core.CoreContext {
	return core.CoreContext{
		ChainID:         viper.GetString(client.FlagChainID),
		Height:          viper.GetInt64(client.FlagHeight),
		TrustNode:       viper.GetBool(client.FlagTrustNode),
		NodeURI:         viper.GetString(client.FlagNode),
		FromAddressName: viper.GetString(client.FlagName),
		Sequence:        viper.GetInt64(client.FlagSequence),
	}
}
