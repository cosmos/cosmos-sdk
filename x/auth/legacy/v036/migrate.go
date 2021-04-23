// DONTCOVER
package v036

import (
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. This migration removes the CollectedFees coins from the old
// FeeCollectorKeeper.
func Migrate(oldGenState v034auth.GenesisState) GenesisState {
	return NewGenesisState(oldGenState.Params)
}
