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
