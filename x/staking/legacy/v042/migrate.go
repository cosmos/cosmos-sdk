package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v040"
	v042staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Migrate accepts exported v0.40 x/staking genesis state and migrates it to
// v0.42 x/staking genesis state. The migration includes:
//
// - Adding power reduction on-chain param
func Migrate(stakingState v040staking.GenesisState) *v042staking.GenesisState {
	newLastValidatorPowers := make([]v042staking.LastValidatorPower, len(stakingState.LastValidatorPowers))
	for i, oldLastValidatorPower := range stakingState.LastValidatorPowers {
		newLastValidatorPowers[i] = v042staking.LastValidatorPower{
			Address: oldLastValidatorPower.Address,
			Power:   oldLastValidatorPower.Power,
		}
	}

	newValidators := make([]v042staking.Validator, len(stakingState.Validators))
	for i, oldValidator := range stakingState.Validators {
		newValidators[i] = v042staking.Validator{
			OperatorAddress: oldValidator.OperatorAddress,
			ConsensusPubkey: oldValidator.ConsensusPubkey,
			Jailed:          oldValidator.Jailed,
			Status:          v042staking.BondStatus(oldValidator.Status),
			Tokens:          oldValidator.Tokens,
			DelegatorShares: oldValidator.DelegatorShares,
			Description: v042staking.Description{
				Moniker:         oldValidator.Description.Moniker,
				Identity:        oldValidator.Description.Identity,
				Website:         oldValidator.Description.Website,
				SecurityContact: oldValidator.Description.SecurityContact,
				Details:         oldValidator.Description.Details,
			},
			UnbondingHeight: oldValidator.UnbondingHeight,
			UnbondingTime:   oldValidator.UnbondingTime,
			Commission: v042staking.Commission{
				CommissionRates: v042staking.CommissionRates{
					Rate:          oldValidator.Commission.Rate,
					MaxRate:       oldValidator.Commission.MaxRate,
					MaxChangeRate: oldValidator.Commission.MaxChangeRate,
				},
				UpdateTime: oldValidator.Commission.UpdateTime,
			},
			MinSelfDelegation: oldValidator.MinSelfDelegation,
		}
	}

	newDelegations := make([]v042staking.Delegation, len(stakingState.Delegations))
	for i, oldDelegation := range stakingState.Delegations {
		newDelegations[i] = v042staking.Delegation{
			DelegatorAddress: oldDelegation.DelegatorAddress,
			ValidatorAddress: oldDelegation.ValidatorAddress,
			Shares:           oldDelegation.Shares,
		}
	}

	newUnbondingDelegations := make([]v042staking.UnbondingDelegation, len(stakingState.UnbondingDelegations))
	for i, oldUnbondingDelegation := range stakingState.UnbondingDelegations {
		newEntries := make([]v042staking.UnbondingDelegationEntry, len(oldUnbondingDelegation.Entries))
		for j, oldEntry := range oldUnbondingDelegation.Entries {
			newEntries[j] = v042staking.UnbondingDelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				Balance:        oldEntry.Balance,
			}
		}

		newUnbondingDelegations[i] = v042staking.UnbondingDelegation{
			DelegatorAddress: oldUnbondingDelegation.DelegatorAddress,
			ValidatorAddress: oldUnbondingDelegation.ValidatorAddress,
			Entries:          newEntries,
		}
	}

	newRedelegations := make([]v042staking.Redelegation, len(stakingState.Redelegations))
	for i, oldRedelegation := range stakingState.Redelegations {
		newEntries := make([]v042staking.RedelegationEntry, len(oldRedelegation.Entries))
		for j, oldEntry := range oldRedelegation.Entries {
			newEntries[j] = v042staking.RedelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				SharesDst:      oldEntry.SharesDst,
			}
		}

		newRedelegations[i] = v042staking.Redelegation{
			DelegatorAddress:    oldRedelegation.DelegatorAddress,
			ValidatorSrcAddress: oldRedelegation.ValidatorSrcAddress,
			ValidatorDstAddress: oldRedelegation.ValidatorDstAddress,
			Entries:             newEntries,
		}
	}

	return &v042staking.GenesisState{
		Params: v042staking.Params{
			UnbondingTime:     stakingState.Params.UnbondingTime,
			MaxValidators:     stakingState.Params.MaxValidators,
			MaxEntries:        stakingState.Params.MaxEntries,
			HistoricalEntries: stakingState.Params.HistoricalEntries,
			BondDenom:         stakingState.Params.BondDenom,
			PowerReduction:    sdk.DefaultPowerReduction,
		},
		LastTotalPower:       stakingState.LastTotalPower,
		LastValidatorPowers:  newLastValidatorPowers,
		Validators:           newValidators,
		Delegations:          newDelegations,
		UnbondingDelegations: newUnbondingDelegations,
		Redelegations:        newRedelegations,
		Exported:             stakingState.Exported,
	}
}
