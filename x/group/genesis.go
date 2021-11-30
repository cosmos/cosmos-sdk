package group

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

func (s GenesisState) Validate() error {
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (s GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, g := range s.GroupAccounts {
		err := g.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	for _, p := range s.Proposals {
		err := p.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}
