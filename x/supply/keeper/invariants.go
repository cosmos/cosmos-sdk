package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants register all supply invariants
func RegisterInvariants(ck CrisisKeeper, k Keeper, sk StakingKeeper) {

	ck.RegisterRoute(ModuleName, "staking-tokens",
		StakingTokensInvariant(k, sk))
}

// AllInvariants runs all invariants of the supply module.
func AllInvariants(k Keeper, sk StakingKeeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		err := StakingTokensInvariant(k, sk)(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

// StakingTokensInvariant checks that the total supply of staking tokens from the staking pool matches the total supply ones
func StakingTokensInvariant(k Keeper, sk StakingKeeper) sdk.Invariant {

	return func(ctx sdk.Context) error {

		totalSupplyOfStake := k.TotalSupply(ctx).AmountOf(sk.BondDenom(ctx))
		stakeCoins := sk.StakingTokenSupply(ctx)

		// Bonded tokens should equal sum of tokens with bonded validators
		if !totalSupplyOfStake.Equal(stakeCoins) {
			return fmt.Errorf("total staking token supply invariance:\n"+
				"\tstaking staking tokens: %v\n"+
				"\tsupply staking tokens: %v", totalSupplyOfStake, stakeCoins)
		}

		return nil
	}
}
