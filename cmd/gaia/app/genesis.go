package app

import (
	"encoding/json"
)

// XXX XXX beef dis up
// Genesis State of the blockchain
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for gaia.
func NewDefaultGenesisState() GenesisState {
	return BasicGaiaApp.DefaultGenesis()
}
