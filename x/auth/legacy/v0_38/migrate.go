package v038

import (
	"encoding/json"

	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_36"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. This migration removes the CollectedFees coins from the old
// FeeCollectorKeeper.
func Migrate(oldGenState v036auth.GenesisState, accounts json.RawMessage) GenesisState {
	return NewGenesisState(oldGenState.Params, accounts)
}
