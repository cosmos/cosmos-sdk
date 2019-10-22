package types

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

var _ exported.PacketI = packet{}

// packet defines a type that carries data across different chains through IBC
type packet struct {
	sequence           uint64 // number corresponds to the order of sends and receives, where a packet with an earlier sequence number must be sent and received before a packet with a later sequence number.
	timeout            uint64 // indicates a consensus height on the destination chain after which the packet will no longer be processed, and will instead count as having timed-out.
	sourcePort         string // identifies the port on the sending chain.
	sourceChannel      string // identifies the channel end on the sending chain.
	destinationPort    string // identifies the port on the receiving chain.
	destinationChannel string // identifies the channel end on the receiving chain.
	data               []byte // opaque value which can be defined by the application logic of the associated modules.
}

// newPacket creates a new Packet instance
func newPacket(
	sequence, timeout uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string, data []byte,
) packet {
	return packet{
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
func (p packet) Sequence() uint64 { return p.sequence }

// TimeoutHeight implements PacketI interface
func (p packet) TimeoutHeight() uint64 { return p.timeout }

// SourcePort implements PacketI interface
func (p packet) SourcePort() string { return p.sourcePort }

// SourceChannel implements PacketI interface
func (p packet) SourceChannel() string { return p.sourceChannel }

// DestPort implements PacketI interface
func (p packet) DestPort() string { return p.destinationPort }

// DestChannel implements PacketI interface
func (p packet) DestChannel() string { return p.destinationChannel }

// Data implements PacketI interface
func (p packet) Data() []byte { return p.data }

// var _ exported.PacketI = OpaquePacket{}

// OpaquePacket is a packet, but cloaked in an obscuring data type by the host
// state machine, such that a module cannot act upon it other than to pass it to
// the IBC handler
type OpaquePacket packet

// TODO: Obscure data type
