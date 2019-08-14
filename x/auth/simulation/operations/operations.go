package operations

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	OpWeightDeductFee = "op_weight_deduct_fee"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(cdc *codec.Codec, ak auth.AccountKeeper, supplyKeeper types.SupplyKeeper) simulation.WeightedOperations {
	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightDeductFee, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			SimulateDeductFee(ak, supplyKeeper),
		),
	}
}
