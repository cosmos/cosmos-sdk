package tx

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoregistry"

	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type config struct {
	handler     *txsigning.HandlerMap
	decoder     sdk.TxDecoder
	encoder     sdk.TxEncoder
	jsonDecoder sdk.TxDecoder
	jsonEncoder sdk.TxEncoder
	protoCodec  codec.ProtoCodecMarshaler
}

// NewTxConfig returns a new protobuf TxConfig using the provided ProtoCodec and sign modes. The
// first enabled sign mode will become the default sign mode.
//
// NOTE: Use NewTxConfigWithHandler to provide a custom signing handler in case the sign mode
// is not supported by default (eg: SignMode_SIGN_MODE_EIP_191). Use NewTxConfigWithTextual
// to enable SIGN_MODE_TEXTUAL (for testing purposes for now).
//
// We prefer to use depinject to provide client.TxConfig, but we permit this constructor usage.  Within the SDK,
// this constructor is primarily used in tests, but also sees usage in app chains like:
// https://github.com/evmos/evmos/blob/719363fbb92ff3ea9649694bd088e4c6fe9c195f/encoding/config.go#L37
// TODO: collapse enabledSignModes and customSignModes
func NewTxConfig(protoCodec codec.ProtoCodecMarshaler, enabledSignModes []signingtypes.SignMode,
	customSignModes ...txsigning.SignModeHandler,
) client.TxConfig {
	typeResolver := protoregistry.GlobalTypes
	protoFiles := protoCodec.InterfaceRegistry()
	signersContext, err := txsigning.NewGetSignersContext(txsigning.GetSignersOptions{ProtoFiles: protoFiles})
	if err != nil {
		panic(err)
	}

	signModeOptions := &SignModeOptions{}
	for _, m := range enabledSignModes {
		switch m {
		case signingtypes.SignMode_SIGN_MODE_DIRECT:
			signModeOptions.Direct = &direct.SignModeHandler{}
		case signingtypes.SignMode_SIGN_MODE_DIRECT_AUX:
			signModeOptions.DirectAux = &directaux.SignModeHandlerOptions{
				FileResolver:   protoFiles,
				TypeResolver:   typeResolver,
				SignersContext: signersContext,
			}
		case signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
			aminoJSONEncoder := aminojson.NewAminoJSON()
			signModeOptions.AminoJSON = &aminojson.SignModeHandlerOptions{
				FileResolver: protoFiles,
				TypeResolver: typeResolver,
				Encoder:      &aminoJSONEncoder,
			}
		case signingtypes.SignMode_SIGN_MODE_TEXTUAL:
			panic("cannot use NewTxConfig with SIGN_MODE_TEXTUAL enabled; please use NewTxConfigWithTextual")
		}
	}

	return NewTxConfigWithHandler(protoCodec, makeSignModeHandler(*signModeOptions, customSignModes...))
}

func NewTxConfigWithOptions(protoCodec codec.ProtoCodecMarshaler, signModeOptions SignModeOptions,
	customSignModes ...txsigning.SignModeHandler,
) client.TxConfig {
	return NewTxConfigWithHandler(protoCodec, makeSignModeHandler(signModeOptions, customSignModes...))
}

// NewTxConfigWithTextual is like NewTxConfig with the ability to add
// a SIGN_MODE_TEXTUAL renderer. It is currently still EXPERIMENTAL, for should
// be used for TESTING purposes only, until Textual is fully released.
//
// Deprecated: use NewTxConfigWithOptions instead.
func NewTxConfigWithTextual(protoCodec codec.ProtoCodecMarshaler, _ []signingtypes.SignMode,
	signModeOptions SignModeOptions, customSignModes ...txsigning.SignModeHandler,
) client.TxConfig {
	return NewTxConfigWithOptions(protoCodec, signModeOptions, customSignModes...)
}

// NewTxConfigWithHandler returns a new protobuf TxConfig using the provided ProtoCodec and signing handler.
func NewTxConfigWithHandler(protoCodec codec.ProtoCodecMarshaler, handler *txsigning.HandlerMap) client.TxConfig {
	return &config{
		handler:     handler,
		decoder:     DefaultTxDecoder(protoCodec),
		encoder:     DefaultTxEncoder(),
		jsonDecoder: DefaultJSONTxDecoder(protoCodec),
		jsonEncoder: DefaultJSONTxEncoder(protoCodec),
		protoCodec:  protoCodec,
	}
}

func (g config) NewTxBuilder() client.TxBuilder {
	return newBuilder(g.protoCodec)
}

// WrapTxBuilder returns a builder from provided transaction
func (g config) WrapTxBuilder(newTx sdk.Tx) (client.TxBuilder, error) {
	newBuilder, ok := newTx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", &wrapper{}, newTx)
	}

	return newBuilder, nil
}

func (g config) SignModeHandler() *txsigning.HandlerMap {
	return g.handler
}

func (g config) TxEncoder() sdk.TxEncoder {
	return g.encoder
}

func (g config) TxDecoder() sdk.TxDecoder {
	return g.decoder
}

func (g config) TxJSONEncoder() sdk.TxEncoder {
	return g.jsonEncoder
}

func (g config) TxJSONDecoder() sdk.TxDecoder {
	return g.jsonDecoder
}
