package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/libs/cli"
	tmlite "github.com/tendermint/tendermint/lite"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Context implements a typical context created in SDK modules for transaction
// handling and queries.
type Context struct {
	FromAddress       sdk.AccAddress
	Client            rpcclient.Client
	ChainID           string
	JSONMarshaler     codec.JSONMarshaler
	InterfaceRegistry codectypes.InterfaceRegistry
	Input             io.Reader
	Keyring           keyring.Keyring
	Output            io.Writer
	OutputFormat      string
	Height            int64
	HomeDir           string
	From              string
	BroadcastMode     string
	FromName          string
	TrustNode         bool
	UseLedger         bool
	Simulate          bool
	GenerateOnly      bool
	Offline           bool
	SkipConfirm       bool
	TxConfig          TxConfig
	AccountRetriever  AccountRetriever
	NodeURI           string
	Verifier          tmlite.Verifier

	// TODO: Deprecated (remove).
	Codec *codec.Codec
}

// TODO: Remove all New* and Init* methods.

// NewContextWithInputAndFrom returns a new initialized Context with parameters from the
// command line using Viper. It takes a io.Reader and and key name or address and populates
// the FromName and  FromAddress field accordingly. It will also create Tendermint verifier
// using  the chain ID, home directory and RPC URI provided by the command line. If using
// a Context in tests or any non CLI-based environment, the verifier will not be created
// and will be set as nil because FlagTrustNode must be set.
func NewContextWithInputAndFrom(input io.Reader, from string) Context {
	ctx := Context{}
	return ctx.InitWithInputAndFrom(input, from)
}

// NewContextWithFrom returns a new initialized Context with parameters from the
// command line using Viper. It takes a key name or address and populates the FromName and
// FromAddress field accordingly. It will also create Tendermint verifier using
// the chain ID, home directory and RPC URI provided by the command line. If using
// a Context in tests or any non CLI-based environment, the verifier will not
// be created and will be set as nil because FlagTrustNode must be set.
func NewContextWithFrom(from string) Context {
	return NewContextWithInputAndFrom(os.Stdin, from)
}

// NewContext returns a new initialized Context with parameters from the
// command line using Viper.
func NewContext() Context { return NewContextWithFrom(viper.GetString(flags.FlagFrom)) }

// InitWithInputAndFrom returns a new Context re-initialized from an existing
// Context with a new io.Reader and from parameter
func (ctx Context) InitWithInputAndFrom(input io.Reader, from string) Context {
	input = bufio.NewReader(input)

	var (
		nodeURI string
		rpc     rpcclient.Client
		err     error
	)

	offline := viper.GetBool(flags.FlagOffline)
	if !offline {
		nodeURI = viper.GetString(flags.FlagNode)
		if nodeURI != "" {
			rpc, err = rpchttp.New(nodeURI, "/websocket")
			if err != nil {
				fmt.Printf("failted to get client: %v\n", err)
				os.Exit(1)
			}
		}
	}

	trustNode := viper.GetBool(flags.FlagTrustNode)

	ctx.Client = rpc
	ctx.ChainID = viper.GetString(flags.FlagChainID)
	ctx.Input = input
	ctx.Output = os.Stdout
	ctx.NodeURI = nodeURI
	ctx.From = viper.GetString(flags.FlagFrom)
	ctx.OutputFormat = viper.GetString(cli.OutputFlag)
	ctx.Height = viper.GetInt64(flags.FlagHeight)
	ctx.TrustNode = trustNode
	ctx.UseLedger = viper.GetBool(flags.FlagUseLedger)
	ctx.BroadcastMode = viper.GetString(flags.FlagBroadcastMode)
	ctx.Simulate = viper.GetBool(flags.FlagDryRun)
	ctx.Offline = offline
	ctx.SkipConfirm = viper.GetBool(flags.FlagSkipConfirmation)
	ctx.HomeDir = viper.GetString(flags.FlagHome)
	ctx.GenerateOnly = viper.GetBool(flags.FlagGenerateOnly)

	backend := viper.GetString(flags.FlagKeyringBackend)
	if len(backend) == 0 {
		backend = keyring.BackendMemory
	}

	kr, err := newKeyringFromFlags(ctx, backend)
	if err != nil {
		panic(fmt.Errorf("couldn't acquire keyring: %v", err))
	}

	fromAddress, fromName, err := GetFromFields(kr, from, ctx.GenerateOnly)
	if err != nil {
		fmt.Printf("failed to get from fields: %v\n", err)
		os.Exit(1)
	}

	ctx.Keyring = kr
	ctx.FromAddress = fromAddress
	ctx.FromName = fromName

	if offline {
		return ctx
	}

	// create a verifier for the specific chain ID and RPC client
	verifier, err := CreateVerifier(ctx, DefaultVerifierCacheSize)
	if err != nil && !trustNode {
		fmt.Printf("failed to create verifier: %s\n", err)
		os.Exit(1)
	}

	ctx.Verifier = verifier
	return ctx
}

