package types

import (
	"errors"
	"fmt"

	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// PacketAckCommitment defines the genesis type necessary to retrieve and store
// acknowlegements.
type PacketAckCommitment struct {
	PortID    string `json:"port_id" yaml:"port_id"`
	ChannelID string `json:"channel_id" yaml:"channel_id"`
	Sequence  uint64 `json:"sequence" yaml:"sequence"`
	Hash      []byte `json:"hash" yaml:"hash"`
}

// NewPacketAckCommitment creates a new PacketAckCommitment instance.
func NewPacketAckCommitment(portID, channelID string, seq uint64, hash []byte) PacketAckCommitment {
	return PacketAckCommitment{
		PortID:    portID,
		ChannelID: channelID,
		Sequence:  seq,
		Hash:      hash,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (pa PacketAckCommitment) Validate() error {
	if len(pa.Hash) == 0 {
		return errors.New("hash bytes cannot be empty")
	}
	return validateGenFields(pa.PortID, pa.ChannelID, pa.Sequence)
}

// PacketSequence defines the genesis type necessary to retrieve and store
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
	Channels         []IdentifiedChannel   `json:"channels" yaml:"channels"`
	Acknowledgements []PacketAckCommitment `json:"acknowledgements" yaml:"acknowledgements"`
	Commitments      []PacketAckCommitment `json:"commitments" yaml:"commitments"`
	SendSequences    []PacketSequence      `json:"send_sequences" yaml:"send_sequences"`
	RecvSequences    []PacketSequence      `json:"recv_sequences" yaml:"recv_sequences"`
	AckSequences     []PacketSequence      `json:"ack_sequences" yaml:"ack_sequences"`
}

// NewGenesisState creates a GenesisState instance.
func NewGenesisState(
	channels []IdentifiedChannel, acks, commitments []PacketAckCommitment,
	sendSeqs, recvSeqs, ackSeqs []PacketSequence,
) GenesisState {
	return GenesisState{
		Channels:         channels,
		Acknowledgements: acks,
		Commitments:      commitments,
		SendSequences:    sendSeqs,
		RecvSequences:    recvSeqs,
		AckSequences:     ackSeqs,
	}
}

// DefaultGenesisState returns the ibc channel submodule's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Channels:         []IdentifiedChannel{},
		Acknowledgements: []PacketAckCommitment{},
		Commitments:      []PacketAckCommitment{},
		SendSequences:    []PacketSequence{},
		RecvSequences:    []PacketSequence{},
		AckSequences:     []PacketSequence{},
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

	for i, as := range gs.AckSequences {
		if err := as.Validate(); err != nil {
			return fmt.Errorf("invalid acknowledgement sequence %d: %w", i, err)
		}
	}

	return nil
}

func validateGenFields(portID, channelID string, sequence uint64) error {
	if err := host.PortIdentifierValidator(portID); err != nil {
		return err
	}
	if err := host.ChannelIdentifierValidator(channelID); err != nil {
		return err
	}
	if sequence == 0 {
		return errors.New("sequence cannot be 0")
	}
	return nil
}
