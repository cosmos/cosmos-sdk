package v038

import (
	"encoding/json"

	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_34"
)

// DONTCOVER

// nolint
const (
	ModuleName = "auth"
)

type GenesisState struct {
	Params   v034auth.Params `json:"params"`
	Accounts json.RawMessage `json:"accounts"`
}

func NewGenesisState(params v034auth.Params, accounts json.RawMessage) GenesisState {
	return GenesisState{
		Params:   params,
		Accounts: accounts,
	}
}
