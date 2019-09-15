package v038

// DONTCOVER
// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_34"
)

const (
	ModuleName = "auth"
)

type GenesisState struct {
	Params   v034auth.Params          `json:"params" yaml:"params"`
	Accounts exported.GenesisAccounts `json:"accounts" yaml:"accounts"`
}

func NewGenesisState(params v034auth.Params, accounts exported.GenesisAccounts) GenesisState {
	return GenesisState{
		Params:   params,
		Accounts: accounts,
	}
}
