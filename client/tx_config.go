package client

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/address"
	txdecode "cosmossdk.io/x/tx/decode"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type (
	// TxEncodingConfig defines an interface that contains transaction
	// encoders and decoders
	TxEncodingConfig interface {
		TxEncoder() sdk.TxEncoder
		TxDecoder() sdk.TxDecoder
		TxJSONEncoder() sdk.TxEncoder
		TxJSONDecoder() sdk.TxDecoder
		MarshalSignatureJSON([]signingtypes.SignatureV2) ([]byte, error)
		UnmarshalSignatureJSON([]byte) ([]signingtypes.SignatureV2, error)
	}

	// TxConfig defines an interface a client can utilize to generate an
	// application-defined concrete transaction type. The type returned must
	// implement TxBuilder.
	TxConfig interface {
		TxEncodingConfig

		NewTxBuilder() TxBuilder
		WrapTxBuilder(sdk.Tx) (TxBuilder, error)
		SignModeHandler() *txsigning.HandlerMap
		SigningContext() *txsigning.Context
	}
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
//
// NOTE: Use NewTxConfigWithOptions to provide a custom signing handler in case the sign mode
// is not supported by default (eg: SignMode_SIGN_MODE_EIP_191), or to enable SIGN_MODE_TEXTUAL.
//
// We prefer to use depinject to provide client.TxConfig, but we permit this constructor usage. Within the SDK,
// this constructor is primarily used in tests, but also sees usage in app chains like:
// https://github.com/evmos/evmos/blob/719363fbb92ff3ea9649694bd088e4c6fe9c195f/encoding/config.go#L37
func NewTxConfig(protoCodec codec.Codec, addressCodec, validatorAddressCodec address.Codec, enabledSignModes []signingtypes.SignMode, customSignModes ...txsigning.SignModeHandler,
) TxConfig {
	txConfig, err := NewTxConfigWithOptions(protoCodec, ConfigOptions{
		EnabledSignModes: enabledSignModes,
		CustomSignModes:  customSignModes,
		SigningOptions: &txsigning.Options{
			AddressCodec:          addressCodec,
			ValidatorAddressCodec: validatorAddressCodec,
		},
	})
	if err != nil {
		panic(err)
	}
	return txConfig
}

// NewSigningOptions returns signing options used by x/tx. This includes account and
// validator address prefix enabled codecs.
func NewSigningOptions(addressCodec, validatorAddressCodec address.Codec) (*txsigning.Options, error) {
	return &txsigning.Options{
		AddressCodec:          addressCodec,
		ValidatorAddressCodec: validatorAddressCodec,
	}, nil
}

// NewSigningHandlerMap returns a new txsigning.HandlerMap using the provided ConfigOptions.
// It is recommended to use types.InterfaceRegistry in the field ConfigOptions.FileResolver as shown in
// NewTxConfigWithOptions but this fn does not enforce it.
func NewSigningHandlerMap(configOpts ConfigOptions) (*txsigning.HandlerMap, error) {
	var err error
	if configOpts.SigningOptions == nil {
		return nil, errors.New("signing options not provided")
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
func NewTxConfigWithOptions(protoCodec codec.Codec, configOptions ConfigOptions) (TxConfig, error) {
	txConfig := &config{
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

func (g config) NewTxBuilder() TxBuilder {
	return newBuilder(g.signingContext.AddressCodec(), g.txDecoder, g.protoCodec)
}

// WrapTxBuilder returns a builder from provided transaction
func (g config) WrapTxBuilder(newTx sdk.Tx) (TxBuilder, error) {
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

// DefaultJSONTxDecoder returns a default protobuf JSON TxDecoder using the provided Marshaler.
func DefaultJSONTxDecoder(addrCodec address.Codec, cdc codec.BinaryCodec, decoder *txdecode.Decoder) sdk.TxDecoder {
	jsonUnmarshaller := protojson.UnmarshalOptions{
		AllowPartial:   false,
		DiscardUnknown: false,
	}
	return func(txBytes []byte) (sdk.Tx, error) {
		jsonTx := new(txv1beta1.Tx)
		err := jsonUnmarshaller.Unmarshal(txBytes, jsonTx)
		if err != nil {
			return nil, err
		}

		// need to convert jsonTx into raw tx.
		bodyBytes, err := marshalOption.Marshal(jsonTx.Body)
		if err != nil {
			return nil, err
		}

		authInfoBytes, err := marshalOption.Marshal(jsonTx.AuthInfo)
		if err != nil {
			return nil, err
		}

		protoTxBytes, err := marshalOption.Marshal(&txv1beta1.TxRaw{
			BodyBytes:     bodyBytes,
			AuthInfoBytes: authInfoBytes,
			Signatures:    jsonTx.Signatures,
		})
		if err != nil {
			return nil, err
		}

		decodedTx, err := decoder.Decode(protoTxBytes)
		if err != nil {
			return nil, err
		}
		return newWrapperFromDecodedTx(addrCodec, cdc, decodedTx)
	}
}

// DefaultTxEncoder returns a default protobuf TxEncoder using the provided Marshaler
func DefaultTxEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		gogoWrapper, ok := tx.(*gogoTxWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected tx type: %T", tx)
		}
		return marshalOption.Marshal(gogoWrapper.TxRaw)
	}
}

// DefaultJSONTxEncoder returns a default protobuf JSON TxEncoder using the provided Marshaler.
func DefaultJSONTxEncoder(cdc codec.Codec) sdk.TxEncoder {
	jsonMarshaler := protojson.MarshalOptions{
		Indent:         "",
		UseProtoNames:  true,
		UseEnumNumbers: false,
	}
	return func(tx sdk.Tx) ([]byte, error) {
		gogoWrapper, ok := tx.(*gogoTxWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected tx type: %T", tx)
		}
		return jsonMarshaler.Marshal(gogoWrapper.Tx)
	}
}

func (g config) MarshalSignatureJSON(sigs []signingtypes.SignatureV2) ([]byte, error) {
	descs := make([]*signingtypes.SignatureDescriptor, len(sigs))

	for i, sig := range sigs {
		descData := signingtypes.SignatureDataToProto(sig.Data)
		any, err := codectypes.NewAnyWithValue(sig.PubKey)
		if err != nil {
			return nil, err
		}

		descs[i] = &signingtypes.SignatureDescriptor{
			PublicKey: any,
			Data:      descData,
			Sequence:  sig.Sequence,
		}
	}

	toJSON := &signingtypes.SignatureDescriptors{Signatures: descs}

	return codec.ProtoMarshalJSON(toJSON, nil)
}

func (g config) UnmarshalSignatureJSON(bz []byte) ([]signingtypes.SignatureV2, error) {
	var sigDescs signingtypes.SignatureDescriptors
	err := g.protoCodec.UnmarshalJSON(bz, &sigDescs)
	if err != nil {
		return nil, err
	}

	sigs := make([]signingtypes.SignatureV2, len(sigDescs.Signatures))
	for i, desc := range sigDescs.Signatures {
		pubKey, _ := desc.PublicKey.GetCachedValue().(cryptotypes.PubKey)

		data := signingtypes.SignatureDataFromProto(desc.Data)

		sigs[i] = signingtypes.SignatureV2{
			PubKey:   pubKey,
			Data:     data,
			Sequence: desc.Sequence,
		}
	}

	return sigs, nil
}
