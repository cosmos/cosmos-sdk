package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC channel sentinel errors
var (
	ErrChannelExists             = sdkerrors.Register(SubModuleName, 2, "channel already exists")
	ErrChannelNotFound           = sdkerrors.Register(SubModuleName, 3, "channel not found")
	ErrInvalidChannel            = sdkerrors.Register(SubModuleName, 4, "invalid channel")
	ErrInvalidChannelState       = sdkerrors.Register(SubModuleName, 5, "invalid channel state")
	ErrInvalidChannelOrdering    = sdkerrors.Register(SubModuleName, 6, "invalid channel ordering")
	ErrInvalidCounterparty       = sdkerrors.Register(SubModuleName, 7, "invalid counterparty channel")
	ErrInvalidChannelCapability  = sdkerrors.Register(SubModuleName, 8, "invalid channel capability")
	ErrChannelCapabilityNotFound = sdkerrors.Register(SubModuleName, 9, "channel capability not found")
	ErrSequenceSendNotFound      = sdkerrors.Register(SubModuleName, 10, "sequence send not found")
	ErrSequenceReceiveNotFound   = sdkerrors.Register(SubModuleName, 11, "sequence receive not found")
	ErrSequenceAckNotFound       = sdkerrors.Register(SubModuleName, 12, "sequence acknowledgement not found")
	ErrInvalidPacket             = sdkerrors.Register(SubModuleName, 13, "invalid packet")
	ErrPacketTimeout             = sdkerrors.Register(SubModuleName, 14, "packet timeout")
	ErrTooManyConnectionHops     = sdkerrors.Register(SubModuleName, 15, "too many connection hops")
	ErrInvalidAcknowledgement    = sdkerrors.Register(SubModuleName, 16, "invalid acknowledgement")
	ErrPacketCommitmentNotFound  = sdkerrors.Register(SubModuleName, 17, "packet commitment not found")
	ErrPacketReceived            = sdkerrors.Register(SubModuleName, 18, "packet already received")
	ErrAcknowledgementExists     = sdkerrors.Register(SubModuleName, 19, "acknowledgement for packet already exists")
	ErrInvalidChannelIdentifier  = sdkerrors.Register(SubModuleName, 20, "invalid channel identifier")

	// Antehandler error
	ErrRedundantTx = sdkerrors.Register(SubModuleName, 22, "packet messages are redundant")
)
