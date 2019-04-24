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

		expectedTotalSupply := k.TotalSupply(ctx)

		if !supplier.TotalSupply.IsEqual(expectedTotalSupply) {
			return fmt.Errorf("total supply invariance:\n"+
				"\texpected total supply: %v\n"+
				"\treal total supply: %v", expectedTotalSupply, supplier.TotalSupply)
		}

		return nil
	}
}
