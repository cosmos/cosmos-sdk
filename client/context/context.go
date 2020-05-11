package context

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	tmlite "github.com/tendermint/tendermint/lite"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Context implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type Context struct {
	FromAddress   sdk.AccAddress
	Client        rpcclient.Client
	ChainID       string
	Marshaler     codec.Marshaler
	Input         io.Reader
	Keyring       keyring.Keyring
	Output        io.Writer
	OutputFormat  string
	Height        int64
	HomeDir       string
	NodeURI       string
	From          string
	BroadcastMode string
	Verifier      tmlite.Verifier
	FromName      string
	TrustNode     bool
	UseLedger     bool
	Simulate      bool
	GenerateOnly  bool
	Offline       bool
	Indent        bool
	SkipConfirm   bool

	// TODO: Deprecated (remove).
	Codec *codec.Codec
}

// NewContextWithInputAndFrom returns a new initialized Context with parameters from the
// command line using Viper. It takes a io.Reader and and key name or address and populates
// the FromName and  FromAddress field accordingly. It will also create Tendermint verifier
// using  the chain ID, home directory and RPC URI provided by the command line. If using
// a Context in tests or any non CLI-based environment, the verifier will not be created
// and will be set as nil because FlagTrustNode must be set.
func NewContextWithInputAndFrom(input io.Reader, from string) Context {
	var nodeURI string
	var rpc rpcclient.Client

	homedir := viper.GetString(flags.FlagHome)
	genOnly := viper.GetBool(flags.FlagGenerateOnly)
	backend := viper.GetString(flags.FlagKeyringBackend)
	if len(backend) == 0 {
		backend = keyring.BackendMemory
	}

	keyring, err := newKeyringFromFlags(backend, homedir, input, genOnly)
	if err != nil {
		panic(fmt.Errorf("couldn't acquire keyring: %v", err))
	}

	fromAddress, fromName, err := GetFromFields(keyring, from, genOnly)
	if err != nil {
		fmt.Printf("failed to get from fields: %v\n", err)
		os.Exit(1)
	}

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
	ctx := Context{
		Client:        rpc,
		ChainID:       viper.GetString(flags.FlagChainID),
		Input:         input,
		Output:        os.Stdout,
		NodeURI:       nodeURI,
		From:          viper.GetString(flags.FlagFrom),
		Keyring:       keyring,
		OutputFormat:  viper.GetString(cli.OutputFlag),
		Height:        viper.GetInt64(flags.FlagHeight),
		HomeDir:       homedir,
		TrustNode:     trustNode,
		UseLedger:     viper.GetBool(flags.FlagUseLedger),
		BroadcastMode: viper.GetString(flags.FlagBroadcastMode),
		Simulate:      viper.GetBool(flags.FlagDryRun),
		GenerateOnly:  genOnly,
		Offline:       offline,
		FromAddress:   fromAddress,
		FromName:      fromName,
		Indent:        viper.GetBool(flags.FlagIndentResponse),
		SkipConfirm:   viper.GetBool(flags.FlagSkipConfirmation),
	}

	if offline {
		return ctx
	}

	// create a verifier for the specific chain ID and RPC client
	verifier, err := CreateVerifier(ctx, DefaultVerifierCacheSize)
	if err != nil && !trustNode {
		fmt.Printf("failed to create verifier: %s\n", err)
		os.Exit(1)
	}

	return ctx.WithVerifier(verifier)
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

// NewContextWithInput returns a new initialized Context with a io.Reader and parameters
// from the command line using Viper.
func NewContextWithInput(input io.Reader) Context {
	return NewContextWithInputAndFrom(input, viper.GetString(flags.FlagFrom))
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

// WithMarshaler returns a copy of the Context with an updated Marshaler.
func (ctx Context) WithMarshaler(m codec.Marshaler) Context {
	ctx.Marshaler = m
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

// Println outputs toPrint to the ctx.Output based on ctx.OutputFormat which is
// either text or json. If text, toPrint will be YAML encoded. Otherwise, toPrint
// will be JSON encoded using ctx.Marshaler. An error is returned upon failure.
func (ctx Context) Println(toPrint interface{}) error {
	var (
		out []byte
		err error
	)

	switch ctx.OutputFormat {
	case "text":
		out, err = yaml.Marshal(&toPrint)

	case "json":
		out, err = ctx.Marshaler.MarshalJSON(toPrint)

		// To JSON indent, we re-encode the already encoded JSON given there is no
		// error. The re-encoded JSON uses the standard library as the initial encoded
		// JSON should have the correct output produced by ctx.Marshaler.
		if ctx.Indent && err == nil {
			out, err = codec.MarshalIndentFromJSON(out)
		}
	}

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(ctx.Output, "%s\n", out)
	return err
}

// PrintOutput prints output while respecting output and indent flags
// NOTE: pass in marshalled structs that have been unmarshaled
// because this function will panic on marshaling errors.
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ref: https://github.com/cosmos/cosmos-sdk/issues/5864
func (ctx Context) PrintOutput(toPrint interface{}) error {
	var (
		out []byte
		err error
	)

	switch ctx.OutputFormat {
	case "text":
		out, err = yaml.Marshal(&toPrint)

	case "json":
		if ctx.Indent {
			out, err = ctx.Codec.MarshalJSONIndent(toPrint, "", "  ")
		} else {
			out, err = ctx.Codec.MarshalJSON(toPrint)
		}
	}

	if err != nil {
		return err
	}

	fmt.Println(string(out))
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

func newKeyringFromFlags(backend, homedir string, input io.Reader, genOnly bool) (keyring.Keyring, error) {
	if genOnly {
		return keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, homedir, input)
	}
	return keyring.New(sdk.KeyringServiceName(), backend, homedir, input)
}
