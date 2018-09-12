package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
	abci "github.com/tendermint/tendermint/abci/types"
)

// AllInvariants runs all invariants of the stake module.
// Currently: total supply, positive power
func AllInvariants(ck bank.Keeper, k stake.Keeper, am auth.AccountMapper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		err := SupplyInvariants(ck, k, am)(app)
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

// SupplyInvariants checks that the total supply reflects all held loose tokens, bonded tokens, and unbonding delegations
// nolint: unparam
func SupplyInvariants(ck bank.Keeper, k stake.Keeper, am auth.AccountMapper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, abci.Header{})
		//pool := k.GetPool(ctx)

		loose := sdk.ZeroInt()
		bonded := sdk.ZeroDec()
		am.IterateAccounts(ctx, func(acc auth.Account) bool {
			loose = loose.Add(acc.GetCoins().AmountOf("steak"))
			return false
		})
		k.IterateUnbondingDelegations(ctx, func(_ int64, ubd stake.UnbondingDelegation) bool {
			loose = loose.Add(ubd.Balance.Amount)
			return false
		})
		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetPower())
			case sdk.Unbonding:
				loose = loose.Add(validator.GetTokens().RoundInt())
			case sdk.Unbonded:
				loose = loose.Add(validator.GetTokens().RoundInt())
			}
			return false
		})

		// Loose tokens should equal coin supply plus unbonding delegations plus tokens on unbonded validators
		// XXX TODO https://github.com/cosmos/cosmos-sdk/issues/2063#issuecomment-413720872
		// require.True(t, pool.LooseTokens.RoundInt64() == loose.Int64(), "expected loose tokens to equal total steak held by accounts - pool.LooseTokens: %v, sum of account tokens: %v\nlog: %s",
		//	 pool.LooseTokens.RoundInt64(), loose.Int64(), log)

		// Bonded tokens should equal sum of tokens with bonded validators
		// XXX TODO https://github.com/cosmos/cosmos-sdk/issues/2063#issuecomment-413720872
		// require.True(t, pool.BondedTokens.RoundInt64() == bonded.RoundInt64(), "expected bonded tokens to equal total steak held by bonded validators - pool.BondedTokens: %v, sum of bonded validator tokens: %v\nlog: %s",
		//   pool.BondedTokens.RoundInt64(), bonded.RoundInt64(), log)

		// TODO Inflation check on total supply
		return nil
	}
}

// PositivePowerInvariant checks that all stored validators have > 0 power
func PositivePowerInvariant(k stake.Keeper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, abci.Header{})
		var err error
		k.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) bool {
			if !validator.GetPower().GT(sdk.ZeroDec()) {
				err = fmt.Errorf("validator with non-positive power stored. (pubkey %v)", validator.GetPubKey())
				return true
			}
			return false
		})
		return err
	}
}

// ValidatorSetInvariant checks equivalence of Tendermint validator set and SDK validator set
func ValidatorSetInvariant(k stake.Keeper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		// TODO
		return nil
	}
}
