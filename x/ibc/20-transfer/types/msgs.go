package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

type MsgTransfer struct {
	SourcePort    string         `json:"source_port" yaml:"source_port"`
	SourceChannel string         `json:"source_channel" yaml:"source_channel"`
	Amount        sdk.Coins      `json:"amount" yaml:"amount"`
	Sender        sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver      sdk.AccAddress `json:"receiver" yaml:"receiver"`
	Source        bool           `json:"source" yaml:"source"`
}

// NewMsgTransfer creates a new MsgTransfer instance
func NewMsgTransfer(
	sourcePort, sourceChannel string, amount sdk.Coins, sender, receiver sdk.AccAddress, source bool,
) MsgTransfer {
	return MsgTransfer{
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Amount:        amount,
		Sender:        sender,
		Receiver:      receiver,
		Source:        source,
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
func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if err := host.DefaultIdentifierValidator(msg.SourcePort); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid port ID: %s", err.Error()))
	}
	if err := host.DefaultIdentifierValidator(msg.SourceChannel); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid channel ID: %s", err.Error()))
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins("transfer amount is invalid")
	}
	if !msg.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("transfer amount must be positive")
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if msg.Receiver.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
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
