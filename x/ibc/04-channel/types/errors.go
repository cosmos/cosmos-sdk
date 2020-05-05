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
	ErrInvalidPacket             = sdkerrors.Register(SubModuleName, 12, "invalid packet")
	ErrPacketTimeout             = sdkerrors.Register(SubModuleName, 13, "packet timeout")
	ErrTooManyConnectionHops     = sdkerrors.Register(SubModuleName, 14, "too many connection hops")
	ErrAcknowledgementTooLong    = sdkerrors.Register(SubModuleName, 15, "acknowledgement too long")
)
