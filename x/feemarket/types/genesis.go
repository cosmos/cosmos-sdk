package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState returns a new genesis state for the module.
func NewGenesisState(
	params Params,
	state State,
) *GenesisState {
	return &GenesisState{
		Params: params,
		State:  state,
	}
}

// ValidateBasic performs basic validation of the genesis state data returning an
// error for any failed validation criteria.
func (gs *GenesisState) ValidateBasic() error {
	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}
	return gs.State.ValidateBasic()
}

// GetGenesisStateFromAppState returns x/feemarket GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var gs GenesisState
	cdc.MustUnmarshalJSON(appState[ModuleName], &gs)
	return gs
}
