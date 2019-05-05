package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants register all supply invariants
func RegisterInvariants(ck CrisisKeeper, k SupplyKeeper, sk StakingKeeper) {
	ck.RegisterRoute(ModuleName, "staking-tokens",
		StakingTokensInvariant(k, sk))
}

// AllInvariants runs all invariants of the supply module.
func AllInvariants(k SupplyKeeper, sk StakingKeeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		err := StakingTokensInvariant(k, sk)(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

// StakingTokensInvariant checks that the total supply of staking tokens from the staking pool matches the total supply ones
func StakingTokensInvariant(k SupplyKeeper, sk StakingKeeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		supply := k.GetSupply(ctx)
		totalStakeSupply := supply.Total.AmountOf(sk.BondDenom(ctx))
		expectedTotalStakeSupply := sk.StakingTokenSupply(ctx)

		// Bonded tokens should equal sum of tokens with bonded validators
		if !totalStakeSupply.Equal(expectedTotalStakeSupply) {
			return fmt.Errorf("total staking token supply invariance:\n"+
				"\tstaking pool tokens: %v\n"+
				"\tsupply staking tokens: %v", expectedTotalStakeSupply, totalStakeSupply)
		}

		return nil
	}
}
