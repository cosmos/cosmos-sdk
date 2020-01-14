package types

import (
	"encoding/binary"

	"github.com/tendermint/tendermint/crypto/tmhash"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// CommitPacket appends bigendian encoded timeout height and commitment bytes
// and then returns the hash of the result bytes.
// TODO: no specification for packet commitment currently,
// make it spec compatible once we have it
func CommitPacket(data exported.PacketDataI) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, data.GetTimeoutHeight())
	buf = append(buf, data.GetBytes()...)
	return tmhash.Sum(buf)
}

// CommitAcknowledgement returns the hash of commitment bytes
func CommitAcknowledgement(data exported.PacketDataI) []byte {
	return tmhash.Sum(data.GetBytes())
}

var _ exported.PacketI = Packet{}

// Packet defines a type that carries data across different chains through IBC
type Packet struct {
	Data exported.PacketDataI `json:"data" yaml:"data"` // opaque value which can be defined by the application logic of the associated modules.

	Sequence           uint64 `json:"sequence" yaml:"sequence"`                       // number corresponds to the order of sends and receives, where a Packet with an earlier sequence number must be sent and received before a Packet with a later sequence number.
	SourcePort         string `json:"source_port" yaml:"source_port"`                 // identifies the port on the sending chain.
	SourceChannel      string `json:"source_channel" yaml:"source_channel"`           // identifies the channel end on the sending chain.
	DestinationPort    string `json:"destination_port" yaml:"destination_port"`       // identifies the port on the receiving chain.
	DestinationChannel string `json:"destination_channel" yaml:"destination_channel"` // identifies the channel end on the receiving chain.
}

// NewPacket creates a new Packet instance
func NewPacket(
	data exported.PacketDataI,
	sequence uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string,
) Packet {
	return Packet{
		Data:               data,
		Sequence:           sequence,
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    destinationPort,
		DestinationChannel: destinationChannel,
	}
}

// Type exports Packet.Data.Type
func (p Packet) Type() string {
	return p.Data.Type()
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
func (p Packet) GetData() exported.PacketDataI { return p.Data }

// GetTimeoutHeight implements PacketI interface
func (p Packet) GetTimeoutHeight() uint64 { return p.Data.GetTimeoutHeight() }

// ValidateBasic implements PacketI interface
func (p Packet) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(p.SourcePort); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid source port ID: %s", p.SourcePort).Error(),
		)
	}
	if err := host.DefaultPortIdentifierValidator(p.DestinationPort); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid destination port ID: %s", p.DestinationPort).Error(),
		)
	}
	if err := host.DefaultChannelIdentifierValidator(p.SourceChannel); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid source channel ID: %s", p.SourceChannel).Error(),
		)
	}
	if err := host.DefaultChannelIdentifierValidator(p.DestinationChannel); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid destination channel ID: %s", p.DestinationChannel).Error(),
		)
	}
	if p.Sequence == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet sequence cannot be 0")
	}
	if p.Data.GetTimeoutHeight() == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet timeout cannot be 0")
	}
	if len(p.Data.GetBytes()) == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet data bytes cannot be empty")
	}
	return p.Data.ValidateBasic()
}
