package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
)

// NewGenesisState creates new GenesisState object
func NewGenesisState(entries []GrantAuthorization) *GenesisState {
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg GrantAuthorization) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var authorization exported.Authorization
	return unpacker.UnpackAny(msg.Authorization, &authorization)
}
