package client

import (
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// TODO: types in this file should be (de/)serialized with proto in the future
// currently amino codec handles it

// ConsensusState is the state of the consensus process.
type ConsensusState interface {
	// Kind() is the kind of the consensus algorithm.
	Kind() Kind
	GetHeight() uint64

	// GetRoot() returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() commitment.Root

	// CheckValidityAndUpdateState() returns the updated consensus state
	// only if the header is a descendent of this consensus state.
	CheckValidityAndUpdateState(Header) (ConsensusState, error)

	// CheckMisbehaviourAndUpdateState() checks any misbehaviour evidence
	// depending on the state type.
	CheckMisbehaviourAndUpdateState(Misbehaviour) bool
}

type Misbehaviour interface {
	Kind() Kind
	// TODO: embed Evidence interface
	// evidence.Evidence
}

// Header is the consensus state update information.
type Header interface {
	// Kind() is the kind of the consensus algorithm.
	Kind() Kind

	GetHeight() uint64
}

type Kind byte

const (
	Tendermint Kind = iota
)
