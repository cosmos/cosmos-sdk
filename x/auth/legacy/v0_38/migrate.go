package v038

import (
	"encoding/json"

	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_36"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.38
// genesis state.
func Migrate(oldGenState v036auth.GenesisState, accounts json.RawMessage) GenesisState {
	return NewGenesisState(oldGenState.Params, accounts)
}
