package app

import (
	"encoding/json"
)

// Genesis State of the blockchain
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for gaia.
func NewDefaultGenesisState() GenesisState {
	return mbm.DefaultGenesis()
}