// InitWithFrom returns a new Context re-initialized from an existing
// Context with a new from parameter
func (ctx Context) InitWithFrom(from string) Context {
	return ctx.InitWithInputAndFrom(os.Stdin, from)
}

// Init returns a new Context re-initialized from an existing
// Context with parameters from the command line using Viper.
func (ctx Context) Init() Context { return ctx.InitWithFrom(viper.GetString(flags.FlagFrom)) }

// InitWithInput returns a new Context re-initialized from an existing
// Context with a new io.Reader and from parameter
func (ctx Context) InitWithInput(input io.Reader) Context {
	return ctx.InitWithInputAndFrom(input, viper.GetString(flags.FlagFrom))
}

// WithKeyring returns a copy of the context with an updated keyring.
func (ctx Context) WithKeyring(k keyring.Keyring) Context {
	ctx.Keyring = k
	return ctx
}

// WithInput returns a copy of the context with an updated input.
func (ctx Context) WithInput(r io.Reader) Context {
	ctx.Input = r
	return ctx
}

// WithJSONMarshaler returns a copy of the Context with an updated JSONMarshaler.
func (ctx Context) WithJSONMarshaler(m codec.JSONMarshaler) Context {
	ctx.JSONMarshaler = m
	return ctx
}

// WithCodec returns a copy of the context with an updated codec.
// TODO: Deprecated (remove).
func (ctx Context) WithCodec(cdc *codec.Codec) Context {
	ctx.Codec = cdc
	return ctx
}

// WithOutput returns a copy of the context with an updated output writer (e.g. stdout).
func (ctx Context) WithOutput(w io.Writer) Context {
	ctx.Output = w
	return ctx
}

// WithFrom returns a copy of the context with an updated from address or name.
func (ctx Context) WithFrom(from string) Context {
	ctx.From = from
	return ctx
}

// WithTrustNode returns a copy of the context with an updated TrustNode flag.
func (ctx Context) WithTrustNode(trustNode bool) Context {
	ctx.TrustNode = trustNode
	return ctx
}

// WithOutputFormat returns a copy of the context with an updated OutputFormat field.
func (ctx Context) WithOutputFormat(format string) Context {
	ctx.OutputFormat = format
	return ctx
}

// WithNodeURI returns a copy of the context with an updated node URI.
func (ctx Context) WithNodeURI(nodeURI string) Context {
	ctx.NodeURI = nodeURI
	client, err := rpchttp.New(nodeURI, "/websocket")
	if err != nil {
		panic(err)
	}

	ctx.Client = client
	return ctx
}

// WithHeight returns a copy of the context with an updated height.
func (ctx Context) WithHeight(height int64) Context {
	ctx.Height = height
	return ctx
}

// WithClient returns a copy of the context with an updated RPC client
// instance.
func (ctx Context) WithClient(client rpcclient.Client) Context {
	ctx.Client = client
	return ctx
}

// WithUseLedger returns a copy of the context with an updated UseLedger flag.
func (ctx Context) WithUseLedger(useLedger bool) Context {
	ctx.UseLedger = useLedger
	return ctx
}

// WithVerifier returns a copy of the context with an updated Verifier.
func (ctx Context) WithVerifier(verifier tmlite.Verifier) Context {
	ctx.Verifier = verifier
	return ctx
}

// WithChainID returns a copy of the context with an updated chain ID.
func (ctx Context) WithChainID(chainID string) Context {
	ctx.ChainID = chainID
	return ctx
}

// WithHomeDir returns a copy of the Context with HomeDir set.
func (ctx Context) WithHomeDir(dir string) Context {
	ctx.HomeDir = dir
	return ctx
}

