package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
)

// AllInvariants runs all invariants of the stake module.
// Currently: total supply, positive power
func AllInvariants(ck bank.Keeper, k stake.Keeper,
	f auth.FeeCollectionKeeper, d distribution.Keeper,
	am auth.AccountKeeper) simulation.Invariant {

	return func(app *baseapp.BaseApp) error {
		err := BondedAmountInvariants(ck, k, f, d, am)(app)
		if err != nil {
			return err
		}

		err = PositivePowerInvariant(k)(app)
		if err != nil {
			return err
		}

		err = ValidatorSetInvariant(k)(app)
		return err
	}
}

// BondedAmountInvariants checks that the logged amount of bonded tokens reflects the amount held in validators.
// nolint: unparam
func BondedAmountInvariants(ck bank.Keeper, k stake.Keeper,
	f auth.FeeCollectionKeeper, d distribution.Keeper, am auth.AccountKeeper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, abci.Header{})
		pool := k.GetPool(ctx)

		bonded := sdk.ZeroDec()
		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			if validator.GetStatus() == sdk.Bonded {
				bonded = bonded.Add(validator.GetPower())
			}
			return false
		})

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
	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, abci.Header{})

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

// ValidatorSetInvariant checks equivalence of Tendermint validator set and SDK validator set
func ValidatorSetInvariant(k stake.Keeper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		// TODO
		return nil
	}
}
