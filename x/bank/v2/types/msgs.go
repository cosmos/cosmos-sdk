package types

import (
	coretransaction "cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ coretransaction.Msg = &MsgSend{}

// NewMsgSend constructs a msg to send coins from one account to another.
func NewMsgSend(fromAddr, toAddr string, amount sdk.Coins) *MsgSend {
	return &MsgSend{FromAddress: fromAddr, ToAddress: toAddr, Amount: amount}
}

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}
