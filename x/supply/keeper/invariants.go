package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// names used as root for pool module accounts
const (
	UnbondedTokensName = "UnbondedTokens"
	BondedTokensName   = "BondedTokens"
)

// RegisterInvariants register all supply invariants
func RegisterInvariants(ck CrisisKeeper, k Keeper, dk DistributionKeeper,
	sk StakingKeeper) {
	ck.RegisterRoute(ModuleName, "total-supply",
		TotalSupply(k, dk, sk))
	ck.RegisterRoute(ModuleName, "staking-token-supply",
		StakingTokensInvariant(k, sk))
}

// AllInvariants runs all invariants of the supply module.
func AllInvariants(k Keeper, dk DistributionKeeper, sk StakingKeeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		err := StakingTokensInvariant(k, sk)(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

// BondedTokensInvariants checks that the total supply reflects all held not-bonded tokens, bonded tokens, and unbonding delegations
func TotalSupply(k Keeper, dk DistributionKeeper, sk StakingKeeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		total := sdk.ZeroInt()

		k.ak.IterateAccounts(ctx, func(acc auth.Account) bool {
			total = total.Add(acc.GetCoins().AmountOf(sk.BondDenom(ctx)))
			return false
		})

		sk.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				total = total.Add(validator.GetBondedTokens())
			case sdk.Unbonding, sdk.Unbonded:
				total = total.Add(validator.GetTokens())
			}
			// add yet-to-be-withdrawn
			total = total.Add(dk.GetValidatorOutstandingRewardsCoins(ctx, validator.GetOperator()).AmountOf(sk.BondDenom(ctx)))
			return false
		})

		return nil
	}
}

// StakingTokensInvariant checks that the total supply of staking tokens from the staking pool matches the total supply ones
func StakingTokensInvariant(k Keeper, sk StakingKeeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		supply := k.GetSupply(ctx)
		bondDenom := sk.BondDenom(ctx)

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
