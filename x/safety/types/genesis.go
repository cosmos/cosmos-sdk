package types

import (
	"cosmossdk.io/core/address"
)

// NewGenesisState creates a new GenesisState.
func NewGenesisState(params Params, disabled []string) *GenesisState {
	return &GenesisState{Params: params, DisabledPatterns: disabled}
}

// DefaultGenesisState returns the default genesis state for the given
// authority: no operator, an empty scope, and an empty disabled list.
func DefaultGenesisState(authority string) *GenesisState {
	return NewGenesisState(DefaultParams(authority), nil)
}

// Validate checks the params and that every disabled pattern is well-formed
// and unique.
func (gs GenesisState) Validate(ac address.Codec) error {
	if err := gs.Params.Validate(ac); err != nil {
		return err
	}
	return validateUniquePatterns(gs.DisabledPatterns)
}
