package v040

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v034"
	v038staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v038"
)

func migrateBondStatus(oldStatus v034staking.BondStatus) BondStatus {
	switch oldStatus {
	case v034staking.Unbonded:
		return Unbonded

	case v034staking.Unbonding:
		return Unbonding

	case v034staking.Bonded:
		return Bonded

	default:
		panic(fmt.Errorf("invalid bond status %d", oldStatus))
	}
}

// Migrate accepts exported v0.38 x/staking genesis state and migrates it to
// v0.40 x/staking genesis state. The migration includes:
//
// - Convert addresses from bytes to bech32 strings.
// - Update BondStatus staking constants.
// - Re-encode in v0.40 GenesisState.
func Migrate(stakingState v038staking.GenesisState) *GenesisState {
	newLastValidatorPowers := make([]LastValidatorPower, len(stakingState.LastValidatorPowers))
	for i, oldLastValidatorPower := range stakingState.LastValidatorPowers {
		newLastValidatorPowers[i] = LastValidatorPower{
			Address: oldLastValidatorPower.Address.String(),
			Power:   oldLastValidatorPower.Power,
		}
	}

	newValidators := make([]Validator, len(stakingState.Validators))
	for i, oldValidator := range stakingState.Validators {
		pkAny, err := codectypes.NewAnyWithValue(oldValidator.ConsPubKey)
		if err != nil {
			panic(fmt.Sprintf("Can't pack validator consensus PK as Any: %s", err))
		}
		newValidators[i] = Validator{
			OperatorAddress: oldValidator.OperatorAddress.String(),
			ConsensusPubkey: pkAny,
			Jailed:          oldValidator.Jailed,
			Status:          migrateBondStatus(oldValidator.Status),
			Tokens:          oldValidator.Tokens,
			DelegatorShares: oldValidator.DelegatorShares,
			Description: Description{
				Moniker:         oldValidator.Description.Moniker,
				Identity:        oldValidator.Description.Identity,
				Website:         oldValidator.Description.Website,
				SecurityContact: oldValidator.Description.SecurityContact,
				Details:         oldValidator.Description.Details,
			},
			UnbondingHeight: oldValidator.UnbondingHeight,
			UnbondingTime:   oldValidator.UnbondingCompletionTime,
			Commission: Commission{
				CommissionRates: CommissionRates{
					Rate:          oldValidator.Commission.Rate,
					MaxRate:       oldValidator.Commission.MaxRate,
					MaxChangeRate: oldValidator.Commission.MaxChangeRate,
				},
				UpdateTime: oldValidator.Commission.UpdateTime,
			},
			MinSelfDelegation: oldValidator.MinSelfDelegation,
		}
	}

	newDelegations := make([]Delegation, len(stakingState.Delegations))
	for i, oldDelegation := range stakingState.Delegations {
		newDelegations[i] = Delegation{
			DelegatorAddress: oldDelegation.DelegatorAddress.String(),
			ValidatorAddress: oldDelegation.ValidatorAddress.String(),
			Shares:           oldDelegation.Shares,
		}
	}

	newUnbondingDelegations := make([]UnbondingDelegation, len(stakingState.UnbondingDelegations))
	for i, oldUnbondingDelegation := range stakingState.UnbondingDelegations {
		newEntries := make([]UnbondingDelegationEntry, len(oldUnbondingDelegation.Entries))
		for j, oldEntry := range oldUnbondingDelegation.Entries {
			newEntries[j] = UnbondingDelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				Balance:        oldEntry.Balance,
			}
		}

		newUnbondingDelegations[i] = UnbondingDelegation{
			DelegatorAddress: oldUnbondingDelegation.DelegatorAddress.String(),
			ValidatorAddress: oldUnbondingDelegation.ValidatorAddress.String(),
			Entries:          newEntries,
		}
	}

	newRedelegations := make([]Redelegation, len(stakingState.Redelegations))
	for i, oldRedelegation := range stakingState.Redelegations {
		newEntries := make([]RedelegationEntry, len(oldRedelegation.Entries))
		for j, oldEntry := range oldRedelegation.Entries {
			newEntries[j] = RedelegationEntry{
				CreationHeight: oldEntry.CreationHeight,
				CompletionTime: oldEntry.CompletionTime,
				InitialBalance: oldEntry.InitialBalance,
				SharesDst:      oldEntry.SharesDst,
			}
		}

		newRedelegations[i] = Redelegation{
			DelegatorAddress:    oldRedelegation.DelegatorAddress.String(),
			ValidatorSrcAddress: oldRedelegation.ValidatorSrcAddress.String(),
			ValidatorDstAddress: oldRedelegation.ValidatorDstAddress.String(),
			Entries:             newEntries,
		}
	}

	return &GenesisState{
		Params: Params{
			UnbondingTime:     stakingState.Params.UnbondingTime,
			MaxValidators:     uint32(stakingState.Params.MaxValidators),
			MaxEntries:        uint32(stakingState.Params.MaxEntries),
			HistoricalEntries: uint32(stakingState.Params.HistoricalEntries),
			BondDenom:         stakingState.Params.BondDenom,
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
