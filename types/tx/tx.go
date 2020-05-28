package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type ProtoTx interface {
	sdk.Tx
	FeeTx
	TxWithMemo
	SigTx

	GetBody() *TxBody
	GetAuthInfo() *AuthInfo
	GetSignatures() [][]byte

	GetBodyBytes() []byte
	GetAuthInfoBytes() []byte
}

var _ ProtoTx = &Tx{}

func NewTx() *Tx {
	return &Tx{
		Body:       &TxBody{},
		AuthInfo:   &AuthInfo{},
		Signatures: nil,
	}
}

func (tx *Tx) GetMsgs() []sdk.Msg {
	anys := tx.Body.Messages
	res := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		msg := any.GetCachedValue().(sdk.Msg)
		res[i] = msg
	}
	return res
}

func (tx *Tx) ValidateBasic() error {
	sigs := tx.GetSignatures()

	if tx.GetGas() > MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid gas supplied; %d > %d", tx.GetGas(), MaxGasWanted,
		)
	}
	if tx.GetFee().IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee,
			"invalid fee provided: %s", tx.GetFee(),
		)
	}
	if len(sigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}
	if len(sigs) != len(tx.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"wrong number of signers; expected %d, got %d", tx.GetSigners(), len(sigs),
		)
	}

	return nil
}

func (m *Tx) GetGas() uint64 {
	return m.AuthInfo.Fee.GasLimit
}

func (m *Tx) GetFee() sdk.Coins {
	return m.AuthInfo.Fee.Amount
}

func (m *Tx) FeePayer() sdk.AccAddress {
	signers := m.GetSigners()
	if signers != nil {
		return signers[0]
	}
	return sdk.AccAddress{}
}

func (m *Tx) GetMemo() string {
	return m.Body.Memo
}

func (m *Tx) GetSigners() []sdk.AccAddress {
	var signers []sdk.AccAddress
	seen := map[string]bool{}

	for _, msg := range m.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}
	return signers
}

func (m *Tx) GetPubKeys() []crypto.PubKey {
	signerInfos := m.AuthInfo.SignerInfos
	res := make([]crypto.PubKey, len(signerInfos))
	for i, si := range signerInfos {
		res[i] = si.PublicKey.GetCachedPubKey()
	}
	return res
}

func (m *Tx) GetBodyBytes() []byte {
	bz, err := m.Body.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

func (m *Tx) GetAuthInfoBytes() []byte {
	bz, err := m.AuthInfo.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

var _ codectypes.UnpackInterfacesMessage = &Tx{}
var _ codectypes.UnpackInterfacesMessage = &TxBody{}
var _ codectypes.UnpackInterfacesMessage = &SignDoc{}

func (m *Tx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Body.UnpackInterfaces(unpacker)
}

func (m *SignDoc) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Body.UnpackInterfaces(unpacker)
}

func (m *TxBody) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, any := range m.Messages {
		var msg sdk.Msg
		err := unpacker.UnpackAny(any, &msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// FeeTx defines the interface to be implemented by Tx to use the FeeDecorators
type FeeTx interface {
	sdk.Tx
	GetGas() uint64
	GetFee() sdk.Coins
	FeePayer() sdk.AccAddress
}

// Tx must have GetMemo() method to use ValidateMemoDecorator
type TxWithMemo interface {
	sdk.Tx
	GetMemo() string
}

type SigTx interface {
	sdk.Tx
	GetSignatures() [][]byte
	GetSigners() []sdk.AccAddress
	GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place
	GetSignaturesV2() ([]SignatureV2, error)
}

type SignatureV2 interface {
	isSignatureV2()
}

type SingleSignature struct {
	SignMode  SignMode
	Signature []byte
}

type MultiSignature struct {
	BitArray   *types.CompactBitArray
	Signatures []SignatureV2
}

var _, _ SignatureV2 = SingleSignature{}, MultiSignature{}

func (m SingleSignature) isSignatureV2() {}
func (m MultiSignature) isSignatureV2()  {}

func ModeInfoToSignatureV2(modeInfo *ModeInfo, sig []byte) (SignatureV2, error) {
	switch modeInfo := modeInfo.Sum.(type) {
	case *ModeInfo_Single_:
		return SingleSignature{
			SignMode:  modeInfo.Single.Mode,
			Signature: sig,
		}, nil

	case *ModeInfo_Multi_:
		multi := modeInfo.Multi

		sigs, err := types.DecodeMultisignatures(sig)
		if err != nil {
			return nil, err
		}

		sigv2s := make([]SignatureV2, len(sigs))
		for i, mi := range multi.ModeInfos {
			sigv2s[i], err = ModeInfoToSignatureV2(mi, sigs[i])
			if err != nil {
				return nil, err
			}
		}

		return MultiSignature{
			BitArray:   multi.Bitarray,
			Signatures: sigv2s,
		}, nil

	default:
		panic("unexpected case")
	}
}

func (m *Tx) GetSignaturesV2() ([]SignatureV2, error) {
	signerInfos := m.AuthInfo.SignerInfos
	sigs := m.Signatures
	n := len(signerInfos)
	res := make([]SignatureV2, n)

	for i, si := range signerInfos {
		var err error
		res[i], err = ModeInfoToSignatureV2(si.ModeInfo, sigs[i])
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
