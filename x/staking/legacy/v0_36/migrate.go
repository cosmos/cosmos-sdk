// DONTCOVER
// nolint
package v0_36

import (
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_34"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. All entries are identical except for validator slashing events
// which now include the period.
func Migrate(oldGenState v034staking.GenesisState) GenesisState {
	return NewGenesisState(
		oldGenState.Params,
		oldGenState.LastTotalPower,
		oldGenState.LastValidatorPowers,
		oldGenState.Validators,
		oldGenState.Delegations,
		oldGenState.UnbondingDelegations,
		oldGenState.Redelegations,
		oldGenState.Exported,
	)
}
