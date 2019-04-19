package app

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Genesis State of the blockchain
type GenesisState struct {
	Modules map[string]json.RawMessage `json:"modules"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(modules map[string]json.RawMessage) GenesisState {
	return GenesisState{
		Modules: modules,
	}
}

// NewDefaultGenesisState generates the default state for gaia.
func NewDefaultGenesisState() GenesisState {
	return NewGenesisState(nil, mbm.DefaultGenesis(), nil)
}

// TODO XXX ERADICATE

// initialize store from a genesis state
func (app *GaiaApp) initFromGenesisState(ctx sdk.Context, genesisState GenesisState) []abci.ValidatorUpdate {
	return app.mm.InitGenesis(ctx, genesisState.Modules)
}

// GaiaValidateGenesisState ensures that the genesis state obeys the expected invariants
func GaiaValidateGenesisState(genesisState GenesisState) error {
	return mbm.ValidateGenesis(genesisState.Modules)
}
