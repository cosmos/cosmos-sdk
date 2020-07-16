package tx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

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

	if tx.GetGas() > authtypes.MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid gas supplied; %d > %d", tx.GetGas(), authtypes.MaxGasWanted,
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
