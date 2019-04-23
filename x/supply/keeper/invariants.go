package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterInvariants registers all supply invariants
func RegisterInvariants(ck CrisisKeeper, k Keeper, ak auth.AccountKeeper) {
	ck.RegisterRoute(ModuleName, "supply", SupplyInvariants(k, ak))
}

// AllInvariants runs all invariants of the staking module.
func AllInvariants(k Keeper, fck types.FeeCollectionKeeper,
	dk types.DistributionKeeper, ak auth.AccountKeeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		err := SupplyInvariants(k, ak)(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

// SupplyInvariants checks that the total supply reflects all held not-bonded tokens, bonded tokens, and unbonding delegations
// nolint: unparam
func SupplyInvariants(k Keeper, ak auth.AccountKeeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		supplier := k.GetSupplier(ctx)

		var circulatingAmount sdk.Coins
		var vestingAmount sdk.Coins
		var modulesAmount sdk.Coins

		ak.IterateAccounts(ctx, func(acc auth.Account) bool {
			vacc, isVestingAccount := acc.(auth.VestingAccount)
			if isVestingAccount && vacc.GetDelegatedVesting().IsAllPositive() && ctx.BlockHeader().Time.Unix() >= vacc.GetEndTime() {

				vestingAmount = vestingAmount.Add(vacc.GetOriginalVesting())
				circulatingAmount = circulatingAmount.Add(vacc.GetCoins())
			}

			macc, isModuleAccount := acc.(auth.ModuleAccount)
			if isModuleAccount {
				modulesAmount = modulesAmount.Add(macc.GetCoins())
			}

			return false
		})

		if !supplier.CirculatingSupply.IsEqual(circulatingAmount) {
			return fmt.Errorf("circulating supply invariance:\n"+
				"\tsupplier.CirculatingSupply: %v\n"+
				"\tsum of circulating tokens: %v", supplier.CirculatingSupply, circulatingAmount)
		}

		if !supplier.VestingSupply.IsEqual(vestingAmount) {
			return fmt.Errorf("vesting supply invariance:\n"+
				"\tsupplier.VestingSupply: %v\n"+
				"\tsum of vesting tokens: %v", supplier.VestingSupply, vestingAmount)
		}

		if !supplier.ModulesSupply.IsEqual(modulesAmount) {
			return fmt.Errorf("modules holdings supply invariance:\n"+
				"\tsupplier.ModulesSupply: %v\n"+
				"\tsum of modules accounts tokens: %v", supplier.ModulesSupply, modulesAmount)
		}

		return nil
	}
}

// // SupplyInvariants checks that the total supply reflects all held not-bonded tokens, bonded tokens, and unbonding delegations
// // nolint: unparam
// func SupplyInvariants(
// 	k Keeper, f FeeCollectionKeeper,
// 	d DistributionKeeper, am auth.AccountKeeper) sdk.Invariant {

// 	return func(ctx sdk.Context) error {
// 		pool := k.GetPool(ctx)

// 		loose := sdk.ZeroDec()
// 		bonded := sdk.ZeroDec()
// 		am.IterateAccounts(ctx, func(acc auth.Account) bool {
// 			loose = loose.Add(acc.GetCoins().AmountOf(k.BondDenom(ctx)).ToDec())
// 			return false
// 		})
// 		k.IterateUnbondingDelegations(ctx, func(_ int64, ubd types.UnbondingDelegation) bool {
// 			for _, entry := range ubd.Entries {
// 				loose = loose.Add(entry.Balance.ToDec())
// 			}
// 			return false
// 		})
// 		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
// 			switch validator.GetStatus() {
// 			case sdk.Bonded:
// 				bonded = bonded.Add(validator.GetBondedTokens().ToDec())
// 			case sdk.Unbonding, sdk.Unbonded:
// 				loose = loose.Add(validator.GetTokens().ToDec())
// 			}
// 			// add yet-to-be-withdrawn
// 			loose = loose.Add(d.GetValidatorOutstandingRewardsCoins(ctx, validator.GetOperator()).AmountOf(k.BondDenom(ctx)))
// 			return false
// 		})

// 		// add outstanding fees
// 		loose = loose.Add(f.GetCollectedFees(ctx).AmountOf(k.BondDenom(ctx)).ToDec())

// 		// add community pool
// 		loose = loose.Add(d.GetFeePoolCommunityCoins(ctx).AmountOf(k.BondDenom(ctx)))

// 		// Not-bonded tokens should equal coin supply plus unbonding delegations
// 		// plus tokens on unbonded validators
// 		if !pool.NotBondedTokens.ToDec().Equal(loose) {
// 			return fmt.Errorf("loose token invariance:\n"+
// 				"\tpool.NotBondedTokens: %v\n"+
// 				"\tsum of account tokens: %v", pool.NotBondedTokens, loose)
// 		}

// 		// Bonded tokens should equal sum of tokens with bonded validators
// 		if !pool.BondedTokens.ToDec().Equal(bonded) {
// 			return fmt.Errorf("bonded token invariance:\n"+
// 				"\tpool.BondedTokens: %v\n"+
// 				"\tsum of account tokens: %v", pool.BondedTokens, bonded)
// 		}

// 		return nil
// 	}
// }
