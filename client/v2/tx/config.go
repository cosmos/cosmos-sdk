package tx

import (
	"errors"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/address"
	txdecode "cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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

// TxConfig is an interface that a client can use to generate a concrete transaction type
// defined by the application.
type TxConfig interface {
	TxEncodingConfig
	TxSigningConfig
}

// TxEncodingConfig defines the interface for transaction encoding and decoding.
// It provides methods for both binary and JSON encoding/decoding.
type TxEncodingConfig interface {
	// TxEncoder returns an encoder for binary transaction encoding.
	TxEncoder() txEncoder
	// TxDecoder returns a decoder for binary transaction decoding.
	TxDecoder() txDecoder
	// TxJSONEncoder returns an encoder for JSON transaction encoding.
	TxJSONEncoder() txEncoder
	// TxJSONDecoder returns a decoder for JSON transaction decoding.
	TxJSONDecoder() txDecoder
	// TxTextEncoder returns an encoder for text transaction encoding.
	TxTextEncoder() txEncoder
	// TxTextDecoder returns a decoder for text transaction decoding.
	TxTextDecoder() txDecoder
	// Decoder returns the Decoder interface for decoding transaction bytes into a DecodedTx.
	Decoder() Decoder
}

// TxSigningConfig defines the interface for transaction signing configurations.
type TxSigningConfig interface {
	// SignModeHandler returns a reference to the HandlerMap which manages the different signing modes.
	SignModeHandler() *signing.HandlerMap
	// SigningContext returns a reference to the Context which holds additional data required during signing.
	SigningContext() *signing.Context
	// MarshalSignatureJSON takes a slice of Signature objects and returns their JSON encoding.
	MarshalSignatureJSON([]Signature) ([]byte, error)
	// UnmarshalSignatureJSON takes a JSON byte slice and returns a slice of Signature objects.
	UnmarshalSignatureJSON([]byte) ([]Signature, error)
}

// ConfigOptions defines the configuration options for transaction processing.
type ConfigOptions struct {
	AddressCodec address.Codec
	Decoder      Decoder
	Cdc          codec.BinaryCodec

	ValidatorAddressCodec address.Codec
	FileResolver          signing.ProtoFileResolver
	TypeResolver          signing.TypeResolver
	CustomGetSigner       map[protoreflect.FullName]signing.GetSignersFunc
	MaxRecursionDepth     int

	EnabledSignModes           []apitxsigning.SignMode
	CustomSignModes            []signing.SignModeHandler
	TextualCoinMetadataQueryFn textual.CoinMetadataQueryFn
}

// validate checks the ConfigOptions for required fields and sets default values where necessary.
// It returns an error if any required field is missing.
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

	// set default signModes if none are provided
	if len(c.EnabledSignModes) == 0 {
		c.EnabledSignModes = defaultEnabledSignModes
	}
	return nil
}

// txConfig is a struct that embeds TxEncodingConfig and TxSigningConfig interfaces.
type txConfig struct {
	TxEncodingConfig
	TxSigningConfig
}

// NewTxConfig creates a new TxConfig instance using the provided ConfigOptions.
// It validates the options, initializes the signing context, and sets up the decoder if not provided.
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
			ProtoCodec:     options.Cdc,
		})
		if err != nil {
			return nil, err
		}
	}

	return &txConfig{
		TxEncodingConfig: defaultEncodingConfig{
			cdc:     options.Cdc,
			decoder: options.Decoder,
		},
		TxSigningConfig: signingCtx,
	}, nil
}

// defaultEncodingConfig is an empty struct that implements the TxEncodingConfig interface.
type defaultEncodingConfig struct {
	cdc     codec.BinaryCodec
	decoder Decoder
}

// TxEncoder returns the default transaction encoder.
func (t defaultEncodingConfig) TxEncoder() txEncoder {
	return encodeTx
}

// TxDecoder returns the default transaction decoder.
func (t defaultEncodingConfig) TxDecoder() txDecoder {
	return decodeTx(t.cdc, t.decoder)
}

// TxJSONEncoder returns the default JSON transaction encoder.
func (t defaultEncodingConfig) TxJSONEncoder() txEncoder {
	return encodeJsonTx
}

