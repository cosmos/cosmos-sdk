package tx

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type txWrapper struct {
	tx          *Tx
	bodyBz      []byte
	authInfoBz  []byte
	pubKeys     []crypto.PubKey
	marshaler   codec.Marshaler
	pubkeyCodec types.PublicKeyCodec
}

type TxWrapper interface {
	sdk.Tx
	ProtoTx

	WithTxBody(*TxBody) TxWrapper
	WithAuthInfo(*AuthInfo) TxWrapper
	WithSignatures([][]byte) TxWrapper
}

func NewTxWrapper(marshaler codec.Marshaler, pubkeyCodec types.PublicKeyCodec) TxWrapper {
	return txWrapper{
		tx: &Tx{
			Body:     &TxBody{},
			AuthInfo: &AuthInfo{},
		},
		marshaler:   marshaler,
		pubkeyCodec: pubkeyCodec,
	}
}

var (
	_ sdk.Tx  = txWrapper{}
	_ ProtoTx = txWrapper{}
)

func (t txWrapper) GetMsgs() []sdk.Msg {
	anys := t.tx.Body.Messages
	res := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		msg := any.GetCachedValue().(sdk.Msg)
		res[i] = msg
	}
	return res
}

// MaxGasWanted defines the max gas allowed.
const MaxGasWanted = uint64((1 << 63) - 1)

func (t txWrapper) ValidateBasic() error {
	tx := t.tx
	if tx == nil {
		return fmt.Errorf("bad Tx")
	}

	body := t.tx.Body
	if body == nil {
		return fmt.Errorf("missing TxBody")
	}

	authInfo := t.tx.AuthInfo
	if authInfo == nil {
		return fmt.Errorf("missing AuthInfo")
	}

	fee := authInfo.Fee
	if fee == nil {
		return fmt.Errorf("missing fee")
	}

	if fee.GasLimit > MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid gas supplied; %d > %d", fee.GasLimit, MaxGasWanted,
		)
	}

	if fee.Amount.IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee,
			"invalid fee provided: %s", fee.Amount,
		)
	}

	sigs := tx.Signatures

	if len(sigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}

	if len(sigs) != len(t.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"wrong number of signers; expected %d, got %d", t.GetSigners(), len(sigs),
		)
	}

	return nil
}

func (t txWrapper) GetBodyBytes() []byte {
	return t.bodyBz
}

func (t txWrapper) GetAuthInfoBytes() []byte {
	return t.authInfoBz
}

func (t txWrapper) GetSigners() []sdk.AccAddress {
	var signers []sdk.AccAddress
	seen := map[string]bool{}

	for _, msg := range t.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}

	return signers
}

func (t txWrapper) WithTxBody(body *TxBody) TxWrapper {
	t.tx.Body = body
	t.bodyBz = t.marshaler.MustMarshalBinaryBare(body)
	return t
}

func (t txWrapper) WithAuthInfo(info *AuthInfo) TxWrapper {
	t.tx.AuthInfo = info
	t.authInfoBz = t.marshaler.MustMarshalBinaryBare(info)
	return t
}

func (t txWrapper) WithSignatures(sigs [][]byte) TxWrapper {
	t.tx.Signatures = sigs
	return t
}
