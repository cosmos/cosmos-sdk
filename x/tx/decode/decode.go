package decode

import (
	"github.com/cosmos/cosmos-proto/anyutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	v1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/errors"
	"cosmossdk.io/x/tx/signing"
)

// DecodedTx contains the decoded transaction, its signers, and other flags.
type DecodedTx struct {
	Messages                     []proto.Message
	Tx                           *v1beta1.Tx
	TxRaw                        *v1beta1.TxRaw
	Signers                      []string
	TxBodyHasUnknownNonCriticals bool
}

// Decoder contains the dependencies required for decoding transactions.
type Decoder struct {
	getSignersCtx *signing.GetSignersContext
	typeResolver  protoregistry.MessageTypeResolver
	protoFiles    *protoregistry.Files
}

// Options are options for creating a Decoder.
type Options struct {
	// ProtoFiles are the protobuf files to use for resolving message descriptors.
	// If it is nil, the global protobuf registry will be used.
	ProtoFiles     *protoregistry.Files
	TypeResolver   protoregistry.MessageTypeResolver
	SigningContext *signing.GetSignersContext
}

// NewDecoder creates a new Decoder for decoding transactions.
func NewDecoder(options Options) (*Decoder, error) {
	if options.ProtoFiles == nil {
		options.ProtoFiles = protoregistry.GlobalFiles
	}

	if options.TypeResolver == nil {
		options.TypeResolver = protoregistry.GlobalTypes
	}

	getSignersCtx := options.SigningContext
	if getSignersCtx == nil {
		var err error
		getSignersCtx, err = signing.NewGetSignersContext(signing.GetSignersOptions{
			ProtoFiles: options.ProtoFiles,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Decoder{
		getSignersCtx: getSignersCtx,
		protoFiles:    options.ProtoFiles,
		typeResolver:  options.TypeResolver,
	}, nil
}

// Decode decodes raw protobuf encoded transaction bytes into a DecodedTx.
func (d *Decoder) Decode(txBytes []byte) (*DecodedTx, error) {
	// Make sure txBytes follow ADR-027.
	err := rejectNonADR027TxRaw(txBytes)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	var raw v1beta1.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err = RejectUnknownFieldsStrict(txBytes, raw.ProtoReflect().Descriptor(), d.protoFiles)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(txBytes, &raw)
	if err != nil {
		return nil, err
	}

	var body v1beta1.TxBody

	// allow non-critical unknown fields in TxBody
	txBodyHasUnknownNonCriticals, err := RejectUnknownFields(raw.BodyBytes, body.ProtoReflect().Descriptor(), true, d.protoFiles)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(raw.BodyBytes, &body)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	var authInfo v1beta1.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = RejectUnknownFieldsStrict(raw.AuthInfoBytes, authInfo.ProtoReflect().Descriptor(), d.protoFiles)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(raw.AuthInfoBytes, &authInfo)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	theTx := &v1beta1.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}

	var signers []string
	var msgs []proto.Message
	for _, anyMsg := range body.Messages {
		msg, signerErr := anyutil.Unpack(anyMsg, d.protoFiles, d.typeResolver)
		if signerErr != nil {
			return nil, errors.Wrap(ErrTxDecode, signerErr.Error())
		}
		msgs = append(msgs, msg)
		ss, signerErr := d.getSignersCtx.GetSigners(msg)
		if signerErr != nil {
			return nil, errors.Wrap(ErrTxDecode, signerErr.Error())
		}
		signers = append(signers, ss...)
	}

	return &DecodedTx{
		Messages:                     msgs,
		Tx:                           theTx,
		TxRaw:                        &raw,
		TxBodyHasUnknownNonCriticals: txBodyHasUnknownNonCriticals,
		Signers:                      signers,
	}, nil
}
