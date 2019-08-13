// DONTCOVER
// nolint
package v0_37

import (
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_36"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. All entries are identical except for validator slashing events
// which now include the period.
func Migrate(oldGenState v036staking.GenesisState) GenesisState {
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

func migrateValidators(oldValidators v036staking.Validators) Validators {
	validators := make(Validators, len(oldValidators))

	for i, val := range oldValidators {
		validators[i] = Validator{
			OperatorAddress: val.OperatorAddress,
			ConsPubKey:      val.ConsPubKey,
			Jailed:          val.Jailed,
			Status:          val.Status,
			Tokens:          val.Tokens,
			DelegatorShares: val.DelegatorShares,
			Description: NewDescription(
				val.Description.Moniker,
				val.Description.Identity,
				val.Description.Website,
				"", // security contact field
				val.Description.Details,
			),
			UnbondingHeight:         val.UnbondingHeight,
			UnbondingCompletionTime: val.UnbondingCompletionTime,
			Commission:              val.Commission,
			MinSelfDelegation:       val.MinSelfDelegation,
		}
	}

	return validators
}
