package keeper

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterInvariants registers all staking invariants
func RegisterInvariants(ir sdk.InvariantRouter, k Keeper) {

	ir.RegisterRoute(types.ModuleName, "bonded-tokens",
		BondedTokensInvariant(k))
	ir.RegisterRoute(types.ModuleName, "nonnegative-power",
		NonNegativePowerInvariant(k))
	ir.RegisterRoute(types.ModuleName, "positive-delegation",
		PositiveDelegationInvariant(k))
	ir.RegisterRoute(types.ModuleName, "delegator-shares",
		DelegatorSharesInvariant(k))
}

// AllInvariants runs all invariants of the staking module.
func AllInvariants(k Keeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		err := BondedTokensInvariant(k)(ctx)
		if err != nil {
			return err
		}

		err = NonNegativePowerInvariant(k)(ctx)
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

// BondedTokensInvariant checks that the bonded pool account reflects the tokens actively bonded
func BondedTokensInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		bonded := sdk.ZeroInt()
		bondedPool, err := k.supplyKeeper.GetPoolAccountByName(ctx, BondedTokensName)
		if err != nil {
			return err
		}

		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetBondedTokens())
			}
			return false
		})

		expectedBonded := bondedPool.GetCoins().AmountOf(k.BondDenom(ctx))

		// Bonded tokens should equal sum of tokens with bonded validators
		if !expectedBonded.Equal(bonded) {
			return fmt.Errorf("bonded token invariance:\n"+
				"\tPool's bonded tokens: %v\n"+
				"\tsum of bonded tokens from delegations: %v", expectedBonded, bonded)
		}

		return nil
	}
}

// NonNegativePowerInvariant checks that all stored validators have >= 0 power.
func NonNegativePowerInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		iterator := k.ValidatorsPowerStoreIterator(ctx)

		for ; iterator.Valid(); iterator.Next() {
			validator, found := k.GetValidator(ctx, iterator.Value())
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %X\n", iterator.Value()))
			}

			powerKey := GetValidatorsByPowerIndexKey(validator)

			if !bytes.Equal(iterator.Key(), powerKey) {
				return fmt.Errorf("power store invariance:\n\tvalidator.Power: %v"+
					"\n\tkey should be: %v\n\tkey in store: %v",
					validator.GetTendermintPower(), powerKey, iterator.Key())
			}

			if validator.Tokens.IsNegative() {
				return fmt.Errorf("negative tokens for validator: %v", validator)
			}
		}
		iterator.Close()
		return nil
	}
}

// PositiveDelegationInvariant checks that all stored delegations have > 0 shares.
func PositiveDelegationInvariant(k Keeper) sdk.Invariant {
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
func DelegatorSharesInvariant(k Keeper) sdk.Invariant {
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
