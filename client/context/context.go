package context

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/cli"
	tmlite "github.com/tendermint/tendermint/lite"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

const ctxAccStoreName = "acc"

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	Codec           *codec.Codec
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
	Certifier       tmlite.Certifier
	DryRun          bool
	GenerateOnly    bool
}

// NewCLIContext returns a new initialized CLIContext with parameters from the
// command line using Viper.
func NewCLIContext() CLIContext {
	var rpc rpcclient.Client

	nodeURI := viper.GetString(client.FlagNode)
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}

	return CLIContext{
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
		Certifier:       createCertifier(),
		DryRun:          viper.GetBool(client.FlagDryRun),
		GenerateOnly:    viper.GetBool(client.FlagGenerateOnly),
	}
}

func createCertifier() tmlite.Certifier {
	trustNode := viper.GetBool(client.FlagTrustNode)
	if trustNode {
		return nil
	}
	chainID := viper.GetString(client.FlagChainID)
	home := viper.GetString(cli.HomeFlag)
	nodeURI := viper.GetString(client.FlagNode)

	var errMsg bytes.Buffer
	if chainID == "" {
		errMsg.WriteString("chain-id ")
	}
	if home == "" {
		errMsg.WriteString("home ")
	}
	if nodeURI == "" {
		errMsg.WriteString("node ")
	}
	// errMsg is not empty
	if errMsg.Len() != 0 {
		panic(fmt.Errorf("can't create certifier for distrust mode, empty values from these options: %s", errMsg.String()))
	}
	certifier, err := tmliteProxy.GetCertifier(chainID, home, nodeURI)
	if err != nil {
		panic(err)
	}
	return certifier
}

// WithCodec returns a copy of the context with an updated codec.
func (ctx CLIContext) WithCodec(cdc *codec.Codec) CLIContext {
	ctx.Codec = cdc
	return ctx
}

// WithAccountDecoder returns a copy of the context with an updated account
// decoder.
func (ctx CLIContext) WithAccountDecoder(decoder auth.AccountDecoder) CLIContext {
	ctx.AccDecoder = decoder
	return ctx
}

// WithLogger returns a copy of the context with an updated logger.
func (ctx CLIContext) WithLogger(w io.Writer) CLIContext {
	ctx.Logger = w
	return ctx
}

// WithAccountStore returns a copy of the context with an updated AccountStore.
func (ctx CLIContext) WithAccountStore(accountStore string) CLIContext {
	ctx.AccountStore = accountStore
	return ctx
}

// WithFromAddressName returns a copy of the context with an updated from
// address.
func (ctx CLIContext) WithFromAddressName(addrName string) CLIContext {
	ctx.FromAddressName = addrName
	return ctx
}

// WithTrustNode returns a copy of the context with an updated TrustNode flag.
func (ctx CLIContext) WithTrustNode(trustNode bool) CLIContext {
	ctx.TrustNode = trustNode
	return ctx
}

// WithNodeURI returns a copy of the context with an updated node URI.
func (ctx CLIContext) WithNodeURI(nodeURI string) CLIContext {
	ctx.NodeURI = nodeURI
	ctx.Client = rpcclient.NewHTTP(nodeURI, "/websocket")
	return ctx
}

// WithClient returns a copy of the context with an updated RPC client
// instance.
func (ctx CLIContext) WithClient(client rpcclient.Client) CLIContext {
	ctx.Client = client
	return ctx
}

// WithUseLedger returns a copy of the context with an updated UseLedger flag.
func (ctx CLIContext) WithUseLedger(useLedger bool) CLIContext {
	ctx.UseLedger = useLedger
	return ctx
}

// WithCertifier - return a copy of the context with an updated Certifier
func (ctx CLIContext) WithCertifier(certifier tmlite.Certifier) CLIContext {
	ctx.Certifier = certifier
	return ctx
}
