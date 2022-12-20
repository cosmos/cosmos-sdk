package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// // NewGenesisState - Create a new genesis state
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// // DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	// return NewGenesisState(DefaultParams())
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// GetGenesisStateFromAppState returns x/auth GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

func ValidateGenesis(data GenesisState) error {
	if err := data.Params.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
