package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
	abci "github.com/tendermint/tendermint/abci/types"
)

// AllInvariants runs all invariants of the stake module.
// Currently: total supply, positive power
func AllInvariants(ck bank.Keeper, k stake.Keeper,
	f auth.FeeCollectionKeeper, d distribution.Keeper,
	am auth.AccountKeeper) simulation.Invariant {

	return func(app *baseapp.BaseApp) error {
		err := SupplyInvariants(ck, k, f, d, am)(app)
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
func SupplyInvariants(ck bank.Keeper, k stake.Keeper,
	f auth.FeeCollectionKeeper, d distribution.Keeper, am auth.AccountKeeper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, abci.Header{})
		pool := k.GetPool(ctx)

		loose := sdk.ZeroDec()
		bonded := sdk.ZeroDec()
		am.IterateAccounts(ctx, func(acc auth.Account) bool {
			loose = loose.Add(sdk.NewDecFromInt(acc.GetCoins().AmountOf("steak")))
			return false
		})
		k.IterateUnbondingDelegations(ctx, func(_ int64, ubd stake.UnbondingDelegation) bool {
			loose = loose.Add(sdk.NewDecFromInt(ubd.Balance.Amount))
			return false
		})
		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetPower())
			case sdk.Unbonding:
				loose = loose.Add(validator.GetTokens())
			case sdk.Unbonded:
				loose = loose.Add(validator.GetTokens())
			}
			return false
		})

		feePool := d.GetFeePool(ctx)

		// add outstanding fees
		loose = loose.Add(sdk.NewDecFromInt(f.GetCollectedFees(ctx).AmountOf("steak")))

		// add community pool
		loose = loose.Add(feePool.CommunityPool.AmountOf("steak"))

		// add validator distribution pool
		loose = loose.Add(feePool.ValPool.AmountOf("steak"))

		// add validator distribution commission and yet-to-be-withdrawn-by-delegators
		d.IterateValidatorDistInfos(ctx,
			func(_ int64, distInfo distribution.ValidatorDistInfo) (stop bool) {
				loose = loose.Add(distInfo.DelPool.AmountOf("steak"))
				loose = loose.Add(distInfo.ValCommission.AmountOf("steak"))
				return false
			},
		)

		// Loose tokens should equal coin supply plus unbonding delegations
		// plus tokens on unbonded validators
		if !pool.LooseTokens.Equal(loose) {
			return fmt.Errorf("loose token invariance:\n\tpool.LooseTokens: %v"+
				"\n\tsum of account tokens: %v", pool.LooseTokens, loose)
		}

		// Bonded tokens should equal sum of tokens with bonded validators
		if !pool.BondedTokens.Equal(bonded) {
			return fmt.Errorf("bonded token invariance:\n\tpool.BondedTokens: %v"+
				"\n\tsum of account tokens: %v", pool.BondedTokens, bonded)
		}

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
				err = fmt.Errorf("validator with non-positive power stored. (pubkey %v)", validator.GetConsPubKey())
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
