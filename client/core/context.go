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

func (c CoreContext) WithChainID(chainID string) CoreContext {
	c.ChainID = chainID
	return c
}

func (c CoreContext) WithHeight(height int64) CoreContext {
	c.Height = height
	return c
}

func (c CoreContext) WithTrustNode(trustNode bool) CoreContext {
	c.TrustNode = trustNode
	return c
}

func (c CoreContext) WithNodeURI(nodeURI string) CoreContext {
	c.NodeURI = nodeURI
	return c
}

func (c CoreContext) WithFromAddressName(fromAddressName string) CoreContext {
	c.FromAddressName = fromAddressName
	return c
}

func (c CoreContext) WithSequence(sequence int64) CoreContext {
	c.Sequence = sequence
	return c
}
