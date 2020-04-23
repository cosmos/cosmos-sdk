package types

import (
	"errors"
	"fmt"
	"strings"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// GenesisState represents the Capability module genesis state
type GenesisState struct {
	// capability global index
	Index uint64 `json:"index" yaml:"index"`

	Owners []Owner `json:"owners" yaml:"owners"`
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() GenesisState {
	return GenesisState{
		Index:  DefaultIndex,
		Owners: []Owner{},
	}
}

// Validate validates the capability GenesiState. It returns an error if
// an owner contains a blank field.
func (gs GenesisState) Validate() error {
	if gs.Index == 0 {
		return errors.New("global index cannot be 0")
	}

	for _, owner := range gs.Owners {
		if strings.TrimSpace(owner.Module) == "" {
			return fmt.Errorf("owner's module cannot be blank: %s", owner)
		}
		if strings.TrimSpace(owner.Name) == "" {
			return fmt.Errorf("owner's name cannot be blank: %s", owner)
		}
	}

	return nil
}
