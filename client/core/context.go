package core

import (
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

type CoreContext struct {
	ChainID         string
	Height          int64
	TrustNode       bool
	NodeURI         string
	FromAddressName string
	Sequence        int64
	Client          rpcclient.Client
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

func (c CoreContext) WithClient(client rpcclient.Client) CoreContext {
	c.Client = client
	return c
}
