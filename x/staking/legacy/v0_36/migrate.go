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
	params := Params{
		UnbondingTime: oldGenState.Params.UnbondingTime,
		MaxValidators: oldGenState.Params.MaxValidators,
		MaxEntries: oldGenState.Params.MaxEntries,
		BondDenom: oldGenState.Params.BondDenom,
	}

	valsPowers := make([]LastValidatorPower, len(oldGenState.LastValidatorPowers))
	for i, valPower := range oldGenState.LastValidatorPowers {
		valsPowers[i] = LastValidatorPower{
			Address: valPower.Address,
			Power: valPower.Power,
		}
	}

	vals := make(Validators, len(oldGenState.Validators))
	for i, val := range oldGenState.Validators {
		vals[i] = Validator{
			OperatorAddress: val.OperatorAddress,
			ConsPubKey: val.ConsPubKey,
			Jailed: val.Jailed,
			Status: val.Status,
			Tokens: val.Tokens,
			DelegatorShares: val.DelegatorShares,
			Description: Description{
				Moniker: val.Description.Moniker,
				Identity: val.Description.Identity,
				Website: val.Description.Website,
				Details: val.Description.Details,
			},
			UnbondingHeight: val.UnbondingHeight,
			UnbondingCompletionTime: val.UnbondingCompletionTime,
			Commission: Commission{
				CommissionRates: CommissionRates{
					Rate: val.Commission.CommissionRates.Rate,
					MaxRate: val.Commission.CommissionRates.MaxRate,
					MaxChangeRate: val.Commission.CommissionRates.MaxChangeRate,
				},
				UpdateTime: val.Commission.UpdateTime,
			},
			MinSelfDelegation: val.MinSelfDelegation,
		}
	}

	dels := make(Delegations, len(oldGenState.Delegations))
	for i, del := range oldGenState.Delegations {
		dels[i] = Delegation{
			DelegatorAddress: del.DelegatorAddress,
			ValidatorAddress: del.ValidatorAddress,
			Shares: del.Shares,
		}
	}

	ubds := make([]UnbondingDelegation, len(oldGenState.UnbondingDelegations))
	for i, ubd := range oldGenState.UnbondingDelegations {
		entries := make([]UnbondingDelegationEntry, len(ubd.Entries))
		for j, entry := range ubd.Entries {
			entries[j] = UnbondingDelegationEntry{
				CreationHeight: entry.CreationHeight,
				CompletionTime: entry.CompletionTime,
				InitialBalance: entry.InitialBalance,
				Balance: entry.Balance,
			}
		}

		ubds[i] = UnbondingDelegation{
			DelegatorAddress: ubd.DelegatorAddress,
			ValidatorAddress: ubd.ValidatorAddress,
			Entries: entries,
		}
	}

	reds := make([]Redelegation, len(oldGenState.Redelegations))
	for i, red := range oldGenState.Redelegations {

		entries := make([]RedelegationEntry, len(red.Entries))
		for j, entry := range red.Entries {
			entries[j] = RedelegationEntry{
				CreationHeight: entry.CreationHeight,
				CompletionTime: entry.CompletionTime,
				InitialBalance: entry.InitialBalance,
				SharesDst: entry.SharesDst,
			}
		}

		reds[i] = Redelegation{
			DelegatorAddress: red.DelegatorAddress,
			ValidatorSrcAddress: red.ValidatorSrcAddress,
			ValidatorDstAddress: red.ValidatorDstAddress,
			Entries: entries,
		}
	}

	return NewGenesisState(
		params,
		oldGenState.LastTotalPower,
		valsPowers,
		vals,
		dels,
		ubds,
		reds,
		true,
	)
}