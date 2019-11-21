package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// PacketDataI defines the standard interface for IBC packet data
type PacketDataI interface {
	GetCommitment() []byte // Commitment form that will be stored in the state.
	GetTimeoutHeight() uint64

	ValidateBasic() sdk.Error
	Type() string
}

// Packet defines a type that carries data across different chains through IBC
type Packet struct {
	Sequence           uint64      `json:"sequence"`            // number corresponds to the order of sends and receives, where a Packet with an earlier sequence number must be sent and received before a Packet with a later sequence number.
	Timeout            uint64      `json:"timeout"`             // indicates a consensus height on the destination chain after which the Packet will no longer be processed, and will instead count as having timed-out.
	SourcePort         string      `json:"source_port"`         // identifies the port on the sending chain.
	SourceChannel      string      `json:"source_channel"`      // identifies the channel end on the sending chain.
	DestinationPort    string      `json:"destination_port"`    // identifies the port on the receiving chain.
	DestinationChannel string      `json:"destination_channel"` // identifies the channel end on the receiving chain.
	Data               PacketDataI `json:"data"`                // opaque value which can be defined by the application logic of the associated modules.
}

// NewPacket creates a new Packet instance
func NewPacket(
	sequence, timeout uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string, data PacketDataI,
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

// GetSequence implements PacketDataI interface
func (p Packet) GetSequence() uint64 { return p.Sequence }

// GetTimeoutHeight implements PacketDataI interface
func (p Packet) GetTimeoutHeight() uint64 { return p.Timeout }

// GetSourcePort implements PacketDataI interface
func (p Packet) GetSourcePort() string { return p.SourcePort }

// GetSourceChannel implements PacketDataI interface
func (p Packet) GetSourceChannel() string { return p.SourceChannel }

// GetDestPort implements PacketDataI interface
func (p Packet) GetDestPort() string { return p.DestinationPort }

// GetDestChannel implements PacketDataI interface
func (p Packet) GetDestChannel() string { return p.DestinationChannel }

// ValidateBasic implements PacketDataI interface
func (p Packet) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(p.SourcePort); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid source port ID: %s", err.Error()))
	}
	if err := host.DefaultPortIdentifierValidator(p.DestinationPort); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid destination port ID: %s", err.Error()))
	}
	if err := host.DefaultChannelIdentifierValidator(p.SourceChannel); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid source channel ID: %s", err.Error()))
	}
	if err := host.DefaultChannelIdentifierValidator(p.DestinationChannel); err != nil {
		return ErrInvalidPacket(DefaultCodespace, fmt.Sprintf("invalid destination channel ID: %s", err.Error()))
	}
	if p.Sequence == 0 {
		return ErrInvalidPacket(DefaultCodespace, "packet sequence cannot be 0")
	}
	if p.Timeout == 0 {
		return ErrPacketTimeout(DefaultCodespace)
	}
	if len(p.Data.GetCommitment()) == 0 {
		return ErrInvalidPacket(DefaultCodespace, "packet commitment cannot be empty")
	}
	return nil
}

// XXX: fix before merge
/*
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

// GetData implements PacketDataI interface
func (op OpaquePacket) GetData() []byte { return nil }
*/
