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
		migrateValidators(oldGenState.Validators),
		oldGenState.Delegations,
		oldGenState.UnbondingDelegations,
		oldGenState.Redelegations,
		oldGenState.Exported,
	)
}

func migrateValidators(oldValidators v034staking.Validators) Validators {
	validators := make(Validators, len(oldValidators))

	for i, val := range oldValidators {
		validators[i] = Validator{
			OperatorAddress:         val.OperatorAddress,
			ConsPubKey:              val.ConsPubKey,
			Jailed:                  val.Jailed,
			Status:                  val.Status,
			Tokens:                  val.Tokens,
			DelegatorShares:         val.DelegatorShares,
			Description:             val.Description,
			UnbondingHeight:         val.UnbondingHeight,
			UnbondingCompletionTime: val.UnbondingCompletionTime,
			Commission: Commission{
				CommissionRates: CommissionRates{
					Rate:          val.Commission.Rate,
					MaxRate:       val.Commission.MaxRate,
					MaxChangeRate: val.Commission.MaxChangeRate,
				},
				UpdateTime: val.Commission.UpdateTime,
			},
			MinSelfDelegation: val.MinSelfDelegation,
		}
	}

	return validators
}
