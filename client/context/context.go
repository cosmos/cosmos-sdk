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
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	FromAddress   sdk.AccAddress
	Client        rpcclient.Client
	ChainID       string
	Keybase       keys.Keybase
	Input         io.Reader
	Output        io.Writer
	OutputFormat  string
	Height        int64
	HomeDir       string
	NodeURI       string
	From          string
	BroadcastMode string
	Verifier      tmlite.Verifier
	FromName      string
	Codec         *codec.Codec
	TrustNode     bool
	UseLedger     bool
	Simulate      bool
	GenerateOnly  bool
	Indent        bool
	SkipConfirm   bool
}

// NewCLIContextWithInputAndFrom returns a new initialized CLIContext with parameters from the
// command line using Viper. It takes a io.Reader and and key name or address and populates
// the FromName and  FromAddress field accordingly. It will also create Tendermint verifier
// using  the chain ID, home directory and RPC URI provided by the command line. If using
// a CLIContext in tests or any non CLI-based environment, the verifier will not be created
// and will be set as nil because FlagTrustNode must be set.
func NewCLIContextWithInputAndFrom(input io.Reader, from string) CLIContext {
	var nodeURI string
	var rpc rpcclient.Client

	genOnly := viper.GetBool(flags.FlagGenerateOnly)
	fromAddress, fromName, err := GetFromFields(input, from, genOnly)
	if err != nil {
		fmt.Printf("failed to get from fields: %v\n", err)
		os.Exit(1)
	}

	if !genOnly {
		nodeURI = viper.GetString(flags.FlagNode)
		if nodeURI != "" {
			rpc, err = rpchttp.New(nodeURI, "/websocket")
			if err != nil {
				fmt.Printf("failted to get client: %v\n", err)
				os.Exit(1)
			}
		}
	}

	ctx := CLIContext{
		Client:        rpc,
		ChainID:       viper.GetString(flags.FlagChainID),
		Input:         input,
		Output:        os.Stdout,
		NodeURI:       nodeURI,
		From:          viper.GetString(flags.FlagFrom),
		OutputFormat:  viper.GetString(cli.OutputFlag),
		Height:        viper.GetInt64(flags.FlagHeight),
		HomeDir:       viper.GetString(flags.FlagHome),
		TrustNode:     viper.GetBool(flags.FlagTrustNode),
		UseLedger:     viper.GetBool(flags.FlagUseLedger),
		BroadcastMode: viper.GetString(flags.FlagBroadcastMode),
		Simulate:      viper.GetBool(flags.FlagDryRun),
		GenerateOnly:  genOnly,
		FromAddress:   fromAddress,
		FromName:      fromName,
		Indent:        viper.GetBool(flags.FlagIndentResponse),
		SkipConfirm:   viper.GetBool(flags.FlagSkipConfirmation),
	}

	// create a verifier for the specific chain ID and RPC client
	verifier, err := CreateVerifier(ctx, DefaultVerifierCacheSize)
	if err != nil && viper.IsSet(flags.FlagTrustNode) {
		fmt.Printf("failed to create verifier: %s\n", err)
		os.Exit(1)
	}

	return ctx.WithVerifier(verifier)
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

// WithInput returns a copy of the context with an updated input.
func (ctx CLIContext) WithInput(r io.Reader) CLIContext {
	ctx.Input = r
	return ctx
}

// WithCodec returns a copy of the context with an updated codec.
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

// PrintOutput prints output while respecting output and indent flags
// NOTE: pass in marshalled structs that have been unmarshaled
// because this function will panic on marshaling errors
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
func GetFromFields(input io.Reader, from string, genOnly bool) (sdk.AccAddress, string, error) {
	if from == "" {
		return nil, "", nil
	}

	if genOnly {
		addr, err := sdk.AccAddressFromBech32(from)
		if err != nil {
			return nil, "", errors.Wrap(err, "must provide a valid Bech32 address for generate-only")
		}

		return addr, "", nil
	}

	keybase, err := keys.NewKeyring(sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), input)
	if err != nil {
		return nil, "", err
	}

	var info keys.Info
	if addr, err := sdk.AccAddressFromBech32(from); err == nil {
		info, err = keybase.GetByAddress(addr)
		if err != nil {
			return nil, "", err
		}
	} else {
		info, err = keybase.Get(from)
		if err != nil {
			return nil, "", err
		}
	}

	return info.GetAddress(), info.GetName(), nil
}
