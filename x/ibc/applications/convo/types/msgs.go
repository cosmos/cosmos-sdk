package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

// msg types
const (
	TypeMsgConvo = "Convo"
)

// NewMsgConvo creates a new MsgConvo instance
//nolint:interfacer
func NewMsgConvo(
	sourcePort, sourceChannel string,
	sender sdk.AccAddress, receiver, message string,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
) *MsgConvo {
	return &MsgConvo{
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		Sender:           sender.String(),
		Receiver:         receiver,
		Message:          message,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
	}
}

// Route implements sdk.Msg
func (MsgConvo) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgConvo) Type() string {
	return TypeMsgConvo
}

// ValidateBasic performs a basic check of the MsgConvo fields.
// NOTE: timeout height or timestamp values can be 0 to disable the timeout.
func (msg MsgConvo) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.SourcePort); err != nil {
		return sdkerrors.Wrap(err, "invalid source port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.SourceChannel); err != nil {
		return sdkerrors.Wrap(err, "invalid source channel ID")
	}
	if msg.Sender == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if msg.Receiver == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing recipient address")
	}
	if msg.Message == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot send empty message")
	}
	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "message length too long, must send less than 100 characters. got %d", len(msg.Message))
}

// GetSignBytes implements sdk.Msg
func (msg MsgConvo) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners implements sdk.Msg
func (msg MsgConvo) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{valAddr}
}
