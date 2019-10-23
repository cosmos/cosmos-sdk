package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
	GetRoot() commitment.RootI

	// CheckValidityAndUpdateState returns the updated consensus state
	// only if the header is a descendent of this consensus state.
	CheckValidityAndUpdateState(Header) (ConsensusState, error)
}

// Evidence from ADR 009: Evidence Module
// TODO: use evidence module interface once merged
type Evidence interface {
	Route() string
	Type() string
	String() string
	ValidateBasic() sdk.Error

	// The consensus address of the malicious validator at time of infraction
	GetConsensusAddress() sdk.ConsAddress

	// Height at which the infraction occurred
	GetHeight() int64

	// The total power of the malicious validator at time of infraction
	GetValidatorPower() int64

	// The total validator set power at time of infraction
	GetTotalPower() int64
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

// Client types
const (
	ClientTypeTendermint string = "tendermint"
)

// ClientType defines the type of the consensus algorithm
type ClientType byte

// Registered consensus types
const (
	Tendermint ClientType = iota + 1 // 1
)

// ValidClientTypes returns the registerd client types for this chain
var ValidClientTypes = map[string]bool{
	ClientTypeTendermint: true,
}

// ClientTypeFromStr returns a byte that corresponds to the registered client
// type. It returns 0 if the type is not found/registered.
func ClientTypeFromStr(clientType string) ClientType {
	switch clientType {
	case ClientTypeTendermint:
		return Tendermint

	default:
		return 0
	}
}

// ClientTypeToString returns the string representation of a client type
func ClientTypeToString(clientType ClientType) string {
	switch clientType {
	case Tendermint:
		return ClientTypeTendermint

	default:
		return ""
	}
}