// TxJSONDecoder returns the default JSON transaction decoder.
func (t defaultEncodingConfig) TxJSONDecoder() txDecoder {
	return decodeJsonTx(t.cdc, t.decoder)
}

// TxTextEncoder returns the default text transaction encoder.
func (t defaultEncodingConfig) TxTextEncoder() txEncoder {
	return encodeTextTx
}

// TxTextDecoder returns the default text transaction decoder.
func (t defaultEncodingConfig) TxTextDecoder() txDecoder {
	return decodeTextTx(t.cdc, t.decoder)
}

// Decoder returns the Decoder instance associated with this encoding configuration.
func (t defaultEncodingConfig) Decoder() Decoder {
	return t.decoder
}

// defaultTxSigningConfig is a struct that holds the signing context and handler map.
type defaultTxSigningConfig struct {
	signingCtx *signing.Context
	handlerMap *signing.HandlerMap
	cdc        codec.BinaryCodec
}

// newDefaultTxSigningConfig creates a new defaultTxSigningConfig instance using the provided ConfigOptions.
// It initializes the signing context and handler map.
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
		cdc:        opts.Cdc,
	}, nil
}

// SignModeHandler returns the handler map that manages the different signing modes.
func (t defaultTxSigningConfig) SignModeHandler() *signing.HandlerMap {
	return t.handlerMap
}

// SigningContext returns the signing context that holds additional data required during signing.
func (t defaultTxSigningConfig) SigningContext() *signing.Context {
	return t.signingCtx
}

// MarshalSignatureJSON takes a slice of Signature objects and returns their JSON encoding.
// This method is not yet implemented and will panic if called.
func (t defaultTxSigningConfig) MarshalSignatureJSON(signatures []Signature) ([]byte, error) {
	descriptor := make([]*apitxsigning.SignatureDescriptor, len(signatures))

	for i, sig := range signatures {
		descData, err := signatureDataToProto(sig.Data)
		if err != nil {
			return nil, err
		}

		anyPk, err := codectypes.NewAnyWithValue(sig.PubKey)
		if err != nil {
			return nil, err
		}

		descriptor[i] = &apitxsigning.SignatureDescriptor{
			PublicKey: &anypb.Any{
				TypeUrl: codectypes.MsgTypeURL(sig.PubKey),
				Value:   anyPk.Value,
			},
			Data:     descData,
			Sequence: sig.Sequence,
		}
	}

	return jsonMarshalOptions.Marshal(&apitxsigning.SignatureDescriptors{Signatures: descriptor})
}

// UnmarshalSignatureJSON takes a JSON byte slice and returns a slice of Signature objects.
// This method is not yet implemented and will panic if called.
func (t defaultTxSigningConfig) UnmarshalSignatureJSON(bz []byte) ([]Signature, error) {
	var descriptor apitxsigning.SignatureDescriptors

	err := protojson.UnmarshalOptions{}.Unmarshal(bz, &descriptor)
	if err != nil {
		return nil, err
	}

	sigs := make([]Signature, len(descriptor.Signatures))
	for i, desc := range descriptor.Signatures {
		var pubkey cryptotypes.PubKey

		anyPk := &codectypes.Any{
			TypeUrl: desc.PublicKey.TypeUrl,
			Value:   desc.PublicKey.Value,
		}

		err = t.cdc.UnpackAny(anyPk, &pubkey)
		if err != nil {
			return nil, err
		}

		data, err := SignatureDataFromProto(desc.Data)
		if err != nil {
			return nil, err
		}

		sigs[i] = Signature{
			PubKey:   pubkey,
			Data:     data,
			Sequence: desc.Sequence,
		}
	}

	return sigs, nil
}

// newSigningContext creates a new signing context using the provided ConfigOptions.
// Returns a signing.Context instance or an error if initialization fails.
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

// newHandlerMap constructs a new HandlerMap based on the provided ConfigOptions and signing context.
// It initializes handlers for each enabled and custom sign mode specified in the options.
func newHandlerMap(opts ConfigOptions, signingCtx *signing.Context) (*signing.HandlerMap, error) {
	lenSignModes := len(opts.EnabledSignModes)
	handlers := make([]signing.SignModeHandler, lenSignModes+len(opts.CustomSignModes))

	for i, m := range opts.EnabledSignModes {
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
