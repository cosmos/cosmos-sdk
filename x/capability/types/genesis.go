package types

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// GenesisState represents the Capability module genesis state
type GenesisState struct {
	// capability global index
	Index uint64 `json:"index" yaml:"index"`

	// map from index to owners of the capability index
	// index key is string to allow amino marshalling
	Owners map[string]CapabilityOwners `json:"owners" yaml:"owners"`
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() GenesisState {
	return GenesisState{
		Index:  DefaultIndex,
		Owners: make(map[string]CapabilityOwners),
	}
}

// Validate validates the capability GenesiState. It returns an error if
// an owner contains a blank field.
func ValidateGenesis(data GenesisState) error {
	// NOTE: Index must be greater than 0
	if data.Index == 0 {
		return fmt.Errorf("capability index must be non-zero")
	}
	for i, co := range data.Owners {
		index, err := strconv.Atoi(i)
		if err != nil {
			return fmt.Errorf("owner string key %s must be a number, %v", i, err)
		}
		if len(co.Owners) == 0 {
			return fmt.Errorf("empty owners in genesis")
		}
		// All exported existing indices must be between [1, data.Index)
		if index <= 0 || uint64(index) >= data.Index {
			return fmt.Errorf("owners exist for index %d outside of valid range: %d-%d", index, 1, data.Index-1)
		}
		for _, owner := range co.Owners {
			if strings.TrimSpace(owner.Module) == "" {
				return fmt.Errorf("owner's module cannot be blank: %s", owner)
			}
			if strings.TrimSpace(owner.Name) == "" {
				return fmt.Errorf("owner's name cannot be blank: %s", owner)
			}
		}
	}

	return nil
}
