package std

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var _ sdk.Tx = Transaction{}

// GetMsgs returns all the messages in a Transaction as a slice of sdk.Msg.
func (tx Transaction) GetMsgs() []sdk.Msg {
	msgs := make([]sdk.Msg, len(tx.Msgs))

	for i, m := range tx.Msgs {
		msgs[i] = m.GetMsg()
	}

	return msgs
}

// GetSigners returns the addresses that must sign the transaction. Addresses are
// returned in a deterministic order. They are accumulated from the GetSigners
// method for each Msg in the order they appear in tx.GetMsgs(). Duplicate addresses
// will be omitted.
func (tx Transaction) GetSigners() []sdk.AccAddress {
	var signers []sdk.AccAddress
	seen := map[string]bool{}

	for _, msg := range tx.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}

	return signers
}

// ValidateBasic does a simple and lightweight validation check that doesn't
// require access to any other information.
func (tx Transaction) ValidateBasic() error {
	stdSigs := tx.Base.GetSignatures()

	if tx.Base.Fee.Gas > auth.MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid gas supplied; %d > %d", tx.Base.Fee.Gas, auth.MaxGasWanted,
		)
	}
	if tx.Base.Fee.Amount.IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee, "invalid fee provided: %s", tx.Base.Fee.Amount,
		)
	}
	if len(stdSigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}
	if len(stdSigs) != len(tx.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized, "wrong number of signers; expected %d, got %d", tx.GetSigners(), len(stdSigs),
		)
	}

	return nil
}
