package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ exported.PacketI = Packet{}

// Packet defines a type that carries data across different chains through IBC
type Packet struct {
	Sequence           uint64 `json:"sequence" yaml:"sequence"`                       // number corresponds to the order of sends and receives, where a Packet with an earlier sequence number must be sent and received before a Packet with a later sequence number.
	Timeout            uint64 `json:"timeout" yaml:"timeout"`                         // indicates a consensus height on the destination chain after which the Packet will no longer be processed, and will instead count as having timed-out.
	SourcePort         string `json:"source_port" yaml:"source_port"`                 // identifies the port on the sending chain.
	SourceChannel      string `json:"source_channel" yaml:"source_channel"`           // identifies the channel end on the sending chain.
	DestinationPort    string `json:"destination_port" yaml:"destination_port"`       // identifies the port on the receiving chain.
	DestinationChannel string `json:"destination_channel" yaml:"destination_channel"` // identifies the channel end on the receiving chain.
	Data               []byte `json:"data" yaml:"data"`                               // opaque value which can be defined by the application logic of the associated modules.
}

// NewPacket creates a new Packet instance
func NewPacket(
	sequence, timeout uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string, data []byte,
) Packet {
	return Packet{
		Sequence:           sequence,
		Timeout:            timeout,
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    destinationPort,
		DestinationChannel: destinationChannel,
		Data:               data,
	}
}

// GetSequence implements PacketI interface
func (p Packet) GetSequence() uint64 { return p.Sequence }

// GetTimeoutHeight implements PacketI interface
func (p Packet) GetTimeoutHeight() uint64 { return p.Timeout }

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
	if p.Timeout == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet timeout cannot be 0")
	}
	if len(p.Data) == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet data cannot be empty")
	}
	return nil
}

var _ exported.PacketI = OpaquePacket{}

// OpaquePacket is a Packet, but cloaked in an obscuring data type by the host
// state machine, such that a module cannot act upon it other than to pass it to
// the IBC handler
type OpaquePacket struct {
	*Packet
}

// NewOpaquePacket creates a new OpaquePacket instance
func NewOpaquePacket(sequence, timeout uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string, data []byte,
) OpaquePacket {
	Packet := NewPacket(
		sequence, timeout, sourcePort, sourceChannel, destinationPort,
		destinationChannel, data,
	)
	return OpaquePacket{&Packet}
}

// GetData implements PacketI interface
func (op OpaquePacket) GetData() []byte { return nil }
