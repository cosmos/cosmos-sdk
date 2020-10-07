package v040

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Migrate accepts exported v0.36 x/gov genesis state and migrates it to
// v0.40 x/gov genesis state. The migration includes:
//
// - Re-encode in v0.40 GenesisState
func Migrate(govState v036gov.GenesisState) *v040gov.GenesisState {
	newLastValidatorPowers := make([]v040gov.LastValidatorPower, len(govState.LastValidatorPowers))
	for i, oldLastValidatorPower := range govState.LastValidatorPowers {
		newLastValidatorPowers[i] = v040gov.LastValidatorPower{
			Address: oldLastValidatorPower.Address.String(),
			Power:   oldLastValidatorPower.Power,
		}
	}

	newValidators := make([]v040gov.Validator, len(govState.Validators))
	for i, oldValidator := range govState.Validators {
		newValidators[i] = v040gov.Validator{
			OperatorAddress: oldValidator.OperatorAddress.String(),
			ConsensusPubkey: sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, oldValidator.ConsPubKey),
			Jailed:          oldValidator.Jailed,
			Status:          oldValidator.Status,
			Tokens:          oldValidator.Tokens,
			DelegatorShares: oldValidator.DelegatorShares,
			Description: v040gov.Description{
				Moniker:         oldValidator.Description.Moniker,
				Identity:        oldValidator.Description.Identity,
				Website:         oldValidator.Description.Website,
				SecurityContact: oldValidator.Description.SecurityContact,
				Details:         oldValidator.Description.Details,
			},
			UnbondingHeight: oldValidator.UnbondingHeight,
			UnbondingTime:   oldValidator.UnbondingCompletionTime,
			Commission: v040gov.Commission{
				CommissionRates: v040gov.CommissionRates{
					Rate:          oldValidator.Commission.Rate,
					MaxRate:       oldValidator.Commission.MaxRate,
					MaxChangeRate: oldValidator.Commission.MaxChangeRate,
				},
				UpdateTime: oldValidator.Commission.UpdateTime,
			},
			MinSelfDelegation: oldValidator.MinSelfDelegation,
		}
	}

	newDelegations := make([]v040gov.Delegation, len(govState.Delegations))
	for i, oldDelegation := range govState.Delegations {
		newDelegations[i] = v040gov.Delegation{
			DelegatorAddress: oldDelegation.DelegatorAddress.String(),
			ValidatorAddress: oldDelegation.ValidatorAddress.String(),
			Shares:           oldDelegation.Shares,
		}
	}

	newUnbondingDelegations := make([]v040gov.UnbondingDelegation, len(govState.UnbondingDelegations))
	for i, oldUnbondingDelegation := range govState.UnbondingDelegations {
		newEntries := make([]v040gov.UnbondingDelegationEntry, len(oldUnbondingDelegation.Entries))
		for j, oldEntry := range oldUnbondingDelegation.Entries {
			newEntries[j] = v040gov.UnbondingDelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				Balance:        oldEntry.Balance,
			}
		}

		newUnbondingDelegations[i] = v040gov.UnbondingDelegation{
			DelegatorAddress: oldUnbondingDelegation.DelegatorAddress.String(),
			ValidatorAddress: oldUnbondingDelegation.ValidatorAddress.String(),
			Entries:          newEntries,
		}
	}

	newRedelegations := make([]v040gov.Redelegation, len(govState.Redelegations))
	for i, oldRedelegation := range govState.Redelegations {
		newEntries := make([]v040gov.RedelegationEntry, len(oldRedelegation.Entries))
		for j, oldEntry := range oldRedelegation.Entries {
			newEntries[j] = v040gov.RedelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				SharesDst:      oldEntry.SharesDst,
			}
		}

		newRedelegations[i] = v040gov.Redelegation{
			DelegatorAddress:    oldRedelegation.DelegatorAddress.String(),
			ValidatorSrcAddress: oldRedelegation.ValidatorSrcAddress.String(),
			ValidatorDstAddress: oldRedelegation.ValidatorDstAddress.String(),
			Entries:             newEntries,
		}
	}

	return &v040gov.GenesisState{
		Params: v040gov.Params{
			UnbondingTime:     govState.Params.UnbondingTime,
			MaxValidators:     uint32(govState.Params.MaxValidators),
			MaxEntries:        uint32(govState.Params.MaxEntries),
			HistoricalEntries: uint32(govState.Params.HistoricalEntries),
			BondDenom:         govState.Params.BondDenom,
		},
		LastTotalPower:       govState.LastTotalPower,
		LastValidatorPowers:  newLastValidatorPowers,
		Validators:           newValidators,
		Delegations:          newDelegations,
		UnbondingDelegations: newUnbondingDelegations,
		Redelegations:        newRedelegations,
		Exported:             govState.Exported,
	}
}
