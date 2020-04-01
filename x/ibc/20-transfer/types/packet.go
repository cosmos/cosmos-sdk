package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// // FungibleTokenPacketData defines a struct for the packet payload
// // See FungibleTokenPacketData spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#data-structures
// type FungibleTokenPacketData struct {
// 	Amount   sdk.Coins      `json:"amount" yaml:"amount"`     // the tokens to be transferred
// 	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`     // the sender address
// 	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"` // the recipient address on the destination chain
// 	Source   bool           `json:"source" yaml:"source"`     // indicates if the sending chain is the source chain of the tokens to be transferred
// }

// NewFungibleTokenPacketData contructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(
	amount sdk.Coins, sender, receiver sdk.AccAddress,
	source bool) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
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
	if ftpd.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if ftpd.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing receiver address")
	}
	return nil
}

// GetBytes is a helper for serialising
func (ftpd FungibleTokenPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ftpd))
}

// GetBytes is a helper for serialising
func (AckDataTransfer) GetBytes() []byte {
	return []byte("fungible token transfer ack")
}
