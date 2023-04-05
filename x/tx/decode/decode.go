package decode

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	v1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/errors"
	"cosmossdk.io/x/tx/signing"
)

// DecodedTx contains the decoded transaction, its signers, and other flags.
type DecodedTx struct {
	Tx                           *v1beta1.Tx
	TxRaw                        *v1beta1.TxRaw
	Signers                      []string
	TxBodyHasUnknownNonCriticals bool
}

// Context contains the dependencies required for decoding transactions.
type Context struct {
	getSignersCtx *signing.GetSignersContext
	protoFiles    *protoregistry.Files
}

// Options are options for creating a Context.
type Options struct {
	// ProtoFiles are the protobuf files to use for resolving message descriptors.
	// If it is nil, the global protobuf registry will be used.
	ProtoFiles     *protoregistry.Files
	SigningContext *signing.GetSignersContext
}

// NewContext creates a new Context for decoding transactions.
func NewContext(options Options) (*Context, error) {
	protoFiles := options.ProtoFiles
	if protoFiles == nil {
		protoFiles = protoregistry.GlobalFiles
	}

	getSignersCtx := options.SigningContext
	if getSignersCtx == nil {
		var err error
		getSignersCtx, err = signing.NewGetSignersContext(signing.GetSignersOptions{
			ProtoFiles: protoFiles,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Context{
		getSignersCtx: getSignersCtx,
		protoFiles:    protoFiles,
	}, nil
}

// Decode decodes raw protobuf encoded transaction bytes into a DecodedTx.
func (c *Context) Decode(txBytes []byte) (*DecodedTx, error) {
	// Make sure txBytes follow ADR-027.
	err := rejectNonADR027TxRaw(txBytes)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	var raw v1beta1.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err = RejectUnknownFieldsStrict(txBytes, raw.ProtoReflect().Descriptor(), c.protoFiles)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(txBytes, &raw)
	if err != nil {
		return nil, err
	}

	var body v1beta1.TxBody

	// allow non-critical unknown fields in TxBody
	txBodyHasUnknownNonCriticals, err := RejectUnknownFields(raw.BodyBytes, body.ProtoReflect().Descriptor(), true, c.protoFiles)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(raw.BodyBytes, &body)
	if err != nil {
		return nil, errors.Wrap(ErrTxDecode, err.Error())
	}

	var authInfo v1beta1.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = RejectUnknownFieldsStrict(raw.AuthInfoBytes, authInfo.ProtoReflect().Descriptor(), c.protoFiles)
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

	signers, err := c.getSignersCtx.GetSigners(theTx)
	if err != nil {
		return nil, err
	}

	return &DecodedTx{
		Tx:                           theTx,
		TxRaw:                        &raw,
		TxBodyHasUnknownNonCriticals: txBodyHasUnknownNonCriticals,
		Signers:                      signers,
	}, nil
}
