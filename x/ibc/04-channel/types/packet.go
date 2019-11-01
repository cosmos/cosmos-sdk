package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ exported.PacketI = Packet{}

// Packet defines a type that carries data across different chains through IBC
type Packet struct {
	sequence           uint64 `json:"sequence"`            // number corresponds to the order of sends and receives, where a Packet with an earlier sequence number must be sent and received before a Packet with a later sequence number.
	timeout            uint64 `json:"timeout"`             // indicates a consensus height on the destination chain after which the Packet will no longer be processed, and will instead count as having timed-out.
	sourcePort         string `json:"source_port"`         // identifies the port on the sending chain.
	sourceChannel      string `json:"source_channel"`      // identifies the channel end on the sending chain.
	destinationPort    string `json:"destination_port"`    // identifies the port on the receiving chain.
	destinationChannel string `json:"destination_channel"` // identifies the channel end on the receiving chain.
	data               []byte `json:"data"`                // opaque value which can be defined by the application logic of the associated modules.
}

// NewPacket creates a new Packet instance
func NewPacket(
	sequence, timeout uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string, data []byte,
) Packet {
	return Packet{
		sequence,
		timeout,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		data,
	}
}

// Sequence implements PacketI interface
func (p Packet) Sequence() uint64 { return p.sequence }

// TimeoutHeight implements PacketI interface
func (p Packet) TimeoutHeight() uint64 { return p.timeout }

// SourcePort implements PacketI interface
func (p Packet) SourcePort() string { return p.sourcePort }

// SourceChannel implements PacketI interface
func (p Packet) SourceChannel() string { return p.sourceChannel }

// DestPort implements PacketI interface
func (p Packet) DestPort() string { return p.destinationPort }

// DestChannel implements PacketI interface
func (p Packet) DestChannel() string { return p.destinationChannel }

// Data implements PacketI interface
func (p Packet) Data() []byte { return p.data }

// ValidateBasic implements PacketI interface
func (p Packet) ValidateBasic() sdk.Error {
	if err := host.DefaultPortIdentifierValidator(p.sourcePort); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid source port ID: %s", err.Error()))
	}
	if err := host.DefaultPortIdentifierValidator(p.destinationPort); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid destination port ID: %s", err.Error()))
	}
	if err := host.DefaultChannelIdentifierValidator(p.sourceChannel); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid source channel ID: %s", err.Error()))
	}
	if err := host.DefaultChannelIdentifierValidator(p.destinationChannel); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid destination channel ID: %s", err.Error()))
	}
	if p.sequence == 0 {
		return ErrInvalidPacket(DefaultCodespace, "packet sequence cannot be 0")
	}
	if p.timeout == 0 {
		return ErrPacketTimeout(DefaultCodespace)
	}
	if len(p.data) == 0 {
		return ErrInvalidPacket(DefaultCodespace, "packet data cannot be empty")
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

// Data implements PacketI interface
func (op OpaquePacket) Data() []byte { return nil }
