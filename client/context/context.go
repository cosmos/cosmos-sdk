package context

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmlite "github.com/tendermint/tendermint/lite"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	verifier     tmlite.Verifier
	verifierHome string
)

// CLIContext implements a typical CLI context created in SDK modules for
// transaction handling and queries.
type CLIContext struct {
	Codec         *codec.Codec
	Client        rpcclient.Client
	Keybase       cryptokeys.Keybase
	Output        io.Writer
	OutputFormat  string
	Height        int64
	NodeURI       string
	From          string
	TrustNode     bool
	UseLedger     bool
	BroadcastMode string
	Verifier      tmlite.Verifier
	VerifierHome  string
	Simulate      bool
	GenerateOnly  bool
	FromAddress   sdk.AccAddress
	FromName      string
	Indent        bool
	SkipConfirm   bool
}

// NewCLIContextWithFrom returns a new initialized CLIContext with parameters from the
// command line using Viper. It takes a key name or address and populates the FromName and
// FromAddress field accordingly.
func NewCLIContextWithFrom(from string) CLIContext {
	var nodeURI string
	var rpc rpcclient.Client

	genOnly := viper.GetBool(flags.FlagGenerateOnly)
	fromAddress, fromName, err := GetFromFields(from, genOnly)
	if err != nil {
		fmt.Printf("failed to get from fields: %v", err)
		os.Exit(1)
	}

	if !genOnly {
		nodeURI = viper.GetString(flags.FlagNode)
		if nodeURI != "" {
			rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
		}
	}

	// We need to use a single verifier for all contexts
	if verifier == nil || verifierHome != viper.GetString(flags.FlagHome) {
		verifier = createVerifier()
		verifierHome = viper.GetString(flags.FlagHome)
	}

	return CLIContext{
		Client:        rpc,
		Output:        os.Stdout,
		NodeURI:       nodeURI,
		From:          viper.GetString(flags.FlagFrom),
		OutputFormat:  viper.GetString(cli.OutputFlag),
		Height:        viper.GetInt64(flags.FlagHeight),
		TrustNode:     viper.GetBool(flags.FlagTrustNode),
		UseLedger:     viper.GetBool(flags.FlagUseLedger),
		BroadcastMode: viper.GetString(flags.FlagBroadcastMode),
		Verifier:      verifier,
		Simulate:      viper.GetBool(flags.FlagDryRun),
		GenerateOnly:  genOnly,
		FromAddress:   fromAddress,
		FromName:      fromName,
		Indent:        viper.GetBool(flags.FlagIndentResponse),
		SkipConfirm:   viper.GetBool(flags.FlagSkipConfirmation),
	}
}

// NewCLIContext returns a new initialized CLIContext with parameters from the
// command line using Viper.
func NewCLIContext() CLIContext { return NewCLIContextWithFrom(viper.GetString(flags.FlagFrom)) }

func createVerifier() tmlite.Verifier {
	trustNodeDefined := viper.IsSet(flags.FlagTrustNode)
	if !trustNodeDefined {
		return nil
	}

	trustNode := viper.GetBool(flags.FlagTrustNode)
	if trustNode {
		return nil
	}

	chainID := viper.GetString(flags.FlagChainID)
	home := viper.GetString(flags.FlagHome)
	nodeURI := viper.GetString(flags.FlagNode)

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
		chainID, filepath.Join(home, ".lite_verifier"),
		node, log.NewNopLogger(), cacheSize,
	)

	if err != nil {
		fmt.Printf("Create verifier failed: %s\n", err.Error())
		fmt.Printf("Please check network connection and verify the address of the node to connect to\n")
		os.Exit(1)
	}

	return verifier
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
	ctx.Client = rpcclient.NewHTTP(nodeURI, "/websocket")
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

// WithVerifier - return a copy of the context with an updated Verifier
func (ctx CLIContext) WithVerifier(verifier tmlite.Verifier) CLIContext {
	ctx.Verifier = verifier
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
func (ctx CLIContext) PrintOutput(toPrint fmt.Stringer) (err error) {
	var out []byte

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
		return
	}

	fmt.Println(string(out))
	return
}

// GetFromFields returns a from account address and Keybase name given either
// an address or key name. If genOnly is true, only a valid Bech32 cosmos
// address is returned.
func GetFromFields(from string, genOnly bool) (sdk.AccAddress, string, error) {
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

	keybase, err := keys.NewKeyBaseFromHomeFlag()
	if err != nil {
		return nil, "", err
	}

	var info cryptokeys.Info
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
