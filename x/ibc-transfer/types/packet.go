package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewFungibleTokenPacketData contructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(
	amount sdk.Coins, sender, receiver string) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
	}
}

// ValidateBasic is used for validating the token transfer
func (ftpd FungibleTokenPacketData) ValidateBasic() error {
	if !ftpd.Amount.IsAllPositive() {
		return sdkerrors.ErrInsufficientFunds
	}
	if !ftpd.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	if ftpd.Sender == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if ftpd.Receiver == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing receiver address")
	}
	return nil
}

// GetBytes is a helper for serialising
func (ftpd FungibleTokenPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ftpd))
}

// GetBytes is a helper for serialising
func (ack FungibleTokenPacketAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ack))
}
