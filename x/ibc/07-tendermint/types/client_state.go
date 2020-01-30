package types

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ clientexported.ClientState = ClientState{}

// ClientState from Tendermint tracks the current validator set, latest height,
// and a possible frozen height.
type ClientState struct {
	// Client ID
	ID string `json:"id" yaml:"id"`
	// Latest block height
	LatestHeight uint64 `json:"latest_height" yaml:"latest_height"`
	// Block height when the client was frozen due to a misbehaviour
	FrozenHeight uint64 `json:"frozen_height" yaml:"frozen_height"`
}

// NewClientState creates a new ClientState instance
func NewClientState(id string, latestHeight uint64) ClientState {
	return ClientState{
		ID:           id,
		LatestHeight: latestHeight,
		FrozenHeight: 0,
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

// GetLatestHeight returns latest block height.
func (cs ClientState) GetLatestHeight() uint64 {
	return cs.LatestHeight
}

// IsFrozen returns true if the frozen height has been set.
func (cs ClientState) IsFrozen() bool {
	return cs.FrozenHeight != 0
}
