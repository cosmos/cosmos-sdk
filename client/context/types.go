package context

import (
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/x/auth"
)

// typical context created in sdk modules for transactions/queries
type CoreContext struct {
	ChainID         string
	Height          int64
	Gas             int64
	Fee             string
	TrustNode       bool
	NodeURI         string
	FromAddressNames []string
	AccountNumbers   []int64
	Sequences        []int64
	Memo            string
	Client          rpcclient.Client
	Decoder         auth.AccountDecoder
	AccountStore    string
	UseLedger       bool
	Async           bool
	JSON            bool
	PrintResponse   bool
}

// WithChainID - return a copy of the context with an updated chainID
func (c CoreContext) WithChainID(chainID string) CoreContext {
	c.ChainID = chainID
	return c
}

// WithHeight - return a copy of the context with an updated height
func (c CoreContext) WithHeight(height int64) CoreContext {
	c.Height = height
	return c
}

// WithGas - return a copy of the context with an updated gas
func (c CoreContext) WithGas(gas int64) CoreContext {
	c.Gas = gas
	return c
}

// WithFee - return a copy of the context with an updated fee
func (c CoreContext) WithFee(fee string) CoreContext {
	c.Fee = fee
	return c
}

// WithTrustNode - return a copy of the context with an updated TrustNode flag
func (c CoreContext) WithTrustNode(trustNode bool) CoreContext {
	c.TrustNode = trustNode
	return c
}

// WithNodeURI - return a copy of the context with an updated node URI
func (c CoreContext) WithNodeURI(nodeURI string) CoreContext {
	c.NodeURI = nodeURI
	c.Client = rpcclient.NewHTTP(nodeURI, "/websocket")
	return c
}

// WithFromAddressName - return a copy of the context with an updated from address
func (c CoreContext) WithFromAddressNames(fromAddressNames []string) CoreContext {
	c.FromAddressNames = fromAddressNames
	return c
}

// WithSequence - return a copy of the context with an account number
func (c CoreContext) WithAccountNumbers(accnums []int64) CoreContext {
	c.AccountNumbers = accnums
	return c
}

// WithSequence - return a copy of the context with an updated sequence number
func (c CoreContext) WithSequences(sequences []int64) CoreContext {
	c.Sequences = sequences
	return c
}

// WithMemo - return a copy of the context with an updated memo
func (c CoreContext) WithMemo(memo string) CoreContext {
	c.Memo = memo
	return c
}

// WithClient - return a copy of the context with an updated RPC client instance
func (c CoreContext) WithClient(client rpcclient.Client) CoreContext {
	c.Client = client
	return c
}

// WithDecoder - return a copy of the context with an updated Decoder
func (c CoreContext) WithDecoder(decoder auth.AccountDecoder) CoreContext {
	c.Decoder = decoder
	return c
}

// WithAccountStore - return a copy of the context with an updated AccountStore
func (c CoreContext) WithAccountStore(accountStore string) CoreContext {
	c.AccountStore = accountStore
	return c
}

// WithUseLedger - return a copy of the context with an updated UseLedger
func (c CoreContext) WithUseLedger(useLedger bool) CoreContext {
	c.UseLedger = useLedger
	return c
}
