package client

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	KeyringDir        string
	From              string
	BroadcastMode     string
	FromName          string
	SignModeStr       string
	UseLedger         bool
	Simulate          bool
	GenerateOnly      bool
	Offline           bool
	SkipConfirm       bool
	TxConfig          TxConfig
	AccountRetriever  AccountRetriever
	NodeURI           string
	Viper             *viper.Viper

	// TODO: Deprecated (remove).
	LegacyAmino *codec.LegacyAmino
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

// WithLegacyAmino returns a copy of the context with an updated LegacyAmino codec.
// TODO: Deprecated (remove).
func (ctx Context) WithLegacyAmino(cdc *codec.LegacyAmino) Context {
	ctx.LegacyAmino = cdc
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

// WithOutputFormat returns a copy of the context with an updated OutputFormat field.
func (ctx Context) WithOutputFormat(format string) Context {
	ctx.OutputFormat = format
	return ctx
}

// WithNodeURI returns a copy of the context with an updated node URI.
func (ctx Context) WithNodeURI(nodeURI string) Context {
	ctx.NodeURI = nodeURI
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

// WithChainID returns a copy of the context with an updated chain ID.
func (ctx Context) WithChainID(chainID string) Context {
	ctx.ChainID = chainID
	return ctx
}

// WithHomeDir returns a copy of the Context with HomeDir set.
func (ctx Context) WithHomeDir(dir string) Context {
	if dir != "" {
		ctx.HomeDir = dir
	}
	return ctx
}

// WithKeyringDir returns a copy of the Context with KeyringDir set.
func (ctx Context) WithKeyringDir(dir string) Context {
	ctx.KeyringDir = dir
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

// WithSignModeStr returns a copy of the context with an updated SignMode
// value.
func (ctx Context) WithSignModeStr(signModeStr string) Context {
	ctx.SignModeStr = signModeStr
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

// WithViper returns the context with Viper field. This Viper instance is used to read
// client-side config from the config file.
func (ctx Context) WithViper(prefix string) Context {
	v := viper.New()
	v.SetEnvPrefix(prefix)
	v.AutomaticEnv()
	ctx.Viper = v
	return ctx
}

// PrintString prints the raw string to ctx.Output if it's defined, otherwise to os.Stdout
func (ctx Context) PrintString(str string) error {
	return ctx.PrintBytes([]byte(str))
}

// PrintBytes prints the raw bytes to ctx.Output if it's defined, otherwise to os.Stdout.
// NOTE: for printing a complex state object, you should use ctx.PrintOutput
func (ctx Context) PrintBytes(o []byte) error {
	writer := ctx.Output
	if writer == nil {
		writer = os.Stdout
	}

	_, err := writer.Write(o)
	return err
}

// PrintProto outputs toPrint to the ctx.Output based on ctx.OutputFormat which is
// either text or json. If text, toPrint will be YAML encoded. Otherwise, toPrint
// will be JSON encoded using ctx.JSONMarshaler. An error is returned upon failure.
func (ctx Context) PrintProto(toPrint proto.Message) error {
	// always serialize JSON initially because proto json can't be directly YAML encoded
	out, err := ctx.JSONMarshaler.MarshalJSON(toPrint)
	if err != nil {
		return err
	}
	return ctx.printOutput(out)
}

// PrintObjectLegacy is a variant of PrintProto that doesn't require a proto.Message type
// and uses amino JSON encoding.
// Deprecated: It will be removed in the near future!
func (ctx Context) PrintObjectLegacy(toPrint interface{}) error {
	out, err := ctx.LegacyAmino.MarshalJSON(toPrint)
	if err != nil {
		return err
	}
	return ctx.printOutput(out)
}

func (ctx Context) printOutput(out []byte) error {
	if ctx.OutputFormat == "text" {
		// handle text format by decoding and re-encoding JSON as YAML
		var j interface{}

		err := json.Unmarshal(out, &j)
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

	_, err := writer.Write(out)
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

// GetFromFields returns a from account address, account name and keyring type, given either
// an address or key name. If genOnly is true, only a valid Bech32 cosmos
// address is returned.
func GetFromFields(kr keyring.Keyring, from string, genOnly bool) (sdk.AccAddress, string, keyring.KeyType, error) {
	if from == "" {
		return nil, "", 0, nil
	}

	if genOnly {
		addr, err := sdk.AccAddressFromBech32(from)
		if err != nil {
			return nil, "", 0, errors.Wrap(err, "must provide a valid Bech32 address in generate-only mode")
		}

		return addr, "", 0, nil
	}

	var info keyring.Info
	if addr, err := sdk.AccAddressFromBech32(from); err == nil {
		info, err = kr.KeyByAddress(addr)
		if err != nil {
			return nil, "", 0, err
		}
	} else {
		info, err = kr.Key(from)
		if err != nil {
			return nil, "", 0, err
		}
	}

	return info.GetAddress(), info.GetName(), info.GetType(), nil
}

// NewKeyringFromBackend gets a Keyring object from a backend
func NewKeyringFromBackend(ctx Context, backend string) (keyring.Keyring, error) {
	if ctx.GenerateOnly || ctx.Simulate {
		return keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, ctx.KeyringDir, ctx.Input)
	}

	return keyring.New(sdk.KeyringServiceName(), backend, ctx.KeyringDir, ctx.Input)
}
