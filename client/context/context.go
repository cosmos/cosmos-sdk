package context

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cskeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/tendermint/tendermint/libs/cli"
	tmlite "github.com/tendermint/tendermint/lite"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	Codec         *codec.Codec
	AccDecoder    auth.AccountDecoder
	Client        rpcclient.Client
	TxBldr        authtxb.TxBuilder
	Output        io.Writer
	OutputFormat  string
	Height        int64
	NodeURI       string
	From          string
	AccountStore  string
	TrustNode     bool
	UseLedger     bool
	Async         bool
	PrintResponse bool
	Verifier      tmlite.Verifier
	Simulate      bool
	GenerateOnly  bool
	fromAddress   types.AccAddress
	fromName      string
	Indent        bool
}

// NewCLIContext returns a new initialized CLIContext with parameters from the
// command line using Viper.
func NewCLIContext(cdc *codec.Codec) *CLIContext {
	var rpc rpcclient.Client

	nodeURI := viper.GetString(client.FlagNode)
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}

	fromAddress, fromName := fromFields(viper.GetString(client.FlagFrom))

	// We need to use a single verifier for all contexts
	if verifier == nil {
		verifier = createVerifier()
	}

	return &CLIContext{
		Client:        rpc,
		Codec:         cdc,
		Output:        os.Stdout,
		NodeURI:       nodeURI,
		AccountStore:  auth.StoreKey,
		From:          viper.GetString(client.FlagFrom),
		OutputFormat:  viper.GetString(cli.OutputFlag),
		Height:        viper.GetInt64(client.FlagHeight),
		TrustNode:     viper.GetBool(client.FlagTrustNode),
		UseLedger:     viper.GetBool(client.FlagUseLedger),
		Async:         viper.GetBool(client.FlagAsync),
		PrintResponse: viper.GetBool(client.FlagPrintResponse),
		Verifier:      verifier,
		Simulate:      viper.GetBool(client.FlagDryRun),
		GenerateOnly:  viper.GetBool(client.FlagGenerateOnly),
		fromAddress:   fromAddress,
		fromName:      fromName,
		Indent:        viper.GetBool(client.FlagIndentResponse),
	}
}

func NewCLIContextTx(cdc *codec.Codec) *CLIContext {
	ctx := NewCLIContext(cdc).SetAccountDecoder()
	ctx.PrepareTxBldrOffline()
	return ctx
}

// PrepareTxBldrOffline sets the transaction builder for the context w/o
// setting the sequence or account numbers
func (ctx *CLIContext) PrepareTxBldrOffline() {
	ctx.TxBldr = authtxb.NewTxBuilderFromCLI().WithTxEncoder(GetTxEncoder(ctx.Codec))
}

// PrepareTxBldr sets the transaction builder for the context while
// setting the sequence and account numbers
func (ctx *CLIContext) PrepareTxBldr() error {
	from, err := ctx.GetFromAddress()
	if err != nil {
		return err
	}
	return ctx.PrepareTxBldrWithAddress(from)
}

// PrepareTxBldrWithAddress looks up the acc and seq numbs for an addr, also
// ensuring it exists
func (ctx *CLIContext) PrepareTxBldrWithAddress(addr sdk.AccAddress) error {
	if err := ctx.EnsureAccountExists(addr); err != nil {
		return err
	}

	if ctx.TxBldr.GetAccountNumber() == 0 || ctx.TxBldr.GetSequence() == 0 {
		accNum, seq, err := ctx.FetchAccAndSeqNums(addr)
		if err != nil {
			return err
		}
		ctx.TxBldr = ctx.TxBldr.WithAccountNumber(accNum).WithSequence(seq)
	}
	return nil
}

// SetAccountDecoder returns a copy of the context with an updated account
// decoder.
func (ctx *CLIContext) SetAccountDecoder() *CLIContext {
	ctx.AccDecoder = ctx.GetAccountDecoder()
	return ctx
}

// GetAccountDecoder gets the account decoder for auth.DefaultAccount.
func (ctx *CLIContext) GetAccountDecoder() auth.AccountDecoder {
	return func(accBytes []byte) (acct auth.Account, err error) {
		err = ctx.Codec.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			// TODO: remove this, and return the error,
			// but first figure out where this is used
			panic(err)
		}
		return acct, err
	}
}

// GetFromAddress returns the from address from the context's name.
func (ctx *CLIContext) GetFromAddress() (sdk.AccAddress, error) {
	return ctx.fromAddress, nil
}

// GetFromName returns the key name for the current context.
func (ctx *CLIContext) GetFromName() (string, error) {
	return ctx.fromName, nil
}

// SetNode returns a copy of the context with an updated node URI.
func (ctx *CLIContext) SetNode(nodeURI string) *CLIContext {
	ctx.NodeURI = nodeURI
	ctx.Client = rpcclient.NewHTTP(nodeURI, "/websocket")
	return ctx
}

// GetNode returns an RPC client. If the context's client is not defined, an
// error is returned.
func (ctx *CLIContext) GetNode() (rpcclient.Client, error) {
	if ctx.Client == nil {
		return nil, errors.New("no RPC client defined")
	}

	return ctx.Client, nil
}

// PrintOutput prints output while respecting output and indent flags
func (ctx *CLIContext) PrintOutput(toPrint fmt.Stringer) (err error) {
	var out []byte

	switch ctx.OutputFormat {
	case "text":
		out = []byte(toPrint.String())
	case "json":
		if ctx.Indent {
			out, err = ctx.Codec.MarshalJSONIndent(toPrint, "", " ")
		} else {
			out, err = ctx.Codec.MarshalJSON(toPrint)
		}
	}

	if err != nil {
		return
	}

	fmt.Println(string(out))
	return
}

// MessagesOutput respects flags while either generating a transaction
// for later signing, or signing and broadcasting those messages in a transaction
func (ctx *CLIContext) MessagesOutput(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
	}

	if ctx.GenerateOnly {
		return ctx.PrintUnsignedStdTx(ctx.Output, msgs, false)
	}

	return ctx.CompleteAndBroadcastTxCli(msgs)
}

// MessageOutput respects flags while either generating a transaction
// for later signing, or signing and broadcasting those messages in a transaction
func (ctx *CLIContext) MessageOutput(msg sdk.Msg) error {
	return ctx.MessagesOutput([]sdk.Msg{msg})
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
