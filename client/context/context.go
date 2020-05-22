package context

import (
	"bufio"
	"fmt"
	"io"
	"os"

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

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	FromAddress      sdk.AccAddress
	Client           rpcclient.Client
	ChainID          string
	JSONMarshaler    codec.JSONMarshaler
	Input            io.Reader
	Keyring          keyring.Keyring
	Output           io.Writer
	OutputFormat     string
	Height           int64
	HomeDir          string
	NodeURI          string
	From             string
	BroadcastMode    string
	Verifier         tmlite.Verifier
	FromName         string
	TrustNode        bool
	UseLedger        bool
	Simulate         bool
	GenerateOnly     bool
	Offline          bool
	Indent           bool
	SkipConfirm      bool
	TxGenerator      TxGenerator
	AccountRetriever AccountRetriever

	// TODO: Deprecated (remove).
	Codec *codec.Codec
}

// NewCLIContextWithInputAndFrom returns a new initialized CLIContext with parameters from the
// command line using Viper. It takes a io.Reader and and key name or address and populates
// the FromName and  FromAddress field accordingly. It will also create Tendermint verifier
// using  the chain ID, home directory and RPC URI provided by the command line. If using
// a CLIContext in tests or any non CLI-based environment, the verifier will not be created
// and will be set as nil because FlagTrustNode must be set.
func NewCLIContextWithInputAndFrom(input io.Reader, from string) CLIContext {
	ctx := CLIContext{}
	return ctx.InitWithInputAndFrom(input, from)
}

// NewCLIContextWithFrom returns a new initialized CLIContext with parameters from the
// command line using Viper. It takes a key name or address and populates the FromName and
// FromAddress field accordingly. It will also create Tendermint verifier using
// the chain ID, home directory and RPC URI provided by the command line. If using
// a CLIContext in tests or any non CLI-based environment, the verifier will not
// be created and will be set as nil because FlagTrustNode must be set.
func NewCLIContextWithFrom(from string) CLIContext {
	return NewCLIContextWithInputAndFrom(os.Stdin, from)
}

// NewCLIContext returns a new initialized CLIContext with parameters from the
// command line using Viper.
func NewCLIContext() CLIContext { return NewCLIContextWithFrom(viper.GetString(flags.FlagFrom)) }

// NewCLIContextWithInput returns a new initialized CLIContext with a io.Reader and parameters
// from the command line using Viper.
func NewCLIContextWithInput(input io.Reader) CLIContext {
	return NewCLIContextWithInputAndFrom(input, viper.GetString(flags.FlagFrom))
}

