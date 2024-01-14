package tx

import (
	"bufio"
	"context"
	"cosmossdk.io/client/v2/autocli/keyring"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"sigs.k8s.io/yaml"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PreprocessTxFn defines a hook by which chains can preprocess transactions before broadcasting
type PreprocessTxFn func(chainID string, key keyring.KeyType, tx TxBuilder) error

// txContext implements a typical context created in SDK modules for transaction
// handling and queries.
type txContext struct {
	FromAddress           sdk.AccAddress
	Client                CometRPC
	GRPCClient            *grpc.ClientConn
	ChainID               string
	Codec                 codec.Codec
	InterfaceRegistry     codectypes.InterfaceRegistry
	Input                 io.Reader
	Keyring               keyring.Keyring
	KeyringOptions        []keyring.Option
	KeyringDir            string
	KeyringDefaultKeyName string
	Output                io.Writer
	OutputFormat          string
	Height                int64
	HomeDir               string
	From                  string
	BroadcastMode         string
	FromName              string
	SignModeStr           string
	Simulate              bool
	GenerateOnly          bool
	Offline               bool
	SkipConfirm           bool
	TxConfig              TxConfig
	AccountRetriever      AccountRetriever
	NodeURI               string
	FeePayer              sdk.AccAddress
	FeeGranter            sdk.AccAddress
	Viper                 *viper.Viper
	PreprocessTxHook      PreprocessTxFn

	// IsAux is true when the signer is an auxiliary signer (e.g. the tipper).
	IsAux bool

	// CmdContext is the context.txContext from the Cobra command.
	CmdContext context.Context

	// Address codecs
	AddressCodec          address.Codec
	ValidatorAddressCodec address.Codec
	ConsensusAddressCodec address.Codec
}

// WithCmdContext returns a copy of the context with an updated context.txContext,
// usually set to the cobra cmd context.
func (ctx txContext) WithCmdContext(c context.Context) txContext {
	ctx.CmdContext = c
	return ctx
}

// WithKeyring returns a copy of the context with an updated keyring.
func (ctx txContext) WithKeyring(k keyring.Keyring) txContext {
	ctx.Keyring = k
	return ctx
}

// WithKeyringOptions returns a copy of the context with an updated keyring.
func (ctx txContext) WithKeyringOptions(opts ...keyring.Option) txContext {
	ctx.KeyringOptions = opts
	return ctx
}

// WithInput returns a copy of the context with an updated input.
func (ctx txContext) WithInput(r io.Reader) txContext {
	// convert to a bufio.Reader to have a shared buffer between the keyring and the
	// the Commands, ensuring a read from one advance the read pointer for the other.
	// see https://github.com/cosmos/cosmos-sdk/issues/9566.
	ctx.Input = bufio.NewReader(r)
	return ctx
}

// WithCodec returns a copy of the txContext with an updated Codec.
func (ctx txContext) WithCodec(m codec.Codec) txContext {
	ctx.Codec = m
	return ctx
}

// WithOutput returns a copy of the context with an updated output writer (e.g. stdout).
func (ctx txContext) WithOutput(w io.Writer) txContext {
	ctx.Output = w
	return ctx
}

// WithFrom returns a copy of the context with an updated from address or name.
func (ctx txContext) WithFrom(from string) txContext {
	ctx.From = from
	return ctx
}

// WithOutputFormat returns a copy of the context with an updated OutputFormat field.
func (ctx txContext) WithOutputFormat(format string) txContext {
	ctx.OutputFormat = format
	return ctx
}

// WithNodeURI returns a copy of the context with an updated node URI.
func (ctx txContext) WithNodeURI(nodeURI string) txContext {
	ctx.NodeURI = nodeURI
	return ctx
}

// WithHeight returns a copy of the context with an updated height.
func (ctx txContext) WithHeight(height int64) txContext {
	ctx.Height = height
	return ctx
}

// WithClient returns a copy of the context with an updated RPC client
// instance.
func (ctx txContext) WithClient(client CometRPC) txContext {
	ctx.Client = client
	return ctx
}

// WithGRPCClient returns a copy of the context with an updated GRPC client
// instance.
func (ctx txContext) WithGRPCClient(grpcClient *grpc.ClientConn) txContext {
	ctx.GRPCClient = grpcClient
	return ctx
}

// WithChainID returns a copy of the context with an updated chain ID.
func (ctx txContext) WithChainID(chainID string) txContext {
	ctx.ChainID = chainID
	return ctx
}

// WithHomeDir returns a copy of the txContext with HomeDir set.
func (ctx txContext) WithHomeDir(dir string) txContext {
	if dir != "" {
		ctx.HomeDir = dir
	}
	return ctx
}

// WithKeyringDir returns a copy of the txContext with KeyringDir set.
func (ctx txContext) WithKeyringDir(dir string) txContext {
	ctx.KeyringDir = dir
	return ctx
}

// WithKeyringDefaultKeyName returns a copy of the txContext with KeyringDefaultKeyName set.
func (ctx txContext) WithKeyringDefaultKeyName(keyName string) txContext {
	ctx.KeyringDefaultKeyName = keyName
	return ctx
}

// WithGenerateOnly returns a copy of the context with updated GenerateOnly value
func (ctx txContext) WithGenerateOnly(generateOnly bool) txContext {
	ctx.GenerateOnly = generateOnly
	return ctx
}

// WithSimulation returns a copy of the context with updated Simulate value
func (ctx txContext) WithSimulation(simulate bool) txContext {
	ctx.Simulate = simulate
	return ctx
}

// WithOffline returns a copy of the context with updated Offline value.
func (ctx txContext) WithOffline(offline bool) txContext {
	ctx.Offline = offline
	return ctx
}

// WithFromName returns a copy of the context with an updated from account name.
func (ctx txContext) WithFromName(name string) txContext {
	ctx.FromName = name
	return ctx
}

// WithFromAddress returns a copy of the context with an updated from account
// address.
func (ctx txContext) WithFromAddress(addr sdk.AccAddress) txContext {
	ctx.FromAddress = addr
	return ctx
}

// WithFeePayerAddress returns a copy of the context with an updated fee payer account
// address.
func (ctx txContext) WithFeePayerAddress(addr sdk.AccAddress) txContext {
	ctx.FeePayer = addr
	return ctx
}

// WithFeeGranterAddress returns a copy of the context with an updated fee granter account
// address.
func (ctx txContext) WithFeeGranterAddress(addr sdk.AccAddress) txContext {
	ctx.FeeGranter = addr
	return ctx
}

// WithBroadcastMode returns a copy of the context with an updated broadcast
// mode.
func (ctx txContext) WithBroadcastMode(mode string) txContext {
	ctx.BroadcastMode = mode
	return ctx
}

// WithSignModeStr returns a copy of the context with an updated SignMode
// value.
func (ctx txContext) WithSignModeStr(signModeStr string) txContext {
	ctx.SignModeStr = signModeStr
	return ctx
}

// WithSkipConfirmation returns a copy of the context with an updated SkipConfirm
// value.
func (ctx txContext) WithSkipConfirmation(skip bool) txContext {
	ctx.SkipConfirm = skip
	return ctx
}

// WithTxConfig returns the context with an updated TxConfig
func (ctx txContext) WithTxConfig(generator TxConfig) txContext {
	ctx.TxConfig = generator
	return ctx
}

// WithAccountRetriever returns the context with an updated AccountRetriever
func (ctx txContext) WithAccountRetriever(retriever AccountRetriever) txContext {
	ctx.AccountRetriever = retriever
	return ctx
}

// WithInterfaceRegistry returns the context with an updated InterfaceRegistry
func (ctx txContext) WithInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) txContext {
	ctx.InterfaceRegistry = interfaceRegistry
	return ctx
}

