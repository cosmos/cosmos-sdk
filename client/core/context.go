package core

import (
	// "fmt"

	// "github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
)

type CoreContext struct {
	ChainID         string
	Height          int64
	TrustNode       bool
	NodeURI         string
	FromAddressName string
	Sequence        int64
}

func NewCoreContextFromViper() CoreContext {
	return CoreContext{
		ChainID:         viper.GetString(client.FlagChainID),
		Height:          viper.GetInt64(client.FlagHeight),
		TrustNode:       viper.GetBool(client.FlagTrustNode),
		NodeURI:         viper.GetString(client.FlagNode),
		FromAddressName: viper.GetString(client.FlagName),
		Sequence:        viper.GetInt64(client.FlagSequence),
	}
}
