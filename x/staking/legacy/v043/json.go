package v043

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v040"
	v043staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateJSON accepts exported v0.40 x/gov genesis state and migrates it to
// v0.43 x/gov genesis state. The migration includes:
//
// - Gov weighted votes.
func MigrateJSON(oldState *v040staking.GenesisState) *v043staking.GenesisState {
	return &v043staking.GenesisState{
		Params:               migrateParams(oldState.Params),
		LastTotalPower:       oldState.LastTotalPower,
		LastValidatorPowers:  migrateLastValidatorPowers(oldState.LastValidatorPowers),
		Validators:           migrateValidators(oldState.Validators),
		Delegations:          migrateDelegations(oldState.Delegations),
		UnbondingDelegations: migrateUnbondingDelegations(oldState.UnbondingDelegations),
		Redelegations:        migrateRedelegations(oldState.Redelegations),
		Exported:             oldState.Exported,
	}
}

func migrateParams(oldParams v040staking.Params) v043staking.Params {
	return v043staking.NewParams(
		oldParams.UnbondingTime,
		oldParams.MaxValidators,
		oldParams.MaxEntries,
		oldParams.HistoricalEntries,
		oldParams.BondDenom,
		sdk.DefaultPowerReduction,
	)
}

func migrateLastValidatorPowers(oldLastValidatorPowers []v040staking.LastValidatorPower) []v043staking.LastValidatorPower {
	newLastValidatorPowers := make([]v043staking.LastValidatorPower, len(oldLastValidatorPowers))
	for i, oldLastValidatorPower := range oldLastValidatorPowers {
		newLastValidatorPowers[i] = v043staking.LastValidatorPower{
			Address: oldLastValidatorPower.Address,
			Power:   oldLastValidatorPower.Power,
		}
	}
	return newLastValidatorPowers
}

func migrateValidators(oldValidators []v040staking.Validator) []v043staking.Validator {
	newValidators := make([]v043staking.Validator, len(oldValidators))

	for i, oldValidator := range oldValidators {
		newValidators[i] = v043staking.Validator{
			OperatorAddress: oldValidator.OperatorAddress,
			ConsensusPubkey: oldValidator.ConsensusPubkey,
			Jailed:          oldValidator.Jailed,
			Status:          v043staking.BondStatus(oldValidator.Status),
			Tokens:          oldValidator.Tokens,
			DelegatorShares: oldValidator.DelegatorShares,
			Description: v043staking.Description{
				Moniker:         oldValidator.Description.Moniker,
				Identity:        oldValidator.Description.Identity,
				Website:         oldValidator.Description.Website,
				SecurityContact: oldValidator.Description.SecurityContact,
				Details:         oldValidator.Description.Details,
			},
			UnbondingHeight: oldValidator.UnbondingHeight,
			UnbondingTime:   oldValidator.UnbondingTime,
			Commission: v043staking.Commission{
				CommissionRates: v043staking.CommissionRates{
					Rate:          oldValidator.Commission.CommissionRates.Rate,
					MaxRate:       oldValidator.Commission.CommissionRates.MaxRate,
					MaxChangeRate: oldValidator.Commission.CommissionRates.MaxChangeRate,
				},
				UpdateTime: oldValidator.Commission.UpdateTime,
			},
			MinSelfDelegation: oldValidator.MinSelfDelegation,
		}
	}

	return newValidators
}

func migrateDelegations(oldDelegations []v040staking.Delegation) []v043staking.Delegation {
	newDelegations := make([]v043staking.Delegation, len(oldDelegations))
	for i, oldDelegation := range oldDelegations {
		newDelegations[i] = v043staking.Delegation{
			DelegatorAddress: oldDelegation.DelegatorAddress,
			ValidatorAddress: oldDelegation.ValidatorAddress,
			Shares:           oldDelegation.Shares,
		}
	}
	return newDelegations
}

func migrateUnbondingDelegations(oldUnbondingDelegations []v040staking.UnbondingDelegation) []v043staking.UnbondingDelegation {
	newUnbondingDelegations := make([]v043staking.UnbondingDelegation, len(oldUnbondingDelegations))
	for i, oldUnbondingDelegation := range oldUnbondingDelegations {
		newEntries := make([]v043staking.UnbondingDelegationEntry, len(oldUnbondingDelegation.Entries))
		for j, oldEntry := range oldUnbondingDelegation.Entries {
			newEntries[j] = v043staking.UnbondingDelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				Balance:        oldEntry.Balance,
			}
		}

		newUnbondingDelegations[i] = v043staking.UnbondingDelegation{
			DelegatorAddress: oldUnbondingDelegation.DelegatorAddress,
			ValidatorAddress: oldUnbondingDelegation.ValidatorAddress,
			Entries:          newEntries,
		}
	}
	return newUnbondingDelegations
}

func migrateRedelegations(oldRedelegations []v040staking.Redelegation) []v043staking.Redelegation {
	newRedelegations := make([]v043staking.Redelegation, len(oldRedelegations))
	for i, oldRedelegation := range oldRedelegations {
		newEntries := make([]v043staking.RedelegationEntry, len(oldRedelegation.Entries))
		for j, oldEntry := range oldRedelegation.Entries {
			newEntries[j] = v043staking.RedelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				SharesDst:      oldEntry.SharesDst,
			}
		}

		newRedelegations[i] = v043staking.Redelegation{
			DelegatorAddress:    oldRedelegation.DelegatorAddress,
			ValidatorSrcAddress: oldRedelegation.ValidatorSrcAddress,
			ValidatorDstAddress: oldRedelegation.ValidatorDstAddress,
			Entries:             newEntries,
		}
	}
	return newRedelegations
}