// WithViper returns the context with Viper field. This Viper instance is used to read
// client-side config from the config file.
func (ctx txContext) WithViper(prefix string) txContext {
	v := viper.New()

	if prefix == "" {
		executableName, _ := os.Executable()
		prefix = path.Base(executableName)
	}

	v.SetEnvPrefix(prefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	ctx.Viper = v
	return ctx
}

// WithAux returns a copy of the context with an updated IsAux value.
func (ctx txContext) WithAux(isAux bool) txContext {
	ctx.IsAux = isAux
	return ctx
}

// WithPreprocessTxHook returns the context with the provided preprocessing hook, which
// enables chains to preprocess the transaction using the builder.
func (ctx txContext) WithPreprocessTxHook(preprocessFn PreprocessTxFn) txContext {
	ctx.PreprocessTxHook = preprocessFn
	return ctx
}

// WithAddressCodec returns the context with the provided address codec.
func (ctx txContext) WithAddressCodec(addressCodec address.Codec) txContext {
	ctx.AddressCodec = addressCodec
	return ctx
}

// WithValidatorAddressCodec returns the context with the provided validator address codec.
func (ctx txContext) WithValidatorAddressCodec(validatorAddressCodec address.Codec) txContext {
	ctx.ValidatorAddressCodec = validatorAddressCodec
	return ctx
}

// WithConsensusAddressCodec returns the context with the provided consensus address codec.
func (ctx txContext) WithConsensusAddressCodec(consensusAddressCodec address.Codec) txContext {
	ctx.ConsensusAddressCodec = consensusAddressCodec
	return ctx
}

// PrintString prints the raw string to ctx.Output if it's defined, otherwise to os.Stdout
func (ctx txContext) PrintString(str string) error {
	return ctx.PrintBytes([]byte(str))
}

// PrintBytes prints the raw bytes to ctx.Output if it's defined, otherwise to os.Stdout.
// NOTE: for printing a complex state object, you should use ctx.PrintOutput
func (ctx txContext) PrintBytes(o []byte) error {
	writer := ctx.Output
	if writer == nil {
		writer = os.Stdout
	}

	_, err := writer.Write(o)
	return err
}

// PrintProto outputs toPrint to the ctx.Output based on ctx.OutputFormat which is
// either text or json. If text, toPrint will be YAML encoded. Otherwise, toPrint
// will be JSON encoded using ctx.Codec. An error is returned upon failure.
func (ctx txContext) PrintProto(toPrint proto.Message) error {
	// always serialize JSON initially because proto json can't be directly YAML encoded
	out, err := ctx.Codec.MarshalJSON(toPrint)
	if err != nil {
		return err
	}
	return ctx.printOutput(out)
}

// PrintRaw is a variant of PrintProto that doesn't require a proto.Message type
// and uses a raw JSON message. No marshaling is performed.
func (ctx txContext) PrintRaw(toPrint json.RawMessage) error {
	return ctx.printOutput(toPrint)
}

func (ctx txContext) printOutput(out []byte) error {
	var err error
	if ctx.OutputFormat == "text" {
		out, err = yaml.JSONToYAML(out)
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

// GetFromFields returns a from account address, account name and keyring type, given either an address or key name.
// If clientCtx.Simulate is true the keystore is not accessed and a valid address must be provided
// If clientCtx.GenerateOnly is true the keystore is only accessed if a key name is provided
// If from is empty, the default key if specified in the context will be used
func GetFromFields(clientCtx txContext, kr keyring.Keyring, from string) (sdk.AccAddress, string, keyring.KeyType, error) {
	if from == "" && clientCtx.KeyringDefaultKeyName != "" {
		from = clientCtx.KeyringDefaultKeyName
		_ = clientCtx.PrintString(fmt.Sprintf("No key name or address provided; using the default key: %s\n", clientCtx.KeyringDefaultKeyName))
	}

	if from == "" {
		return nil, "", 0, nil
	}

	addr, err := clientCtx.AddressCodec.StringToBytes(from)
	switch {
	case clientCtx.Simulate:
		if err != nil {
			return nil, "", 0, fmt.Errorf("a valid address must be provided in simulation mode: %w", err)
		}

		return addr, "", 0, nil

	case clientCtx.GenerateOnly:
		if err == nil {
			return addr, "", 0, nil
		}
	}

	var k *keyring.Record
	if err == nil {
		k, err = kr.KeyByAddress(addr)
		if err != nil {
			return nil, "", 0, err
		}
	} else {
		k, err = kr.Key(from)
		if err != nil {
			return nil, "", 0, err
		}
	}

	addr, err = kr.GetRecordAddress(k)
	if err != nil {
		return nil, "", 0, err
	}

	return addr, kr.GetRecordName(k), kr.GetRecordType(k), nil
}
