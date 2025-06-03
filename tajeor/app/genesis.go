package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// GenesisState represents the genesis state of the blockchain
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState returns the default genesis state
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return GenesisState{}
}

// ValidateGenesis validates the genesis state
func (gs GenesisState) ValidateGenesis() error {
	// Add validation logic here
	return nil
}
