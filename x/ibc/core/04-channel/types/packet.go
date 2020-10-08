package types

import (
	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// CommitPacket returns a packet commitment bytes. The commitment consists of:
// hash(timeout_timestamp + timeout_version_number + timeout_version_height + data) from a given packet.
func CommitPacket(packet exported.PacketI) []byte {
	buf := sdk.Uint64ToBigEndian(packet.GetTimeoutTimestamp())
	buf = append(buf, sdk.Uint64ToBigEndian(packet.GetTimeoutHeight().GetVersionNumber())...)
	buf = append(buf, sdk.Uint64ToBigEndian(packet.GetTimeoutHeight().GetVersionHeight())...)
	buf = append(buf, packet.GetData()...)
	return tmhash.Sum(buf)
}

// CommitAcknowledgement returns the hash of commitment bytes
func CommitAcknowledgement(data []byte) []byte {
	return tmhash.Sum(data)
}

var _ exported.PacketI = (*Packet)(nil)

// NewPacket creates a new Packet instance. It panics if the provided
// packet data interface is not registered.
func NewPacket(
	data []byte,
	sequence uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string,
	timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
) Packet {
	return Packet{
		Data:               data,
		Sequence:           sequence,
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    destinationPort,
		DestinationChannel: destinationChannel,
		TimeoutHeight:      timeoutHeight,
		TimeoutTimestamp:   timeoutTimestamp,
	}
}

// GetSequence implements PacketI interface
func (p Packet) GetSequence() uint64 { return p.Sequence }

// GetSourcePort implements PacketI interface
func (p Packet) GetSourcePort() string { return p.SourcePort }

// GetSourceChannel implements PacketI interface
func (p Packet) GetSourceChannel() string { return p.SourceChannel }

// GetDestPort implements PacketI interface
func (p Packet) GetDestPort() string { return p.DestinationPort }

// GetDestChannel implements PacketI interface
func (p Packet) GetDestChannel() string { return p.DestinationChannel }

// GetData implements PacketI interface
func (p Packet) GetData() []byte { return p.Data }

// GetTimeoutHeight implements PacketI interface
func (p Packet) GetTimeoutHeight() exported.Height { return p.TimeoutHeight }

// GetTimeoutTimestamp implements PacketI interface
func (p Packet) GetTimeoutTimestamp() uint64 { return p.TimeoutTimestamp }

// ValidateBasic implements PacketI interface
func (p Packet) ValidateBasic() error {
	if err := host.PortIdentifierValidator(p.SourcePort); err != nil {
		return sdkerrors.Wrap(err, "invalid source port ID")
	}
	if err := host.PortIdentifierValidator(p.DestinationPort); err != nil {
		return sdkerrors.Wrap(err, "invalid destination port ID")
	}
	if err := host.ChannelIdentifierValidator(p.SourceChannel); err != nil {
		return sdkerrors.Wrap(err, "invalid source channel ID")
	}
	if err := host.ChannelIdentifierValidator(p.DestinationChannel); err != nil {
		return sdkerrors.Wrap(err, "invalid destination channel ID")
	}
	if p.Sequence == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet sequence cannot be 0")
	}
	if p.TimeoutHeight.IsZero() && p.TimeoutTimestamp == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet timeout height and packet timeout timestamp cannot both be 0")
	}
	if len(p.Data) == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet data bytes cannot be empty")
	}
	return nil
}
