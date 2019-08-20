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
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak auth.AccountKeeper, supplyKeeper types.SupplyKeeper) simulation.WeightedOperations {

	var weightDeductFee int

	appParams.GetOrGenerate(cdc, OpWeightDeductFee, &weightDeductFee, nil,
		func(_ *rand.Rand) { weightDeductFee = 5 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightDeductFee,
			SimulateDeductFee(ak, supplyKeeper),
		),
	}
}
