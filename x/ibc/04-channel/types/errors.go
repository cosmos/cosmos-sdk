package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC channel sentinel errors
var (
	ErrChannelExists             = sdkerrors.Register(SubModuleName, 1, "channel already exists")
	ErrChannelNotFound           = sdkerrors.Register(SubModuleName, 2, "channel not found")
	ErrInvalidChannel            = sdkerrors.Register(SubModuleName, 3, "invalid channel")
	ErrInvalidChannelState       = sdkerrors.Register(SubModuleName, 4, "invalid channel state")
	ErrInvalidChannelOrdering    = sdkerrors.Register(SubModuleName, 5, "invalid channel ordering")
	ErrInvalidCounterparty       = sdkerrors.Register(SubModuleName, 6, "invalid counterparty channel")
	ErrInvalidChannelCapability  = sdkerrors.Register(SubModuleName, 7, "invalid channel capability")
	ErrChannelCapabilityNotFound = sdkerrors.Register(SubModuleName, 8, "channel capability not found")
	ErrSequenceSendNotFound      = sdkerrors.Register(SubModuleName, 9, "sequence send not found")
	ErrSequenceReceiveNotFound   = sdkerrors.Register(SubModuleName, 10, "sequence receive not found")
	ErrInvalidPacket             = sdkerrors.Register(SubModuleName, 11, "invalid packet")
	ErrPacketTimeout             = sdkerrors.Register(SubModuleName, 12, "packet timeout")
	ErrTooManyConnectionHops     = sdkerrors.Register(SubModuleName, 13, "too many connection hops")
	ErrAcknowledgementTooLong    = sdkerrors.Register(SubModuleName, 14, "acknowledgement too long")
)
