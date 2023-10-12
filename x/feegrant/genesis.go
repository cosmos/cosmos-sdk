package feegrant

import (
	"errors"

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
	for _, f := range data.Allowances {
		grant, err := f.GetGrant()
		if err != nil {
			return err
		}
		err = grant.ValidateBasic()
		if err != nil {
			return err
		}
	}

	if data.Params.EndblockPrunes == 0 || data.Params.ManualPrunes == 0 {
		return errors.New("invalid prune params")
	}

	return nil
}

// DefaultGenesisState returns default state for feegrant module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
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

func DefaultParams() Params {
	return Params{
		EndblockPrunes: 50,
		ManualPrunes:   5,
	}
}
