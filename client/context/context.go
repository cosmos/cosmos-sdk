package context

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cskeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmlite "github.com/tendermint/tendermint/lite"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

const ctxAccStoreName = "acc"

var (
	verifier tmlite.Verifier
)

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	Codec         *codec.Codec
	AccDecoder    auth.AccountDecoder
	Client        rpcclient.Client
	Output        io.Writer
	Height        int64
	NodeURI       string
	From          string
	AccountStore  string
	TrustNode     bool
	UseLedger     bool
	Async         bool
	JSON          bool
	PrintResponse bool
	Verifier      tmlite.Verifier
	DryRun        bool
	GenerateOnly  bool
	fromAddress   types.AccAddress
	fromName      string
	Indent        bool
}

// NewCLIContext returns a new initialized CLIContext with parameters from the
// command line using Viper.
func NewCLIContext() CLIContext {
	var rpc rpcclient.Client

	nodeURI := viper.GetString(client.FlagNode)
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}

	from := viper.GetString(client.FlagFrom)
	fromAddress, fromName := fromFields(from)

	// We need to use a single verifier for all contexts
	if verifier == nil {
		verifier = createVerifier()
	}

	return CLIContext{
		Client:        rpc,
		Output:        os.Stdout,
		NodeURI:       nodeURI,
		AccountStore:  ctxAccStoreName,
		From:          viper.GetString(client.FlagFrom),
		Height:        viper.GetInt64(client.FlagHeight),
		TrustNode:     viper.GetBool(client.FlagTrustNode),
		UseLedger:     viper.GetBool(client.FlagUseLedger),
		Async:         viper.GetBool(client.FlagAsync),
		JSON:          viper.GetBool(client.FlagJson),
		PrintResponse: viper.GetBool(client.FlagPrintResponse),
		Verifier:      verifier,
		DryRun:        viper.GetBool(client.FlagDryRun),
		GenerateOnly:  viper.GetBool(client.FlagGenerateOnly),
		fromAddress:   fromAddress,
		fromName:      fromName,
		Indent:        viper.GetBool(client.FlagIndentResponse),
	}
}

func createVerifier() tmlite.Verifier {
	trustNodeDefined := viper.IsSet(client.FlagTrustNode)
	if !trustNodeDefined {
		return nil
	}

	trustNode := viper.GetBool(client.FlagTrustNode)
	if trustNode {
		return nil
	}

	chainID := viper.GetString(client.FlagChainID)
	home := viper.GetString(cli.HomeFlag)
	nodeURI := viper.GetString(client.FlagNode)

	var errMsg bytes.Buffer
	if chainID == "" {
		errMsg.WriteString("--chain-id ")
	}
	if home == "" {
		errMsg.WriteString("--home ")
	}
	if nodeURI == "" {
		errMsg.WriteString("--node ")
	}
	if errMsg.Len() != 0 {
		fmt.Printf("Must specify these options: %s when --trust-node is false\n", errMsg.String())
		os.Exit(1)
	}

	node := rpcclient.NewHTTP(nodeURI, "/websocket")
	cacheSize := 10 // TODO: determine appropriate cache size
	verifier, err := tmliteProxy.NewVerifier(
		chainID, filepath.Join(home, ".gaialite"),
		node, log.NewNopLogger(), cacheSize,
	)

	if err != nil {
		fmt.Printf("Create verifier failed: %s\n", err.Error())
		fmt.Printf("Please check network connection and verify the address of the node to connect to\n")
		os.Exit(1)
	}

	return verifier
}

func fromFields(from string) (fromAddr types.AccAddress, fromName string) {
	if from == "" {
		return nil, ""
	}

	keybase, err := keys.GetKeyBase()
	if err != nil {
		fmt.Println("no keybase found")
		os.Exit(1)
	}

	var info cskeys.Info
	if addr, err := types.AccAddressFromBech32(from); err == nil {
		info, err = keybase.GetByAddress(addr)
		if err != nil {
			fmt.Printf("could not find key %s\n", from)
			os.Exit(1)
		}
	} else {
		info, err = keybase.Get(from)
		if err != nil {
			fmt.Printf("could not find key %s\n", from)
			os.Exit(1)
		}
	}

	fromAddr = info.GetAddress()
	fromName = info.GetName()
	return
}

// WithCodec returns a copy of the context with an updated codec.
func (ctx CLIContext) WithCodec(cdc *codec.Codec) CLIContext {
	ctx.Codec = cdc
	return ctx
}

// GetAccountDecoder gets the account decoder for auth.DefaultAccount.
func GetAccountDecoder(cdc *codec.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (acct auth.Account, err error) {
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}

		return acct, err
	}
}

// WithAccountDecoder returns a copy of the context with an updated account
// decoder.
func (ctx CLIContext) WithAccountDecoder(cdc *codec.Codec) CLIContext {
	ctx.AccDecoder = GetAccountDecoder(cdc)
	return ctx
}

// WithOutput returns a copy of the context with an updated output writer (e.g. stdout).
func (ctx CLIContext) WithOutput(w io.Writer) CLIContext {
	ctx.Output = w
	return ctx
}

// WithAccountStore returns a copy of the context with an updated AccountStore.
func (ctx CLIContext) WithAccountStore(accountStore string) CLIContext {
	ctx.AccountStore = accountStore
	return ctx
}

// WithFrom returns a copy of the context with an updated from address or name.
func (ctx CLIContext) WithFrom(from string) CLIContext {
	ctx.From = from
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

// WithVerifier - return a copy of the context with an updated Verifier
func (ctx CLIContext) WithVerifier(verifier tmlite.Verifier) CLIContext {
	ctx.Verifier = verifier
	return ctx
}
