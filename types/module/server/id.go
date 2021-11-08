package server

import sdk "github.com/cosmos/cosmos-sdk/types"

type ModuleID struct {
	ModuleName string
	Path       []byte
}

func (m ModuleID) Address() sdk.AccAddress {
	return AddressHash(m.ModuleName, m.Path)
}
