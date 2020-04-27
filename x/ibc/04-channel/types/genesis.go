package types

import (
	"errors"
	"fmt"

	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// PacketAcknowledgement defines the genesis type necesary to retrieve and store
// acknowlegements.
type PacketAcknowledgement struct {
	PortID    string `json:"port_id" yaml:"port_id"`
	ChannelID string `json:"channel_id" yaml:"channel_id"`
	Sequence  uint64 `json:"sequence" yaml:"sequence"`
	Ack       []byte `json:"ack" yaml:"ack"`
}

// NewPacketAcknowledgement creates a new PacketAcknowledgement instance.
func NewPacketAcknowledgement(portID, channelID string, seq uint64, ack []byte) PacketAcknowledgement {
	return PacketAcknowledgement{
		PortID:    portID,
		ChannelID: channelID,
		Sequence:  seq,
		Ack:       ack,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (pa PacketAcknowledgement) Validate() error {
	if len(pa.Ack) == 0 {
		return errors.New("acknowledgement bytes cannot be empty")
	}
	return validateGenFields(pa.PortID, pa.ChannelID, pa.Sequence)
}

// PacketCommitment defines the genesis type necesary to retrieve and store
// packet commitments.
type PacketCommitment struct {
	PortID    string `json:"port_id" yaml:"port_id"`
	ChannelID string `json:"channel_id" yaml:"channel_id"`
	Sequence  uint64 `json:"sequence" yaml:"sequence"`
	Hash      []byte `json:"hash" yaml:"hash"`
}

// NewPacketCommitment creates a new PacketCommitment instance.
func NewPacketCommitment(portID, channelID string, seq uint64, hash []byte) PacketCommitment {
	return PacketCommitment{
		PortID:    portID,
		ChannelID: channelID,
		Sequence:  seq,
		Hash:      hash,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (pc PacketCommitment) Validate() error {
	if len(pc.Hash) == 0 {
		return errors.New("hash bytes cannot be empty")
	}
	return validateGenFields(pc.PortID, pc.ChannelID, pc.Sequence)
}

// PacketSequence defines the genesis type necesary to retrieve and store
// next send and receive sequences.
type PacketSequence struct {
	PortID    string `json:"port_id" yaml:"port_id"`
	ChannelID string `json:"channel_id" yaml:"channel_id"`
	Sequence  uint64 `json:"sequence" yaml:"sequence"`
}

// NewPacketSequence creates a new PacketSequences instance.
func NewPacketSequence(portID, channelID string, seq uint64) PacketSequence {
	return PacketSequence{
		PortID:    portID,
		ChannelID: channelID,
		Sequence:  seq,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (ps PacketSequence) Validate() error {
	return validateGenFields(ps.PortID, ps.ChannelID, ps.Sequence)
}

// GenesisState defines the ibc channel submodule's genesis state.
type GenesisState struct {
	Channels         []Channel               `json:"channels" yaml:"channels"`
	Acknowledgements []PacketAcknowledgement `json:"acknowledgements" yaml:"acknowledgements"`
	Commitments      []PacketCommitment      `json:"commitments" yaml:"commitments"`
	SendSequences    []PacketSequence        `json:"send_sequences" yaml:"send_sequences"`
	RecvSequences    []PacketSequence        `json:"recv_sequences" yaml:"recv_sequences"`
}

// NewGenesisState creates a GenesisState instance.
func NewGenesisState(
	channels []Channel, acks []PacketAcknowledgement, commitments []PacketCommitment,
	sendSeqs, recvSeqs []PacketSequence,
) GenesisState {
	return GenesisState{
		Channels:         channels,
		Acknowledgements: acks,
		Commitments:      commitments,
		SendSequences:    sendSeqs,
		RecvSequences:    recvSeqs,
	}
}

// DefaultGenesisState returns the ibc channel submodule's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Channels:         []Channel{},
		Acknowledgements: []PacketAcknowledgement{},
		Commitments:      []PacketCommitment{},
		SendSequences:    []PacketSequence{},
		RecvSequences:    []PacketSequence{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for i, channel := range gs.Channels {
		if err := channel.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid channel %d: %w", i, err)
		}
	}

	for i, ack := range gs.Acknowledgements {
		if err := ack.Validate(); err != nil {
			return fmt.Errorf("invalid acknowledgement %d: %w", i, err)
		}
	}

	for i, commitment := range gs.Commitments {
		if err := commitment.Validate(); err != nil {
			return fmt.Errorf("invalid commitment %d: %w", i, err)
		}
	}

	for i, ss := range gs.SendSequences {
		if err := ss.Validate(); err != nil {
			return fmt.Errorf("invalid send sequence %d: %w", i, err)
		}
	}

	for i, rs := range gs.RecvSequences {
		if err := rs.Validate(); err != nil {
			return fmt.Errorf("invalid receive sequence %d: %w", i, err)
		}
	}

	return nil
}

func validateGenFields(portID, channelID string, sequence uint64) error {
	if err := host.DefaultPortIdentifierValidator(portID); err != nil {
		return err
	}
	if err := host.DefaultChannelIdentifierValidator(channelID); err != nil {
		return err
	}
	if sequence == 0 {
		return errors.New("sequence cannot be 0")
	}
	return nil
}
