package simulation

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// AllInvariants runs all invariants of the distribution module
// Currently: total supply, positive power
func AllInvariants(d distr.Keeper, stk stake.Keeper) simulation.Invariant {
	sk := distr.StakeKeeper(stk)
	return func(ctx sdk.Context) error {
		err := ValAccumInvariants(d, sk)(ctx)
		if err != nil {
			return err
		}
		err = DelAccumInvariants(d, sk)(ctx)
		if err != nil {
			return err
		}
		err = CanWithdrawInvariant(d, stk)(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

// ValAccumInvariants checks that the fee pool accum == sum all validators' accum
func ValAccumInvariants(k distr.Keeper, sk distr.StakeKeeper) simulation.Invariant {

	return func(ctx sdk.Context) error {
		height := ctx.BlockHeight()

		valAccum := sdk.ZeroDec()
		k.IterateValidatorDistInfos(ctx, func(_ int64, vdi distr.ValidatorDistInfo) bool {
			lastValPower := sk.GetLastValidatorPower(ctx, vdi.OperatorAddr)
			valAccum = valAccum.Add(vdi.GetValAccum(height, sdk.NewDecFromInt(lastValPower)))
			return false
		})

		lastTotalPower := sdk.NewDecFromInt(sk.GetLastTotalPower(ctx))
		totalAccum := k.GetFeePool(ctx).GetTotalValAccum(height, lastTotalPower)

		if !totalAccum.Equal(valAccum) {
			return fmt.Errorf("validator accum invariance: \n\tfee pool totalAccum: %v"+
				"\n\tvalidator accum \t%v\n", totalAccum.String(), valAccum.String())
		}

		return nil
	}
}

// DelAccumInvariants checks that each validator del accum == sum all delegators' accum
func DelAccumInvariants(k distr.Keeper, sk distr.StakeKeeper) simulation.Invariant {

	return func(ctx sdk.Context) error {
		height := ctx.BlockHeight()

		totalDelAccumFromVal := make(map[string]sdk.Dec) // key is the valOpAddr string
		totalDelAccum := make(map[string]sdk.Dec)

		// iterate the validators
		iterVal := func(_ int64, vdi distr.ValidatorDistInfo) bool {
			key := vdi.OperatorAddr.String()
			validator := sk.Validator(ctx, vdi.OperatorAddr)
			totalDelAccumFromVal[key] = vdi.GetTotalDelAccum(height,
				validator.GetDelegatorShares())

			// also initialize the delegation map
			totalDelAccum[key] = sdk.ZeroDec()

			return false
		}
		k.IterateValidatorDistInfos(ctx, iterVal)

		// iterate the delegations
		iterDel := func(_ int64, ddi distr.DelegationDistInfo) bool {
			key := ddi.ValOperatorAddr.String()
			delegation := sk.Delegation(ctx, ddi.DelegatorAddr, ddi.ValOperatorAddr)
			totalDelAccum[key] = totalDelAccum[key].Add(
				ddi.GetDelAccum(height, delegation.GetShares()))
			return false
		}
		k.IterateDelegationDistInfos(ctx, iterDel)

		// compare
		for key, delAccumFromVal := range totalDelAccumFromVal {
			sumDelAccum := totalDelAccum[key]

			if !sumDelAccum.Equal(delAccumFromVal) {

				logDelAccums := ""
				iterDel := func(_ int64, ddi distr.DelegationDistInfo) bool {
					keyLog := ddi.ValOperatorAddr.String()
					if keyLog == key {
						delegation := sk.Delegation(ctx, ddi.DelegatorAddr, ddi.ValOperatorAddr)
						accum := ddi.GetDelAccum(height, delegation.GetShares())
						if accum.IsPositive() {
							logDelAccums += fmt.Sprintf("\n\t\tdel: %v, accum: %v",
								ddi.DelegatorAddr.String(),
								accum.String())
						}
					}
					return false
				}
				k.IterateDelegationDistInfos(ctx, iterDel)

				operAddr, err := sdk.ValAddressFromBech32(key)
				if err != nil {
					panic(err)
				}
				validator := sk.Validator(ctx, operAddr)

				return fmt.Errorf("delegator accum invariance: \n"+
					"\tvalidator key: %v\n"+
					"\tvalidator: %+v\n"+
					"\tsum delegator accum: %v\n"+
					"\tvalidator's total delegator accum: %v\n"+
					"\tlog of delegations with accum: %v\n",
					key, validator, sumDelAccum.String(),
					delAccumFromVal.String(), logDelAccums)
			}
		}

		return nil
	}
}

// CanWithdrawInvariant checks that current rewards can be completely withdrawn
func CanWithdrawInvariant(k distr.Keeper, sk stake.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		// we don't want to write the changes
		ctx, _ = ctx.CacheContext()

		// withdraw all delegator & validator rewards
		vdiIter := func(_ int64, valInfo distr.ValidatorDistInfo) (stop bool) {
			err := k.WithdrawValidatorRewardsAll(ctx, valInfo.OperatorAddr)
			if err != nil {
				panic(err)
			}
			return false
		}
		k.IterateValidatorDistInfos(ctx, vdiIter)

		ddiIter := func(_ int64, distInfo distr.DelegationDistInfo) (stop bool) {
			err := k.WithdrawDelegationReward(
				ctx, distInfo.DelegatorAddr, distInfo.ValOperatorAddr)
			if err != nil {
				panic(err)
			}
			return false
		}
		k.IterateDelegationDistInfos(ctx, ddiIter)

		// assert that the fee pool is empty
		feePool := k.GetFeePool(ctx)
		if !feePool.TotalValAccum.Accum.IsZero() {
			return fmt.Errorf("unexpected leftover validator accum")
		}
		bondDenom := sk.GetParams(ctx).BondDenom
		if !feePool.ValPool.AmountOf(bondDenom).IsZero() {
			return fmt.Errorf("unexpected leftover validator pool coins: %v",
				feePool.ValPool.AmountOf(bondDenom).String())
		}

		// all ok
		return nil
	}
}