// WithGenerateOnly returns a copy of the context with updated GenerateOnly value
func (ctx Context) WithGenerateOnly(generateOnly bool) Context {
	ctx.GenerateOnly = generateOnly
	return ctx
}

// WithSimulation returns a copy of the context with updated Simulate value
func (ctx Context) WithSimulation(simulate bool) Context {
	ctx.Simulate = simulate
	return ctx
}

// WithOffline returns a copy of the context with updated Offline value.
func (ctx Context) WithOffline(offline bool) Context {
	ctx.Offline = offline
	return ctx
}

// WithFromName returns a copy of the context with an updated from account name.
func (ctx Context) WithFromName(name string) Context {
	ctx.FromName = name
	return ctx
}

// WithFromAddress returns a copy of the context with an updated from account
// address.
func (ctx Context) WithFromAddress(addr sdk.AccAddress) Context {
	ctx.FromAddress = addr
	return ctx
}

// WithBroadcastMode returns a copy of the context with an updated broadcast
// mode.
func (ctx Context) WithBroadcastMode(mode string) Context {
	ctx.BroadcastMode = mode
	return ctx
}

// WithSkipConfirmation returns a copy of the context with an updated SkipConfirm
// value.
func (ctx Context) WithSkipConfirmation(skip bool) Context {
	ctx.SkipConfirm = skip
	return ctx
}

// WithTxConfig returns the context with an updated TxConfig
func (ctx Context) WithTxConfig(generator TxConfig) Context {
	ctx.TxConfig = generator
	return ctx
}

// WithAccountRetriever returns the context with an updated AccountRetriever
func (ctx Context) WithAccountRetriever(retriever AccountRetriever) Context {
	ctx.AccountRetriever = retriever
	return ctx
}

// WithInterfaceRegistry returns the context with an updated InterfaceRegistry
func (ctx Context) WithInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) Context {
	ctx.InterfaceRegistry = interfaceRegistry
	return ctx
}

// PrintOutput outputs toPrint to the ctx.Output based on ctx.OutputFormat which is
// either text or json. If text, toPrint will be YAML encoded. Otherwise, toPrint
// will be JSON encoded using ctx.JSONMarshaler. An error is returned upon failure.
func (ctx Context) PrintOutput(toPrint interface{}) error {
	// always serialize JSON initially because proto json can't be directly YAML encoded
	out, err := ctx.JSONMarshaler.MarshalJSON(toPrint)
	if err != nil {
		return err
	}

	if ctx.OutputFormat == "text" {
		// handle text format by decoding and re-encoding JSON as YAML
		var j interface{}

		err = json.Unmarshal(out, &j)
		if err != nil {
			return err
		}

		out, err = yaml.Marshal(j)
		if err != nil {
			return err
		}
	}

	writer := ctx.Output
	if writer == nil {
		writer = os.Stdout
	}

	_, err = writer.Write(out)
	if err != nil {
		return err
	}

	if ctx.OutputFormat != "text" {
		// append new-line for formats besides YAML
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFromFields returns a from account address and Keybase name given either
// an address or key name. If genOnly is true, only a valid Bech32 cosmos
// address is returned.
func GetFromFields(kr keyring.Keyring, from string, genOnly bool) (sdk.AccAddress, string, error) {
	if from == "" {
		return nil, "", nil
	}

	if genOnly {
		addr, err := sdk.AccAddressFromBech32(from)
		if err != nil {
			return nil, "", errors.Wrap(err, "must provide a valid Bech32 address in generate-only mode")
		}

		return addr, "", nil
	}

	var info keyring.Info
	if addr, err := sdk.AccAddressFromBech32(from); err == nil {
		info, err = kr.KeyByAddress(addr)
		if err != nil {
			return nil, "", err
		}
	} else {
		info, err = kr.Key(from)
		if err != nil {
			return nil, "", err
		}
	}

	return info.GetAddress(), info.GetName(), nil
}

func newKeyringFromFlags(ctx Context, backend string) (keyring.Keyring, error) {
	if ctx.GenerateOnly {
		return keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, ctx.HomeDir, ctx.Input)
	}

	return keyring.New(sdk.KeyringServiceName(), backend, ctx.HomeDir, ctx.Input)
}
