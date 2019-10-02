package exported

import (
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Blockchain is consensus algorithm which generates valid Headers. It generates
// a unique list of headers starting from a genesis ConsensusState with arbitrary messages.
// This interface is implemented as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#blockchain.
type Blockchain interface {
	Genesis() ConsensusState // Consensus state defined in the genesis
	Consensus() Header       // Header generating funciton
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	Kind() Kind // Consensus kind
	GetHeight() uint64

	// GetRoot returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() ics23.Root

	// CheckValidityAndUpdateState returns the updated consensus state
	// only if the header is a descendent of this consensus state.
	CheckValidityAndUpdateState(Header) (ConsensusState, error)

	// CheckMisbehaviourAndUpdateState checks any misbehaviour evidence
	// depending on the state type.
	CheckMisbehaviourAndUpdateState(Misbehaviour) bool
}

// Misbehaviour defines the evidence
type Misbehaviour interface {
	Kind() Kind
	// TODO: embed Evidence interface
	// evidence.Evidence
}

// Header is the consensus state update information
type Header interface {
	Kind() Kind
	GetHeight() uint64
	ValidateBasic(chainID string) error // NOTE: added for msg validation
}

// Kind defines the type of the consensus algorithm
type Kind byte

// Registered consensus types
const (
	Tendermint Kind = iota
)
