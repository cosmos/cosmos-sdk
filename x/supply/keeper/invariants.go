package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// RegisterInvariants registers all supply invariants
func RegisterInvariants(ck CrisisKeeper, k Keeper, accountKeeper AccountKeeper,
	distributionKeeper DistributionKeeper, feeCollectionKeeper FeeCollectionKeeper,
	stakingKeeper StakingKeeper) {
	ck.RegisterRoute(
		ModuleName, "supply",
		SupplyInvariants(k, accountKeeper, distributionKeeper, feeCollectionKeeper, stakingKeeper))
}

// AllInvariants runs all invariants of the staking module.
func AllInvariants(k Keeper, accountKeeper AccountKeeper, distributionKeeper DistributionKeeper,
	feeCollectionKeeper FeeCollectionKeeper, stakingKeeper StakingKeeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		err := SupplyInvariants(k, accountKeeper, distributionKeeper, feeCollectionKeeper, stakingKeeper)(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

// SupplyInvariants checks that the total supply reflects all held not-bonded tokens, bonded tokens, and unbonding delegations
func SupplyInvariants(k Keeper,
	accountKeeper AccountKeeper, distributionKeeper DistributionKeeper,
	feeCollectionKeeper FeeCollectionKeeper, stakingKeeper StakingKeeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		supplier := k.GetSupplier(ctx)

		var circulatingAmount sdk.Coins
		var modulesAmount sdk.Coins
		var initialVestingAmount sdk.Coins

		accountKeeper.IterateAccounts(ctx, func(acc auth.Account) bool {

			vacc, isVestingAccount := acc.(auth.VestingAccount)
			if isVestingAccount && ctx.BlockHeader().Time.Unix() < vacc.GetEndTime() {
				initialVestingAmount = initialVestingAmount.Add(vacc.GetOriginalVesting())
			}

			macc, isModuleAccount := acc.(auth.ModuleAccount)
			if isModuleAccount {
				modulesAmount = modulesAmount.Add(macc.GetCoins())
			} else {
				// basic or vesting accounts
				circulatingAmount = circulatingAmount.Add(acc.GetCoins())
			}

			return false
		})

		if !supplier.CirculatingSupply.IsEqual(circulatingAmount) {
			return fmt.Errorf("circulating supply invariance:\n"+
				"\tsupplier.CirculatingSupply: %v\n"+
				"\tsum of circulating tokens: %v", supplier.CirculatingSupply, circulatingAmount)
		}

		if !supplier.ModulesSupply.IsEqual(modulesAmount) {
			return fmt.Errorf("modules holdings supply invariance:\n"+
				"\tsupplier.ModulesSupply: %v\n"+
				"\tsum of modules accounts tokens: %v", supplier.ModulesSupply, modulesAmount)
		}

		if !supplier.InitialVestingSupply.IsEqual(initialVestingAmount) {
			return fmt.Errorf("vesting supply invariance:\n"+
				"\tsupplier.InitialVestingSupply: %v\n"+
				"\tsum of vesting tokens: %v", supplier.InitialVestingSupply, initialVestingAmount)
		}

		bondedSupply := sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), stakingKeeper.TotalBondedTokens(ctx)))
		collectedFees := feeCollectionKeeper.GetCollectedFees(ctx)
		communityPool, remainingCommunityPool := distributionKeeper.GetFeePoolCommunityCoins(ctx).TruncateDecimal()
		totalRewards, remainingRewards := distributionKeeper.GetTotalRewards(ctx).TruncateDecimal()

		remaining, _ := remainingCommunityPool.Add(remainingRewards).TruncateDecimal()

		expectedTotalSupply := supplier.CirculatingSupply.
			Add(supplier.ModulesSupply).
			Add(bondedSupply).
			Add(collectedFees).
			Add(communityPool).
			Add(totalRewards).
			Add(remaining)

		if !supplier.TotalSupply.IsEqual(expectedTotalSupply) {
			return fmt.Errorf("total supply invariance:\n"+
				"\texpected total supply: %v\n"+
				"\treal total supply: %v", expectedTotalSupply, supplier.TotalSupply)
		}

		return nil
	}
}
