package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

// AllInvariants runs all invariants of the distribution module
// Currently: total supply, positive power
func AllInvariants(
	d distribution.Keeper, height int64) simulation.Invariant {

	return func(app *baseapp.BaseApp) error {
		err := ValAccumInvariants(ck, k, f, d, am)(app)
		if err != nil {
			return err
		}
	}
}

// SupplyInvariants checks that the total supply reflects all held loose tokens, bonded tokens, and unbonding delegations
// nolint: unparam
func ValAccumInvariants(d distribution.Keeper, header abci.Header) simulation.Invariant {

	return func(app *baseapp.BaseApp) error {
		ctx := app.NewContext(false, header)

		valAccum := int64(0)
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

		//totalBondedTokens :=
		totalAccum := d.GetFeePool(ctx).GetTotalValAccum(ctx.GetHeight(), totalBondedTokens)

		if totalAccum != valAccum {
			fmt.Errorf("validator accum invariance: \n\tfee pool totalAccum: %v"+
				"\n\tvalidator accum \t%v\n", totalAccum, valAccum)
		}

		return nil
	}
}
