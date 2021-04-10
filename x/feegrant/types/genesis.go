package types

import "github.com/cosmos/cosmos-sdk/codec/types"

var _ types.UnpackInterfacesMessage = GenesisState{}

// NewGenesisState creates new GenesisState object
func NewGenesisState(entries []FeeAllowanceGrant) *GenesisState {
	return &GenesisState{
		FeeAllowances: entries,
	}
}

// ValidateGenesis ensures all grants in the genesis state are valid
func ValidateGenesis(data GenesisState) error {
	for _, f := range data.FeeAllowances {
		grant, err := f.GetFeeGrant()
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
	for _, f := range data.FeeAllowances {
		err := f.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}

	return nil
}
