package feegrant

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

var _ types.UnpackInterfacesMessage = GenesisState{}

// NewGenesisState creates new GenesisState object
func NewGenesisState(entries []Grant) *GenesisState {
	return &GenesisState{
		Allowances: entries,
	}
}

// ValidateGenesis ensures all grants in the genesis state are valid
func ValidateGenesis(data GenesisState) error {
	// Check for duplicate grants by (granter, grantee) pair
	seen := make(map[string]struct{})

	for _, f := range data.Allowances {
		// Create a unique key for (granter, grantee) pair using a delimiter
		key := fmt.Sprintf("%s|%s", f.Granter, f.Grantee)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate feegrant found from granter %q to grantee %q", f.Granter, f.Grantee)
		}
		seen[key] = struct{}{}

		grant, err := f.GetGrant()
		if err != nil {
			return err
		}
		err = grant.ValidateBasic()
		if err != nil {
			return err
		}
	}
	return nil
}

// DefaultGenesisState returns default state for feegrant module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (data GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, f := range data.Allowances {
		err := f.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}

	return nil
}
