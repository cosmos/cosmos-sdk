package keeper

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
)

// RegisterInvariants registers all poa invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {

	ir.RegisterRoute(types.ModuleName, "nonnegative-power",
		NonNegativePowerInvariant(k))
}

// AllInvariants runs all invariants of the poa module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := NonNegativePowerInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		return "", false
	}
}

// NonNegativePowerInvariant checks that all stored validators have >= 0 power.
func NonNegativePowerInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var broken bool

		iterator := k.ValidatorsPowerStoreIterator(ctx)

		for ; iterator.Valid(); iterator.Next() {
			validator, found := k.GetValidator(ctx, iterator.Value())
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %X\n", iterator.Value()))
			}

			powerKey := types.GetValidatorsByPowerIndexKey(validator)

			if !bytes.Equal(iterator.Key(), powerKey) {
				broken = true
				msg += fmt.Sprintf("power store invariance:\n\tvalidator.Power: %v"+
					"\n\tkey should be: %v\n\tkey in store: %v\n",
					validator.GetConsensusPower(), powerKey, iterator.Key())
			}

			if validator.Weight.IsNegative() {
				broken = true
				msg += fmt.Sprintf("\tnegative tokens for validator: %v\n", validator)
			}
		}
		iterator.Close()
		return sdk.FormatInvariant(types.ModuleName, "nonnegative power", fmt.Sprintf("found invalid validator powers\n%s", msg)), broken
	}
}
