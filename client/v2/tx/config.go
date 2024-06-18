package tx

import (
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/address"
	txdecode "cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"
	"errors"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	_ TxConfig         = txConfig{}
	_ TxEncodingConfig = defaultEncodingConfig{}
	_ TxSigningConfig  = defaultTxSigningConfig{}

	defaultEnabledSignModes = []apitxsigning.SignMode{
		apitxsigning.SignMode_SIGN_MODE_DIRECT,
		apitxsigning.SignMode_SIGN_MODE_DIRECT_AUX,
		apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}
)

// TxConfig defines an interface a client can utilize to generate an
// application-defined concrete transaction type. The type returned must
// implement TxBuilder.
type TxConfig interface {
	TxEncodingConfig
	TxSigningConfig
	TxBuilderProvider
}

// TxEncodingConfig defines an interface that contains transaction
// encoders and decoders
type TxEncodingConfig interface {
	TxEncoder() txApiEncoder
	TxDecoder() txApiDecoder
	TxJSONEncoder() txApiEncoder
	TxJSONDecoder() txApiDecoder
}

type TxSigningConfig interface {
	SignModeHandler() *signing.HandlerMap
	SigningContext() *signing.Context
	MarshalSignatureJSON([]Signature) ([]byte, error)
	UnmarshalSignatureJSON([]byte) ([]Signature, error)
}

type ConfigOptions struct {
	AddressCodec address.Codec
	Decoder      Decoder
	Cdc          codec.BinaryCodec

	ValidatorAddressCodec address.Codec
	FileResolver          signing.ProtoFileResolver
	TypeResolver          signing.TypeResolver
	CustomGetSigner       map[protoreflect.FullName]signing.GetSignersFunc
	MaxRecursionDepth     int

	EnablesSignModes           []apitxsigning.SignMode
	CustomSignModes            []signing.SignModeHandler
	TextualCoinMetadataQueryFn textual.CoinMetadataQueryFn
}

func (c *ConfigOptions) validate() error {
	if c.AddressCodec == nil {
		return errors.New("address codec cannot be nil")
	}
	if c.Cdc == nil {
		return errors.New("codec cannot be nil")
	}
	if c.ValidatorAddressCodec == nil {
		return errors.New("validator address codec cannot be nil")
	}

	// set default signModes
	if len(c.EnablesSignModes) == 0 {
		c.EnablesSignModes = defaultEnabledSignModes
	}
	return nil
}

type txConfig struct {
	TxBuilderProvider
	TxEncodingConfig
	TxSigningConfig
}

func NewTxConfig(options ConfigOptions) (TxConfig, error) {
	err := options.validate()
	if err != nil {
		return nil, err
	}

	signingCtx, err := newDefaultTxSigningConfig(options)
	if err != nil {
		return nil, err
	}

	if options.Decoder == nil {
		options.Decoder, err = txdecode.NewDecoder(txdecode.Options{
			SigningContext: signingCtx.SigningContext(),
			ProtoCodec:     options.Cdc})
		if err != nil {
			return nil, err
		}
	}

	return &txConfig{
		TxBuilderProvider: NewBuilderProvider(options.AddressCodec, options.Decoder, options.Cdc),
		TxEncodingConfig:  defaultEncodingConfig{},
		TxSigningConfig:   signingCtx,
	}, nil
}

type defaultEncodingConfig struct{}

func (t defaultEncodingConfig) TxEncoder() txApiEncoder {
	return txEncoder
}

func (t defaultEncodingConfig) TxDecoder() txApiDecoder {
	return txDecoder
}

func (t defaultEncodingConfig) TxJSONEncoder() txApiEncoder {
	return txJsonEncoder
}

func (t defaultEncodingConfig) TxJSONDecoder() txApiDecoder {
	return txJsonDecoder
}

type defaultTxSigningConfig struct {
	signingCtx *signing.Context
	handlerMap *signing.HandlerMap
}

func newDefaultTxSigningConfig(opts ConfigOptions) (*defaultTxSigningConfig, error) {
	signingCtx, err := newSigningContext(opts)
	if err != nil {
		return nil, err
	}

	handlerMap, err := newHandlerMap(opts, signingCtx)
	if err != nil {
		return nil, err
	}

	return &defaultTxSigningConfig{
		signingCtx: signingCtx,
		handlerMap: handlerMap,
	}, nil
}

func (t defaultTxSigningConfig) SignModeHandler() *signing.HandlerMap {
	return t.handlerMap
}

func (t defaultTxSigningConfig) SigningContext() *signing.Context {
	return t.signingCtx
}

func (t defaultTxSigningConfig) MarshalSignatureJSON(signatures []Signature) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t defaultTxSigningConfig) UnmarshalSignatureJSON(bytes []byte) ([]Signature, error) {
	//TODO implement me
	panic("implement me")
}

func newSigningContext(opts ConfigOptions) (*signing.Context, error) {
	return signing.NewContext(signing.Options{
		FileResolver:          opts.FileResolver,
		TypeResolver:          opts.TypeResolver,
		AddressCodec:          opts.AddressCodec,
		ValidatorAddressCodec: opts.ValidatorAddressCodec,
		CustomGetSigners:      opts.CustomGetSigner,
		MaxRecursionDepth:     opts.MaxRecursionDepth,
	})
}

func newHandlerMap(opts ConfigOptions, signingCtx *signing.Context) (*signing.HandlerMap, error) {
	lenSignModes := len(opts.EnablesSignModes)
	handlers := make([]signing.SignModeHandler, lenSignModes+len(opts.CustomSignModes))

	for i, m := range opts.EnablesSignModes {
		var err error
		switch m {
		case apitxsigning.SignMode_SIGN_MODE_DIRECT:
			handlers[i] = &direct.SignModeHandler{}
		case apitxsigning.SignMode_SIGN_MODE_TEXTUAL:
			if opts.TextualCoinMetadataQueryFn == nil {
				return nil, errors.New("cannot enable SIGN_MODE_TEXTUAL without a TextualCoinMetadataQueryFn")
			}
			handlers[i], err = textual.NewSignModeHandler(textual.SignModeOptions{
				CoinMetadataQuerier: opts.TextualCoinMetadataQueryFn,
				FileResolver:        signingCtx.FileResolver(),
				TypeResolver:        signingCtx.TypeResolver(),
			})
			if err != nil {
				return nil, err
			}
		case apitxsigning.SignMode_SIGN_MODE_DIRECT_AUX:
			handlers[i], err = directaux.NewSignModeHandler(directaux.SignModeHandlerOptions{
				TypeResolver:   signingCtx.TypeResolver(),
				SignersContext: signingCtx,
			})
			if err != nil {
				return nil, err
			}
		case apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
			handlers[i] = aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
				FileResolver: signingCtx.FileResolver(),
				TypeResolver: opts.TypeResolver,
			})
		}
	}
	for i, m := range opts.CustomSignModes {
		handlers[i+lenSignModes] = m
	}

	handler := signing.NewHandlerMap(handlers...)
	return handler, nil
}
