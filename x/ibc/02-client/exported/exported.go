package exported

import (
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Blockchain is consensus algorithm which generates valid Headers. It generates
// a unique list of headers starting from a genesis ConsensusState with arbitrary messages.
// This interface is implemented as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#blockchain.
type Blockchain interface {
	Genesis() ConsensusState // Consensus state defined in the genesis
	Consensus() Header       // Header generating function
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	ClientType() ClientType // Consensus kind
	GetHeight() uint64

	// GetRoot returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() ics23.Root

	// CheckValidityAndUpdateState returns the updated consensus state
	// only if the header is a descendent of this consensus state.
	CheckValidityAndUpdateState(Header) (ConsensusState, error)
}

// Evidence contains two disctict headers used to submit client equivocation
// TODO: use evidence module type
type Evidence interface {
	H1() Header
	H2() Header
}

// Misbehaviour defines a specific consensus kind and an evidence
type Misbehaviour interface {
	ClientType() ClientType
	Evidence() Evidence
}

// Header is the consensus state update information
type Header interface {
	ClientType() ClientType
	GetHeight() uint64
}

// ClientType defines the type of the consensus algorithm
type ClientType byte

// Registered consensus types
const (
	Tendermint ClientType = iota + 1 // 1
)
