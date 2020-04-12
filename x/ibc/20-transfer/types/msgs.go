package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// MsgTransfer defines a msg to transfer fungible tokens (i.e Coins) between ICS20 enabled chains.
// See ICS Spec here: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#data-structures
type MsgTransfer struct {
	SourcePort    string         `json:"source_port" yaml:"source_port"`       // the port on which the packet will be sent
	SourceChannel string         `json:"source_channel" yaml:"source_channel"` // the channel by which the packet will be sent
	DestHeight    uint64         `json:"dest_height" yaml:"dest_height"`       // the current height of the destination chain
	Amount        sdk.Coins      `json:"amount" yaml:"amount"`                 // the tokens to be transferred
	Sender        sdk.AccAddress `json:"sender" yaml:"sender"`                 // the sender address
	Receiver      sdk.AccAddress `json:"receiver" yaml:"receiver"`             // the recipient address on the destination chain
}

// NewMsgTransfer creates a new MsgTransfer instance
func NewMsgTransfer(
	sourcePort, sourceChannel string, destHeight uint64, amount sdk.Coins, sender, receiver sdk.AccAddress,
) MsgTransfer {
	return MsgTransfer{
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		DestHeight:    destHeight,
		Amount:        amount,
		Sender:        sender,
		Receiver:      receiver,
	}
}

// Route implements sdk.Msg
func (MsgTransfer) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgTransfer) Type() string {
	return "transfer"
}

// ValidateBasic implements sdk.Msg
func (msg MsgTransfer) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(msg.SourcePort); err != nil {
		return sdkerrors.Wrap(err, "invalid source port ID")
	}
	if err := host.DefaultChannelIdentifierValidator(msg.SourceChannel); err != nil {
		return sdkerrors.Wrap(err, "invalid source channel ID")
	}
	if !msg.Amount.IsAllPositive() {
		return sdkerrors.ErrInsufficientFunds
	}
	if !msg.Amount.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing recipient address")
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
