package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

var _ channelexported.PacketDataI = FungibleTokenPacketData{}

// NewFungibleTokenPacketData contructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(
	amount sdk.Coins, sender, receiver sdk.AccAddress,
	source bool, timeout uint64) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
		Timeout:  timeout,
	}
}

// ValidateBasic implements channelexported.PacketDataI
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
	if ftpd.Timeout == 0 {
		return sdkerrors.Wrap(ErrInvalidPacketTimeout, "timeout cannot be 0")
	}
	return nil
}

// GetBytes implements channelexported.PacketDataI
func (ftpd FungibleTokenPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ftpd))
}

// GetTimeoutHeight implements channelexported.PacketDataI
func (ftpd FungibleTokenPacketData) GetTimeoutHeight() uint64 {
	return ftpd.Timeout
}

// Type implements channelexported.PacketDataI
func (ftpd FungibleTokenPacketData) Type() string {
	return "ics20/transfer"
}

var _ channelexported.PacketAcknowledgementI = AckDataTransfer{}

// GetBytes implements channelexported.PacketAcknowledgementI
func (ack AckDataTransfer) GetBytes() []byte {
	return []byte("fungible token transfer ack")
}
