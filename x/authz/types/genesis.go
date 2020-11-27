package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// NewGenesisState creates new GenesisState object
func NewGenesisState(entries []MsgGrantAuthorization) *GenesisState {
	return &GenesisState{
		Authorization: entries,
	}
}

// ValidateGenesis check the given genesis state has no integrity issues
func ValidateGenesis(data GenesisState) error {
	return nil
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

var _ types.UnpackInterfacesMessage = GenesisState{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (data GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, authorization := range data.Authorization {
		err := authorization.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}
