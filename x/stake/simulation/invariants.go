package simulation

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

// AllInvariants runs all invariants of the stake module.
// Currently: total supply, positive power
func AllInvariants(ck bank.Keeper, k stake.Keeper,
	f auth.FeeCollectionKeeper, d distribution.Keeper,
	am auth.AccountKeeper) simulation.Invariant {

	return func(ctx sdk.Context) error {
		err := SupplyInvariants(ck, k, f, d, am)(ctx)
		if err != nil {
			return err
		}

		err = PositivePowerInvariant(k)(ctx)
		if err != nil {
			return err
		}

		err = PositiveDelegationInvariant(k)(ctx)
		if err != nil {
			return err
		}

		err = DelegatorSharesInvariant(k)(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

// SupplyInvariants checks that the total supply reflects all held loose tokens, bonded tokens, and unbonding delegations
// nolint: unparam
func SupplyInvariants(ck bank.Keeper, k stake.Keeper,
	f auth.FeeCollectionKeeper, d distribution.Keeper, am auth.AccountKeeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		pool := k.GetPool(ctx)

		loose := sdk.ZeroDec()
		bonded := sdk.ZeroDec()
		am.IterateAccounts(ctx, func(acc auth.Account) bool {
			loose = loose.Add(sdk.NewDecFromInt(acc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom)))
			return false
		})
		k.IterateUnbondingDelegations(ctx, func(_ int64, ubd stake.UnbondingDelegation) bool {
			loose = loose.Add(sdk.NewDecFromInt(ubd.Balance.Amount))
			return false
		})
		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetPower())
			case sdk.Unbonding:
				loose = loose.Add(validator.GetTokens())
			case sdk.Unbonded:
				loose = loose.Add(validator.GetTokens())
			}
			return false
		})

		feePool := d.GetFeePool(ctx)

		// add outstanding fees
		loose = loose.Add(sdk.NewDecFromInt(f.GetCollectedFees(ctx).AmountOf(stakeTypes.DefaultBondDenom)))

		// add community pool
		loose = loose.Add(feePool.CommunityPool.AmountOf(stakeTypes.DefaultBondDenom))

		// add validator distribution pool
		loose = loose.Add(feePool.ValPool.AmountOf(stakeTypes.DefaultBondDenom))

		// add validator distribution commission and yet-to-be-withdrawn-by-delegators
		d.IterateValidatorDistInfos(ctx,
			func(_ int64, distInfo distribution.ValidatorDistInfo) (stop bool) {
				loose = loose.Add(distInfo.DelPool.AmountOf(stakeTypes.DefaultBondDenom))
				loose = loose.Add(distInfo.ValCommission.AmountOf(stakeTypes.DefaultBondDenom))
				return false
			},
		)

		// Loose tokens should equal coin supply plus unbonding delegations
		// plus tokens on unbonded validators
		if !pool.LooseTokens.Equal(loose) {
			return fmt.Errorf("loose token invariance:\n\tpool.LooseTokens: %v"+
				"\n\tsum of account tokens: %v", pool.LooseTokens, loose)
		}

		// Bonded tokens should equal sum of tokens with bonded validators
		if !pool.BondedTokens.Equal(bonded) {
			return fmt.Errorf("bonded token invariance:\n\tpool.BondedTokens: %v"+
				"\n\tsum of account tokens: %v", pool.BondedTokens, bonded)
		}

		return nil
	}
}

// PositivePowerInvariant checks that all stored validators have > 0 power.
func PositivePowerInvariant(k stake.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		iterator := k.ValidatorsPowerStoreIterator(ctx)
		pool := k.GetPool(ctx)

		for ; iterator.Valid(); iterator.Next() {
			validator, found := k.GetValidator(ctx, iterator.Value())
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %X\n", iterator.Value()))
			}

			powerKey := keeper.GetValidatorsByPowerIndexKey(validator, pool)

			if !bytes.Equal(iterator.Key(), powerKey) {
				return fmt.Errorf("power store invariance:\n\tvalidator.Power: %v"+
					"\n\tkey should be: %v\n\tkey in store: %v", validator.GetPower(), powerKey, iterator.Key())
			}
		}
		iterator.Close()
		return nil
	}
}

// PositiveDelegationInvariant checks that all stored delegations have > 0 shares.
func PositiveDelegationInvariant(k stake.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		delegations := k.GetAllDelegations(ctx)
		for _, delegation := range delegations {
			if delegation.Shares.IsNegative() {
				return fmt.Errorf("delegation with negative shares: %+v", delegation)
			}
			if delegation.Shares.IsZero() {
				return fmt.Errorf("delegation with zero shares: %+v", delegation)
			}
		}

		return nil
	}
}

// DelegatorSharesInvariant checks whether all the delegator shares which persist
// in the delegator object add up to the correct total delegator shares
// amount stored in each validator
func DelegatorSharesInvariant(k stake.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		validators := k.GetAllValidators(ctx)
		for _, validator := range validators {

			valTotalDelShares := validator.GetDelegatorShares()

			totalDelShares := sdk.ZeroDec()
			delegations := k.GetValidatorDelegations(ctx, validator.GetOperator())
			for _, delegation := range delegations {
				totalDelShares = totalDelShares.Add(delegation.Shares)
			}

			if !valTotalDelShares.Equal(totalDelShares) {
				return fmt.Errorf("broken delegator shares invariance:\n"+
					"\tvalidator.DelegatorShares: %v\n"+
					"\tsum of Delegator.Shares: %v", valTotalDelShares, totalDelShares)
			}
		}
		return nil
	}
}
