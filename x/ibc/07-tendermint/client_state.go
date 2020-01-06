package tendermint

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

var _ clientexported.ClientState = ClientState{}

// ClientState from Tendermint tracks the current validator set, latest height,
// and a possible frozen height.
type ClientState struct {
	// Client ID
	ID string `json:"id" yaml:"id"`
	// Latests block height
	LastestHeight uint64 `json:"latest_height" yaml:"latest_height"`
	// Block height when the client was frozen due to a misbehaviour
	FrozenHeight uint64 `json:"frozen_height" yaml:"frozen_height"`
}

// NewClientState creates a new ClientState instance
func NewClientState(id string) ClientState {
	return ClientState{
		ID:            id,
		LastestHeight: 0,
		FrozenHeight:  0,
	}
}

// GetID returns the tendermint client state identifier.
func (cs ClientState) GetID() string {
	return cs.ID
}

// ClientType is tendermint.
func (cs ClientState) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// IsFrozen returns true if the frozen height has been set.
func (cs ClientState) IsFrozen() bool {
	return cs.FrozenHeight != 0
}

func (cs ClientState) VerifyClientConsensusState(
	height uint64, prefix commitment.PrefixI, proof commitment.ProofI,
	clientID string, consensusState clientexported.ConsensusState,
) error {

	return nil
}
