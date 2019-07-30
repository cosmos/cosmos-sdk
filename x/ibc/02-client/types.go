package client

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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

	// Validate() returns the updated consensus state
	// only if the header is a descendent of this consensus state.
	Validate(Header) (ConsensusState, error) // ValidityPredicate

	// Equivocation checks two headers' confliction.
	Equivocation(Header, Header) bool // EquivocationPredicate
}

/*
func Equal(client1, client2 ConsensusState) bool {
	return client1.Kind() == client2.Kind() &&
		client1.GetBase().Equal(client2.GetBase())
}
*/

// Header is the consensus state update information.
type Header interface {
	// Kind() is the kind of the consensus algorithm.
	Kind() Kind

	GetHeight() uint64
}

// XXX: Kind should be enum?

type Kind byte

const (
	Tendermint Kind = iota
)