// InitWithInputAndFrom returns a new CLIContext re-initialized from an existing
// CLIContext with a new io.Reader and from parameter
func (ctx CLIContext) InitWithInputAndFrom(input io.Reader, from string) CLIContext {
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
	ctx.Indent = viper.GetBool(flags.FlagIndentResponse)
	ctx.SkipConfirm = viper.GetBool(flags.FlagSkipConfirmation)

	homedir := viper.GetString(flags.FlagHome)
	genOnly := viper.GetBool(flags.FlagGenerateOnly)
	backend := viper.GetString(flags.FlagKeyringBackend)
	if len(backend) == 0 {
		backend = keyring.BackendMemory
	}

	kr, err := newKeyringFromFlags(backend, homedir, input, genOnly)
	if err != nil {
		panic(fmt.Errorf("couldn't acquire keyring: %v", err))
	}

	fromAddress, fromName, err := GetFromFields(kr, from, genOnly)
	if err != nil {
		fmt.Printf("failed to get from fields: %v\n", err)
		os.Exit(1)
	}

	ctx.HomeDir = homedir

	ctx.Keyring = kr
	ctx.FromAddress = fromAddress
	ctx.FromName = fromName
	ctx.GenerateOnly = genOnly

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

// InitWithFrom returns a new CLIContext re-initialized from an existing
// CLIContext with a new from parameter
func (ctx CLIContext) InitWithFrom(from string) CLIContext {
	return ctx.InitWithInputAndFrom(os.Stdin, from)
}

// Init returns a new CLIContext re-initialized from an existing
// CLIContext with parameters from the command line using Viper.
func (ctx CLIContext) Init() CLIContext { return ctx.InitWithFrom(viper.GetString(flags.FlagFrom)) }

// InitWithInput returns a new CLIContext re-initialized from an existing
// CLIContext with a new io.Reader and from parameter
func (ctx CLIContext) InitWithInput(input io.Reader) CLIContext {
	return ctx.InitWithInputAndFrom(input, viper.GetString(flags.FlagFrom))
}

// WithKeyring returns a copy of the context with an updated keyring.
func (ctx CLIContext) WithKeyring(k keyring.Keyring) CLIContext {
	ctx.Keyring = k
	return ctx
}

// WithInput returns a copy of the context with an updated input.
func (ctx CLIContext) WithInput(r io.Reader) CLIContext {
	ctx.Input = r
	return ctx
}

// WithJSONMarshaler returns a copy of the CLIContext with an updated JSONMarshaler.
func (ctx CLIContext) WithJSONMarshaler(m codec.JSONMarshaler) CLIContext {
	ctx.JSONMarshaler = m
	return ctx
}

// WithCodec returns a copy of the context with an updated codec.
// TODO: Deprecated (remove).
func (ctx CLIContext) WithCodec(cdc *codec.Codec) CLIContext {
	ctx.Codec = cdc
	return ctx
}

// WithOutput returns a copy of the context with an updated output writer (e.g. stdout).
func (ctx CLIContext) WithOutput(w io.Writer) CLIContext {
	ctx.Output = w
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
	client, err := rpchttp.New(nodeURI, "/websocket")
	if err != nil {
		panic(err)
	}

	ctx.Client = client
	return ctx
}

// WithHeight returns a copy of the context with an updated height.
func (ctx CLIContext) WithHeight(height int64) CLIContext {
	ctx.Height = height
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

// WithVerifier returns a copy of the context with an updated Verifier.
func (ctx CLIContext) WithVerifier(verifier tmlite.Verifier) CLIContext {
	ctx.Verifier = verifier
	return ctx
}

// WithChainID returns a copy of the context with an updated chain ID.
func (ctx CLIContext) WithChainID(chainID string) CLIContext {
	ctx.ChainID = chainID
	return ctx
}

// WithGenerateOnly returns a copy of the context with updated GenerateOnly value
func (ctx CLIContext) WithGenerateOnly(generateOnly bool) CLIContext {
	ctx.GenerateOnly = generateOnly
	return ctx
}

// WithSimulation returns a copy of the context with updated Simulate value
func (ctx CLIContext) WithSimulation(simulate bool) CLIContext {
	ctx.Simulate = simulate
	return ctx
}

// WithFromName returns a copy of the context with an updated from account name.
func (ctx CLIContext) WithFromName(name string) CLIContext {
	ctx.FromName = name
	return ctx
}

// WithFromAddress returns a copy of the context with an updated from account
// address.
func (ctx CLIContext) WithFromAddress(addr sdk.AccAddress) CLIContext {
	ctx.FromAddress = addr
	return ctx
}

// WithBroadcastMode returns a copy of the context with an updated broadcast
// mode.
func (ctx CLIContext) WithBroadcastMode(mode string) CLIContext {
	ctx.BroadcastMode = mode
	return ctx
}

// WithTxGenerator returns the context with an updated TxGenerator
func (ctx CLIContext) WithTxGenerator(generator TxGenerator) CLIContext {
	ctx.TxGenerator = generator
	return ctx
}

// WithAccountRetriever returns the context with an updated AccountRetriever
func (ctx CLIContext) WithAccountRetriever(retriever AccountRetriever) CLIContext {
	ctx.AccountRetriever = retriever
	return ctx
}

// Println outputs toPrint to the ctx.Output based on ctx.OutputFormat which is
// either text or json. If text, toPrint will be YAML encoded. Otherwise, toPrint
// will be JSON encoded using ctx.JSONMarshaler. An error is returned upon failure.
func (ctx CLIContext) Println(toPrint interface{}) error {
	var (
		out []byte
		err error
	)

	switch ctx.OutputFormat {
	case "text":
		out, err = yaml.Marshal(&toPrint)

	case "json":
		out, err = ctx.JSONMarshaler.MarshalJSON(toPrint)

		// To JSON indent, we re-encode the already encoded JSON given there is no
		// error. The re-encoded JSON uses the standard library as the initial encoded
		// JSON should have the correct output produced by ctx.JSONMarshaler.
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
func (ctx CLIContext) PrintOutput(toPrint interface{}) error {
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
