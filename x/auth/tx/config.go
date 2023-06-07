package tx

import (
	"fmt"

	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
)

type config struct {
	handler        *txsigning.HandlerMap
	decoder        sdk.TxDecoder
	encoder        sdk.TxEncoder
	jsonDecoder    sdk.TxDecoder
	jsonEncoder    sdk.TxEncoder
	protoCodec     codec.ProtoCodecMarshaler
	signingContext *txsigning.Context
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
//
// NOTE: Use NewTxConfigWithOptions to provide a custom signing handler in case the sign mode
// is not supported by default (eg: SignMode_SIGN_MODE_EIP_191), or to enable SIGN_MODE_TEXTUAL.
//
// We prefer to use depinject to provide client.TxConfig, but we permit this constructor usage. Within the SDK,
// this constructor is primarily used in tests, but also sees usage in app chains like:
// https://github.com/evmos/evmos/blob/719363fbb92ff3ea9649694bd088e4c6fe9c195f/encoding/config.go#L37
func NewTxConfig(protoCodec codec.ProtoCodecMarshaler, enabledSignModes []signingtypes.SignMode,
	customSignModes ...txsigning.SignModeHandler,
) client.TxConfig {
	txConfig, err := NewTxConfigWithOptions(protoCodec, ConfigOptions{
		EnabledSignModes: enabledSignModes,
		CustomSignModes:  customSignModes,
	})
	if err != nil {
		panic(err)
	}
	return txConfig
}

// NewDefaultSigningOptions returns the sdk default signing options used by x/tx.  This includes account and
// validator address prefix enabled codecs.
func NewDefaultSigningOptions() (*txsigning.Options, error) {
	sdkConfig := sdk.GetConfig()
	return &txsigning.Options{
		AddressCodec:          authcodec.NewBech32Codec(sdkConfig.GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdkConfig.GetBech32ValidatorAddrPrefix()),
	}, nil
}

// NewSigningHandlerMap returns a new txsigning.HandlerMap using the provided ConfigOptions.
// It is recommended to use types.InterfaceRegistry in the field ConfigOptions.FileResolver as shown in
// NewTxConfigWithOptions but this fn does not enforce it.
func NewSigningHandlerMap(configOptions ConfigOptions) (*txsigning.HandlerMap, error) {
	var err error
	configOpts := &configOptions
	if configOpts.SigningOptions == nil {
		configOpts.SigningOptions, err = NewDefaultSigningOptions()
		if err != nil {
			return nil, err
		}
	}
	if configOpts.SigningContext == nil {
		configOpts.SigningContext, err = txsigning.NewContext(*configOpts.SigningOptions)
		if err != nil {
			return nil, err
		}
	}

	signingOpts := configOpts.SigningOptions

	if len(configOpts.EnabledSignModes) == 0 {
		configOpts.EnabledSignModes = DefaultSignModes
	}

	lenSignModes := len(configOpts.EnabledSignModes)
	handlers := make([]txsigning.SignModeHandler, lenSignModes+len(configOpts.CustomSignModes))
	for i, m := range configOpts.EnabledSignModes {
		var err error
		switch m {
		case signingtypes.SignMode_SIGN_MODE_DIRECT:
			handlers[i] = &direct.SignModeHandler{}
		case signingtypes.SignMode_SIGN_MODE_DIRECT_AUX:
			handlers[i], err = directaux.NewSignModeHandler(directaux.SignModeHandlerOptions{
				TypeResolver:   signingOpts.TypeResolver,
				SignersContext: configOpts.SigningContext,
			})
			if err != nil {
				return nil, err
			}
		case signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
			handlers[i] = aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
				FileResolver: signingOpts.FileResolver,
				TypeResolver: signingOpts.TypeResolver,
			})
		case signingtypes.SignMode_SIGN_MODE_TEXTUAL:
			handlers[i], err = textual.NewSignModeHandler(textual.SignModeOptions{
				CoinMetadataQuerier: configOpts.TextualCoinMetadataQueryFn,
				FileResolver:        signingOpts.FileResolver,
				TypeResolver:        signingOpts.TypeResolver,
			})
			if configOpts.TextualCoinMetadataQueryFn == nil {
				return nil, fmt.Errorf("cannot enable SIGN_MODE_TEXTUAL without a TextualCoinMetadataQueryFn")
			}
			if err != nil {
				return nil, err
			}
		}
	}
	for i, m := range configOpts.CustomSignModes {
		handlers[i+lenSignModes] = m
	}

	handler := txsigning.NewHandlerMap(handlers...)
	return handler, nil
}

// NewTxConfigWithOptions returns a new protobuf TxConfig using the provided ProtoCodec, ConfigOptions and
// custom sign mode handlers. If ConfigOptions is an empty struct then default values will be used.
func NewTxConfigWithOptions(protoCodec codec.ProtoCodecMarshaler, configOptions ConfigOptions) (client.TxConfig, error) {
	txConfig := &config{
		protoCodec: protoCodec,
	}
	if configOptions.ProtoDecoder == nil {
		txConfig.decoder = DefaultTxDecoder(protoCodec)
	}
	if configOptions.ProtoEncoder == nil {
		txConfig.encoder = DefaultTxEncoder()
	}
	if configOptions.JSONDecoder == nil {
		txConfig.jsonDecoder = DefaultJSONTxDecoder(protoCodec)
	}
	if configOptions.JSONEncoder == nil {
		txConfig.jsonEncoder = DefaultJSONTxEncoder(protoCodec)
	}

	var err error
	opts := &configOptions
	if opts.SigningContext == nil {
		signingOpts := configOptions.SigningOptions
		if signingOpts == nil {
			signingOpts, err = NewDefaultSigningOptions()
			if err != nil {
				return nil, err
			}
		}
		if signingOpts.FileResolver == nil {
			signingOpts.FileResolver = protoCodec.InterfaceRegistry()
		}
		opts.SigningContext, err = txsigning.NewContext(*signingOpts)
		if err != nil {
			return nil, err
		}
	}
	txConfig.signingContext = opts.SigningContext

	if opts.SigningHandler != nil {
		txConfig.handler = opts.SigningHandler
		return txConfig, nil
	}

	txConfig.handler, err = NewSigningHandlerMap(configOptions)
	if err != nil {
		return nil, err
	}

	return txConfig, nil
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

func (g config) SigningContext() *txsigning.Context {
	return g.signingContext
}
