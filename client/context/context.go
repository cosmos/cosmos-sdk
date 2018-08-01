package context

import (
	"io"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/spf13/viper"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

const ctxAccStoreName = "acc"

// QueryContext implements a typical context created in SDK modules for
// queries.
type QueryContext struct {
	Codec           *wire.Codec
	AccDecoder      auth.AccountDecoder
	Client          rpcclient.Client
	Logger          io.Writer
	Height          int64
	NodeURI         string
	FromAddressName string
	AccountStore    string
	TrustNode       bool
	UseLedger       bool
	Async           bool
	JSON            bool
	PrintResponse   bool
}

// NewQueryContextFromCLI returns a new initialized QueryContext with
// parameters from the command line using Viper.
func NewQueryContextFromCLI() QueryContext {
	var rpc rpcclient.Client

	nodeURI := viper.GetString(client.FlagNode)
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}

	return QueryContext{
		Client:          rpc,
		NodeURI:         nodeURI,
		AccountStore:    ctxAccStoreName,
		FromAddressName: viper.GetString(client.FlagFrom),
		Height:          viper.GetInt64(client.FlagHeight),
		TrustNode:       viper.GetBool(client.FlagTrustNode),
		UseLedger:       viper.GetBool(client.FlagUseLedger),
		Async:           viper.GetBool(client.FlagAsync),
		JSON:            viper.GetBool(client.FlagJson),
		PrintResponse:   viper.GetBool(client.FlagPrintResponse),
	}
}

// WithCodec returns a copy of the context with an updated codec.
func (ctx QueryContext) WithCodec(cdc *wire.Codec) QueryContext {
	ctx.Codec = cdc
	return ctx
}

// WithAccountDecoder returns a copy of the context with an updated account
// decoder.
func (ctx QueryContext) WithAccountDecoder(decoder auth.AccountDecoder) QueryContext {
	ctx.AccDecoder = decoder
	return ctx
}

// WithLogger returns a copy of the context with an updated logger.
func (ctx QueryContext) WithLogger(w io.Writer) QueryContext {
	ctx.Logger = w
	return ctx
}

// WithAccountStore returns a copy of the context with an updated AccountStore.
func (ctx QueryContext) WithAccountStore(accountStore string) QueryContext {
	ctx.AccountStore = accountStore
	return ctx
}

// WithFromAddressName returns a copy of the context with an updated from
// address.
func (ctx QueryContext) WithFromAddressName(addrName string) QueryContext {
	ctx.FromAddressName = addrName
	return ctx
}

// WithTrustNode returns a copy of the context with an updated TrustNode flag.
func (ctx QueryContext) WithTrustNode(trustNode bool) QueryContext {
	ctx.TrustNode = trustNode
	return ctx
}

// WithNodeURI returns a copy of the context with an updated node URI.
func (ctx QueryContext) WithNodeURI(nodeURI string) QueryContext {
	ctx.NodeURI = nodeURI
	ctx.Client = rpcclient.NewHTTP(nodeURI, "/websocket")
	return ctx
}

// WithClient returns a copy of the context with an updated RPC client
// instance.
func (ctx QueryContext) WithClient(client rpcclient.Client) QueryContext {
	ctx.Client = client
	return ctx
}

// WithUseLedger returns a copy of the context with an updated UseLedger flag.
func (ctx QueryContext) WithUseLedger(useLedger bool) QueryContext {
	ctx.UseLedger = useLedger
	return ctx
}
