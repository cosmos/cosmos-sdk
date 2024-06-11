package decode

import (
	"crypto/sha256"
	"errors"

	"github.com/cosmos/cosmos-proto/anyutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"

	v1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/transaction"
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

	// Cache for hash and full bytes
	cachedHash   [32]byte
	cachedBytes  []byte
	cachedHashed bool
}

var _ transaction.Tx = &DecodedTx{}

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

// Hash implements the interface for the Tx interface.
func (dtx *DecodedTx) Hash() [32]byte {
	if !dtx.cachedHashed {
		dtx.computeHashAndBytes()
	}
	return dtx.cachedHash
}

func (dtx *DecodedTx) GetGasLimit() (uint64, error) {
	if dtx == nil || dtx.Tx == nil || dtx.Tx.AuthInfo == nil || dtx.Tx.AuthInfo.Fee == nil {
		return 0, errors.New("gas limit not available or one or more required fields are nil")
	}
	return dtx.Tx.AuthInfo.Fee.GasLimit, nil
}

func (dtx *DecodedTx) GetMessages() ([]transaction.Msg, error) {
	if dtx == nil || dtx.Messages == nil {
		return nil, errors.New("messages not available or are nil")
	}

	msgs := make([]transaction.Msg, len(dtx.Messages))
	for i, msg := range dtx.Messages {
		msgs[i] = protoadapt.MessageV1Of(msg)
	}

	return msgs, nil
}

func (dtx *DecodedTx) GetSenders() ([][]byte, error) {
	if dtx == nil || dtx.Signers == nil {
		return nil, errors.New("senders not available or are nil")
	}
	return dtx.Signers, nil
}

func (dtx *DecodedTx) Bytes() []byte {
	if !dtx.cachedHashed {
		dtx.computeHashAndBytes()
	}
	return dtx.cachedBytes
}

func (dtx *DecodedTx) computeHashAndBytes() {
	bz, err := proto.Marshal(dtx.TxRaw)
	if err != nil {
		panic(err)
	}

	dtx.cachedBytes = bz
	dtx.cachedHash = sha256.Sum256(bz)
	dtx.cachedHashed = true
}
