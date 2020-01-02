package exported

import (
	"encoding/json"
	"fmt"

	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	ClientType() ClientType // Consensus kind
	GetHeight() uint64

	// GetRoot returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() commitment.RootI

	// GetCommitter returns the committer that committed the consensus state
	GetCommitter() Committer

	// CheckValidityAndUpdateState returns the updated consensus state
	// only if the header is a descendent of this consensus state.
	CheckValidityAndUpdateState(Header) (ConsensusState, error)
}

// Misbehaviour defines a specific consensus kind and an evidence
type Misbehaviour interface {
	ClientType() ClientType
	GetEvidence() evidenceexported.Evidence
}

// Header is the consensus state update information
type Header interface {
	ClientType() ClientType
	GetCommitter() Committer
	GetHeight() uint64
}

// Committer defines the type that is responsible for
// updating the consensusState at a given height
//
// In Tendermint, this is the ValidatorSet at the given height
type Committer interface {
	ClientType() ClientType
	GetHeight() uint64
}

// ClientType defines the type of the consensus algorithm
type ClientType byte

// available client types
const (
	Tendermint ClientType = iota + 1 // 1
)

// string representation of the client types
const (
	ClientTypeTendermint string = "tendermint"
)

func (ct ClientType) String() string {
	switch ct {
	case Tendermint:
		return ClientTypeTendermint
	default:
		return ""
	}
}

// MarshalJSON marshal to JSON using string.
func (ct ClientType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}

// UnmarshalJSON decodes from JSON.
func (ct *ClientType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	clientType := ClientTypeFromString(s)
	if clientType == 0 {
		return fmt.Errorf("invalid client type '%s'", s)
	}

	*ct = clientType
	return nil
}

// ClientTypeFromString returns a byte that corresponds to the registered client
// type. It returns 0 if the type is not found/registered.
func ClientTypeFromString(clientType string) ClientType {
	switch clientType {
	case ClientTypeTendermint:
		return Tendermint

	default:
		return 0
	}
}
