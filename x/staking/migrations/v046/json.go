package v046

import "github.com/cosmos/cosmos-sdk/x/staking/types"

// MigrateJSON accepts exported v0.43 x/stakinng genesis state and migrates it to
// v0.46 x/staking genesis state. The migration includes:
//
// - Add MinCommissionRate param.
func MigrateJSON(oldState *types.GenesisState) (*types.GenesisState, error) {
	return &types.GenesisState{
		Params: types.Params{
			UnbondingTime:     oldState.Params.UnbondingTime,
			MaxValidators:     oldState.Params.MaxValidators,
			MaxEntries:        oldState.Params.MaxEntries,
			HistoricalEntries: oldState.Params.HistoricalEntries,
			BondDenom:         oldState.Params.BondDenom,
			MinCommissionRate: types.DefaultMinCommissionRate,
		},
		LastTotalPower:       oldState.LastTotalPower,
		LastValidatorPowers:  oldState.LastValidatorPowers,
		Validators:           oldState.Validators,
		Delegations:          oldState.Delegations,
		UnbondingDelegations: oldState.UnbondingDelegations,
		Redelegations:        oldState.Redelegations,
		Exported:             oldState.Exported,
	}, nil
}
