package tx

import (
	"errors"
	"fmt"

	"cosmossdk.io/core/address"
	txdecode "cosmossdk.io/x/tx/decode"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type config struct {
	handler        *txsigning.HandlerMap
	decoder        sdk.TxDecoder
	encoder        sdk.TxEncoder
	jsonDecoder    sdk.TxDecoder
	jsonEncoder    sdk.TxEncoder
	protoCodec     codec.Codec
	signingContext *txsigning.Context
	txDecoder      *txdecode.Decoder
}

// ConfigOptions define the configuration of a TxConfig when calling NewTxConfigWithOptions.
// An empty struct is a valid configuration and will result in a TxConfig with default values.
type ConfigOptions struct {
	// If SigningHandler is specified it will be used instead constructing one.
	// This option supersedes all below options, whose sole purpose are to configure the creation of
	// txsigning.HandlerMap.
	SigningHandler *txsigning.HandlerMap
	// EnabledSignModes is the list of sign modes that will be enabled in the txsigning.HandlerMap.
	EnabledSignModes []signingtypes.SignMode
	// If SigningContext is specified it will be used when constructing sign mode handlers. If nil, one will be created
	// with the options specified in SigningOptions.
	SigningContext *txsigning.Context
	// SigningOptions are the options that will be used when constructing a txsigning.Context and sign mode handlers.
	// If nil defaults will be used.
	SigningOptions *txsigning.Options
	// TextualCoinMetadataQueryFn is the function that will be used to query coin metadata when constructing
	// textual sign mode handler. This is required if SIGN_MODE_TEXTUAL is enabled.
	TextualCoinMetadataQueryFn textual.CoinMetadataQueryFn
	// CustomSignModes are the custom sign modes that will be added to the txsigning.HandlerMap.
	CustomSignModes []txsigning.SignModeHandler
	// ProtoDecoder is the decoder that will be used to decode protobuf transactions.
	ProtoDecoder sdk.TxDecoder
	// ProtoEncoder is the encoder that will be used to encode protobuf transactions.
	ProtoEncoder sdk.TxEncoder
	// JSONDecoder is the decoder that will be used to decode json transactions.
	JSONDecoder sdk.TxDecoder
	// JSONEncoder is the encoder that will be used to encode json transactions.
	JSONEncoder sdk.TxEncoder
}

// DefaultSignModes are the default sign modes enabled for protobuf transactions.
var DefaultSignModes = []signingtypes.SignMode{
	signingtypes.SignMode_SIGN_MODE_DIRECT,
	signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
	signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	// signingtypes.SignMode_SIGN_MODE_TEXTUAL is not enabled by default, as it requires a x/bank keeper or gRPC connection.
}

// NewTxConfig returns a new protobuf TxConfig using the provided ProtoCodec and sign modes. The
// first enabled sign mode will become the default sign mode.
// NOTE: Use NewTxConfigWithHandler to provide a custom signing handler in case the sign mode
// is not supported by default (eg: SignMode_SIGN_MODE_EIP_191).
func NewTxConfig(protoCodec codec.ProtoCodecMarshaler, enabledSignModes []signingtypes.SignMode) client.TxConfig {
	return NewTxConfigWithHandler(protoCodec, makeSignModeHandler(enabledSignModes))
}

// NewTxConfig returns a new protobuf TxConfig using the provided ProtoCodec and signing handler.
func NewTxConfigWithHandler(protoCodec codec.ProtoCodecMarshaler, handler signing.SignModeHandler) client.TxConfig {
	return &config{
		handler:     handler,
		decoder:     DefaultTxDecoder(protoCodec),
		encoder:     DefaultTxEncoder(),
		jsonDecoder: DefaultJSONTxDecoder(protoCodec),
		jsonEncoder: DefaultJSONTxEncoder(protoCodec),
		protoCodec:  protoCodec,
		decoder:     configOptions.ProtoDecoder,
		encoder:     configOptions.ProtoEncoder,
		jsonDecoder: configOptions.JSONDecoder,
		jsonEncoder: configOptions.JSONEncoder,
	}

	var err error
	if configOptions.SigningContext == nil {
		if configOptions.SigningOptions == nil {
			return nil, errors.New("signing options not provided")
		}
		if configOptions.SigningOptions.FileResolver == nil {
			configOptions.SigningOptions.FileResolver = protoCodec.InterfaceRegistry()
		}
		configOptions.SigningContext, err = txsigning.NewContext(*configOptions.SigningOptions)
		if err != nil {
			return nil, err
		}
	}

	if configOptions.ProtoDecoder == nil {
		dec, err := txdecode.NewDecoder(txdecode.Options{
			SigningContext: configOptions.SigningContext,
			ProtoCodec:     protoCodec,
		},
		)
		if err != nil {
			return nil, err
		}
		txConfig.decoder = txV2toInterface(configOptions.SigningOptions.AddressCodec, protoCodec, dec)
		txConfig.txDecoder = dec
	}
	if configOptions.ProtoEncoder == nil {
		txConfig.encoder = DefaultTxEncoder()
	}
	if configOptions.JSONDecoder == nil {
		txConfig.jsonDecoder = DefaultJSONTxDecoder(configOptions.SigningOptions.AddressCodec, protoCodec, txConfig.txDecoder)
	}
	if configOptions.JSONEncoder == nil {
		txConfig.jsonEncoder = DefaultJSONTxEncoder(protoCodec)
	}

	txConfig.signingContext = configOptions.SigningContext

	if configOptions.SigningHandler != nil {
		txConfig.handler = configOptions.SigningHandler
		return txConfig, nil
	}

	txConfig.handler, err = NewSigningHandlerMap(configOptions)
	if err != nil {
		return nil, err
	}

	return txConfig, nil
}

func (g config) NewTxBuilder() client.TxBuilder {
	return newBuilder(g.signingContext.AddressCodec(), g.txDecoder, g.protoCodec)
}

// WrapTxBuilder returns a builder from provided transaction
func (g config) WrapTxBuilder(newTx sdk.Tx) (client.TxBuilder, error) {
	gogoTx, ok := newTx.(*gogoTxWrapper)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", &gogoTxWrapper{}, newTx)
	}

	return newBuilderFromDecodedTx(g.signingContext.AddressCodec(), g.txDecoder, g.protoCodec, gogoTx)
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

func (g config) SigningContext() *txsigning.Context {
	return g.signingContext
}

func txV2toInterface(addrCodec address.Codec, cdc codec.BinaryCodec, decoder *txdecode.Decoder) func([]byte) (sdk.Tx, error) {
	return func(txBytes []byte) (sdk.Tx, error) {
		decodedTx, err := decoder.Decode(txBytes)
		if err != nil {
			return nil, err
		}
		return newWrapperFromDecodedTx(addrCodec, cdc, decodedTx)
	}
}
