package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewConvoPacketData constructs a new ConvoPacketData instance
func NewConvoPacketData(sender, receiver, message string) ConvoPacketData {
	return ConvoPacketData{
		Sender:   sender,
		Receiver: receiver,
		Message:  message,
	}
}

// ValidateBasic is used for validating the convo packet
func (cpd ConvoPacketData) ValidateBasic() error {
	if strings.TrimSpace(cpd.Sender) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be blank")
	}
	if strings.TrimSpace(cpd.Receiver) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be blank")
	}
	if cpd.Message == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot send empty message")
	}
	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "message length too long, must send less than 100 characters. got %d", len(cpd.Message))
}

// GetBytes is a helper for serialising
func (cpd ConvoPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&cpd))
}
