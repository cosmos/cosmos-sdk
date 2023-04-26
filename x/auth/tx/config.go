package tx

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/core/address"
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
	handler     *txsigning.HandlerMap
	decoder     sdk.TxDecoder
	encoder     sdk.TxEncoder
	jsonDecoder sdk.TxDecoder
	jsonEncoder sdk.TxEncoder
	protoCodec  codec.ProtoCodecMarshaler
}

type ConfigOptions struct {
	SigningHandler             *txsigning.HandlerMap
	SigningContext             *txsigning.Context
	EnabledSignModes           []signingtypes.SignMode
	TypeResolver               protoregistry.MessageTypeResolver
	FileResolver               txsigning.ProtoFileResolver
	AddressCodec               address.Codec
	ValidatorCodec             address.Codec
	TextualCoinMetadataQueryFn textual.CoinMetadataQueryFn
}

// DefaultSignModes are the default sign modes enabled for protobuf transactions.
var DefaultSignModes = []signingtypes.SignMode{
	signingtypes.SignMode_SIGN_MODE_DIRECT,
	signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
	signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	// We currently don't add SIGN_MODE_TEXTUAL as part of the default sign
	// modes, as it's not released yet (including the Ledger app). However,
	// textual's sign mode handler is already available in this package. If you
	// want to use textual for **TESTING** purposes, feel free to create a
	// handler that includes SIGN_MODE_TEXTUAL.
	// ref: Tracking issue for SIGN_MODE_TEXTUAL https://github.com/cosmos/cosmos-sdk/issues/11970
}

// NewTxConfig returns a new protobuf TxConfig using the provided ProtoCodec and sign modes. The
// first enabled sign mode will become the default sign mode.
//
// NOTE: Use NewTxConfigWithHandler to provide a custom signing handler in case the sign mode
// is not supported by default (eg: SignMode_SIGN_MODE_EIP_191). Use NewTxConfigWithOptions
// to enable SIGN_MODE_TEXTUAL (for testing purposes for now).
//
// We prefer to use depinject to provide client.TxConfig, but we permit this constructor usage.  Within the SDK,
// this constructor is primarily used in tests, but also sees usage in app chains like:
// https://github.com/evmos/evmos/blob/719363fbb92ff3ea9649694bd088e4c6fe9c195f/encoding/config.go#L37
func NewTxConfig(protoCodec codec.ProtoCodecMarshaler, enabledSignModes []signingtypes.SignMode,
	customSignModes ...txsigning.SignModeHandler,
) client.TxConfig {
	return NewTxConfigWithOptions(protoCodec, ConfigOptions{EnabledSignModes: enabledSignModes}, customSignModes...)
}

func NewTxConfigWithOptions(protoCodec codec.ProtoCodecMarshaler, configOptions ConfigOptions,
	customSignModes ...txsigning.SignModeHandler,
) client.TxConfig {
	var err error
	txConfig := &config{
		decoder:     DefaultTxDecoder(protoCodec),
		encoder:     DefaultTxEncoder(),
		jsonDecoder: DefaultJSONTxDecoder(protoCodec),
		jsonEncoder: DefaultJSONTxEncoder(protoCodec),
		protoCodec:  protoCodec,
	}

	opts := &configOptions
	if opts.SigningHandler != nil {
		txConfig.handler = opts.SigningHandler
		return txConfig
	}

	if opts.TypeResolver == nil {
		opts.TypeResolver = protoregistry.GlobalTypes
	}
	if opts.FileResolver == nil {
		opts.FileResolver = protoCodec.InterfaceRegistry()
	}
	if len(opts.EnabledSignModes) == 0 {
		opts.EnabledSignModes = DefaultSignModes
	}

	if opts.SigningContext == nil {
		if opts.AddressCodec == nil {
			opts.AddressCodec = authcodec.NewBech32Codec(sdk.Bech32MainPrefix)
		}
		if opts.ValidatorCodec == nil {
			opts.ValidatorCodec = authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr)
		}
		opts.SigningContext, err = txsigning.NewContext(txsigning.Options{
			FileResolver:          opts.FileResolver,
			AddressCodec:          opts.AddressCodec,
			ValidatorAddressCodec: opts.ValidatorCodec,
		})
		if err != nil {
			panic(err)
		}
	}

	lenSignModes := len(configOptions.EnabledSignModes)
	handlers := make([]txsigning.SignModeHandler, lenSignModes+len(customSignModes))
	for i, m := range configOptions.EnabledSignModes {
		switch m {
		case signingtypes.SignMode_SIGN_MODE_DIRECT:
			handlers[i] = &direct.SignModeHandler{}
		case signingtypes.SignMode_SIGN_MODE_DIRECT_AUX:
			handlers[i], err = directaux.NewSignModeHandler(directaux.SignModeHandlerOptions{
				TypeResolver:   opts.TypeResolver,
				SignersContext: opts.SigningContext,
			})
			if err != nil {
				panic(err)
			}
		case signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
			aminoJSONEncoder := aminojson.NewAminoJSON()
			handlers[i] = aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
				FileResolver: opts.FileResolver,
				TypeResolver: opts.TypeResolver,
				Encoder:      &aminoJSONEncoder,
			})
		case signingtypes.SignMode_SIGN_MODE_TEXTUAL:
			handlers[i], err = textual.NewSignModeHandler(textual.SignModeOptions{
				CoinMetadataQuerier: opts.TextualCoinMetadataQueryFn,
				FileResolver:        opts.FileResolver,
				TypeResolver:        opts.TypeResolver,
			})
			if opts.TextualCoinMetadataQueryFn == nil {
				panic("cannot enable SIGN_MODE_TEXTUAL without a TextualCoinMetadataQueryFn")
			}
			if err != nil {
				panic(err)
			}
		}
	}
	for i, m := range customSignModes {
		handlers[i+len(configOptions.EnabledSignModes)] = m
	}

	txConfig.handler = txsigning.NewHandlerMap(handlers...)
	return txConfig
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
