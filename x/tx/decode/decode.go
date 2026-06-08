package decode

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-proto/anyutil"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"

	errorsmod "cosmossdk.io/errors"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// DecodedTx contains the decoded transaction, its signers, and other flags.
type DecodedTx struct {
	Messages                     []proto.Message
	Tx                           *txtypes.Tx
	TxRaw                        *txtypes.TxRaw
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

// msgDescriptor returns the protoreflect.MessageDescriptor for the named message
// from the global type registry. The gogoproto init() functions register all SDK
// proto types into protoregistry.GlobalTypes.
func msgDescriptor(fullName protoreflect.FullName) (protoreflect.MessageType, error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(fullName)
	if err != nil {
		return nil, fmt.Errorf("descriptor for %s not found in global registry: %w", fullName, err)
	}
	return mt, nil
}

// Decode decodes raw protobuf encoded transaction bytes into a DecodedTx.
func (d *Decoder) Decode(txBytes []byte) (*DecodedTx, error) {
	// Make sure txBytes follow ADR-027.
	if err := rejectNonADR027TxRaw(txBytes); err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	// Get descriptors from the global registry — gogoproto's proto.RegisterType
	// populates protoregistry.GlobalTypes, so no pulsar import is needed.
	txRawMT, err := msgDescriptor("cosmos.tx.v1beta1.TxRaw")
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}
	txBodyMT, err := msgDescriptor("cosmos.tx.v1beta1.TxBody")
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}
	authInfoMT, err := msgDescriptor("cosmos.tx.v1beta1.AuthInfo")
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	fileResolver := d.signingCtx.FileResolver()

	// reject all unknown proto fields in the root TxRaw
	if err = signing.RejectUnknownFieldsStrict(txBytes, txRawMT.Descriptor(), fileResolver); err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	var raw txtypes.TxRaw
	if err = gogoproto.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body txtypes.TxBody

	// allow non-critical unknown fields in TxBody
	txBodyHasUnknownNonCriticals, err := signing.RejectUnknownFields(raw.BodyBytes, txBodyMT.Descriptor(), true, fileResolver)
	if err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}
	if err = gogoproto.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	var authInfo txtypes.AuthInfo

	// reject all unknown proto fields in AuthInfo
	if err = signing.RejectUnknownFieldsStrict(raw.AuthInfoBytes, authInfoMT.Descriptor(), fileResolver); err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}
	if err = gogoproto.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
	}

	theTx := &txtypes.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}

	var signers [][]byte
	var msgs []proto.Message
	seenSigners := map[string]struct{}{}
	for _, gogoMsg := range body.Messages {
		// Convert gogoproto Any to protov2 Any for anyutil.Unpack.
		anyMsg := &anypb.Any{TypeUrl: gogoMsg.TypeUrl, Value: gogoMsg.Value}
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
			if _, seen := seenSigners[string(s)]; !seen {
				signers = append(signers, s)
				seenSigners[string(s)] = struct{}{}
			}
		}
	}

	// If a fee payer is specified in the AuthInfo, it must be added to the list of signers
	if authInfo.Fee != nil && authInfo.Fee.Payer != "" {
		feeAddr, err := d.signingCtx.AddressCodec().StringToBytes(authInfo.Fee.Payer)
		if err != nil {
			return nil, errorsmod.Wrap(ErrTxDecode, err.Error())
		}
		if _, seen := seenSigners[string(feeAddr)]; !seen {
			signers = append(signers, feeAddr)
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
