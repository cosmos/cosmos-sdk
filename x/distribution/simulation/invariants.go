package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

// AllInvariants runs all invariants of the distribution module
// Currently: total supply, positive power
func AllInvariants(d distr.Keeper, sk distr.StakeKeeper) simulation.Invariant {
	return func(app *baseapp.BaseApp) error {
		err := ValAccumInvariants(d, sk)(app)
		if err != nil {
			return err
		}
		err = DelAccumInvariants(d, sk)(app)
		if err != nil {
			return err
		}
		return nil
	}
}

// ValAccumInvariants checks that the fee pool accum == sum all validators' accum
func ValAccumInvariants(k distr.Keeper, sk distr.StakeKeeper) simulation.Invariant {

	return func(app *baseapp.BaseApp) error {
		mockHeader := abci.Header{Height: app.LastBlockHeight() + 1}
		ctx := app.NewContext(false, mockHeader)
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

	return func(app *baseapp.BaseApp) error {
		mockHeader := abci.Header{Height: app.LastBlockHeight() + 1}
		ctx := app.NewContext(false, mockHeader)
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
