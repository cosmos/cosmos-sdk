package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// staking pools name; importing them introduces circle dependencies
const (
	UnbondedTokensName = "UnbondedTokens"
	BondedTokensName   = "BondedTokens"
)

// RegisterInvariants register all supply invariants
func RegisterInvariants(ck CrisisKeeper, k Keeper, bondDenom string) {
	ck.RegisterRoute(ModuleName, "staking-tokens",
		StakingTokensInvariant(k, bondDenom))
}

// AllInvariants runs all invariants of the supply module.
func AllInvariants(k Keeper, bondDenom string) sdk.Invariant {
	return func(ctx sdk.Context) error {
		err := StakingTokensInvariant(k, bondDenom)(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

// StakingTokensInvariant checks that the total supply of staking tokens from the staking pool matches the total supply ones
func StakingTokensInvariant(k Keeper, bondDenom string) sdk.Invariant {
	return func(ctx sdk.Context) error {
		supply := k.GetSupply(ctx)
		totalStakeSupply := supply.Total.AmountOf(bondDenom)
		bondedPool, err := k.GetAccountByName(ctx, BondedTokensName)
		if err != nil {
			return err
		}
		unbondedPool, err := k.GetAccountByName(ctx, UnbondedTokensName)
		if err != nil {
			return err
		}

		stakeSupply := bondedPool.GetCoins().AmountOf(bondDenom).Add(unbondedPool.GetCoins().AmountOf(bondDenom))

		// Bonded tokens should equal sum of tokens with bonded validators
		if !totalStakeSupply.Equal(stakeSupply) {
			return fmt.Errorf("total staking token supply invariance:\n"+
				"\tstaking pool tokens: %v\n"+
				"\tsupply staking tokens: %v", stakeSupply, totalStakeSupply)
		}

		return nil
	}
}
