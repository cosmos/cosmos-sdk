package decode

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-proto/anyutil"
	"google.golang.org/protobuf/proto"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/tx/signing"
)

// DecodedTx contains the decoded transaction, its signers, and other flags.
type DecodedTx struct {
	Messages                     []proto.Message
	Tx                           *v1beta1.Tx
	TxRaw                        *v1beta1.TxRaw
	Signers                      [][]byte
	TxBodyHasUnknownNonCriticals bool
}

// Decoder contains the dependencies required for decoding transactions.
type Decoder struct {
	signingCtx *signing.Context
}

// Options are options for creating a Decoder.
type Options struct {
	SigningContext *signing.Context
}

// NewDecoder creates a new Decoder for decoding transactions.
func NewDecoder(options Options) (*Decoder, error) {
	if options.SigningContext == nil {
		return nil, errors.New("signing context is required")
	}

	return &Decoder{
		signingCtx: options.SigningContext,
	}, nil
}

// Decode decodes raw protobuf encoded transaction bytes into a DecodedTx.
func (d *Decoder) Decode(txBytes []byte) (*DecodedTx, error) {
	// Make sure txBytes follow ADR-027.
	err := rejectNonADR027TxRaw(txBytes)
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	var raw v1beta1.TxRaw

	// reject all unknown proto fields in the root TxRaw
	fileResolver := d.signingCtx.FileResolver()
	err = RejectUnknownFieldsStrict(txBytes, raw.ProtoReflect().Descriptor(), fileResolver)
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(txBytes, &raw)
	if err != nil {
		return nil, err
	}

	var body v1beta1.TxBody

	// allow non-critical unknown fields in TxBody
	txBodyHasUnknownNonCriticals, err := RejectUnknownFields(raw.BodyBytes, body.ProtoReflect().Descriptor(), true, fileResolver)
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(raw.BodyBytes, &body)
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	var authInfo v1beta1.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = RejectUnknownFieldsStrict(raw.AuthInfoBytes, authInfo.ProtoReflect().Descriptor(), fileResolver)
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	err = proto.Unmarshal(raw.AuthInfoBytes, &authInfo)
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	if err := rejectAminoUnorderedTx(&authInfo, &body); err != nil {
		return nil, err
	}

	theTx := &v1beta1.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}

	var signers [][]byte
	var msgs []proto.Message
	seenSigners := map[string]struct{}{}
	for _, anyMsg := range body.Messages {
		msg, signerErr := anyutil.Unpack(anyMsg, fileResolver, d.signingCtx.TypeResolver())
		if signerErr != nil {
			return nil, errorsmod.Wrap(ErrTxDecode, signerErr.Error())
		}
		msgs = append(msgs, msg)
		ss, signerErr := d.signingCtx.GetSigners(msg)
		if signerErr != nil {
			return nil, errorsmod.Wrap(ErrTxDecode, signerErr.Error())
		}
		for _, s := range ss {
			_, seen := seenSigners[string(s)]
			if seen {
				continue
			}
			signers = append(signers, s)
			seenSigners[string(s)] = struct{}{}
		}
	}

	return &DecodedTx{
		Messages:                     msgs,
		Tx:                           theTx,
		TxRaw:                        &raw,
		TxBodyHasUnknownNonCriticals: txBodyHasUnknownNonCriticals,
		Signers:                      signers,
	}, nil
}

// rejectAminoUnorderedTx checks if the given transaction should be rejected
// based on the provided authentication information and transaction body. It
// ensures the following:
//
// - The transaction must be unordered
// - Each SignerInfo in the transaction must be of type ModeInfo_Single (i.e. multisigs are not allowed)
// - The mode of each SignerInfo must NOT be SIGN_MODE_LEGACY_AMINO_JSON
//
// We return <nil> if the transaction is valid, otherwise an error is returned.
func rejectAminoUnorderedTx(authInfo *v1beta1.AuthInfo, body *v1beta1.TxBody) error {
	if !body.Unordered {
		return nil
	}

	for _, info := range authInfo.SignerInfos {
		single, ok := info.ModeInfo.Sum.(*v1beta1.ModeInfo_Single_)
		if !ok {
			return fmt.Errorf("unexpected mode info: %T; must be %T", info.ModeInfo.Sum, &v1beta1.ModeInfo_Single_{})
		}

		if single.Single.Mode == signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
			return errors.New("signing unordered txs with amino is prohibited")
		}
	}

	return nil
}
