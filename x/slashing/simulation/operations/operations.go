package operations

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

// Simulation parameter constants
const (
	OpWeightMsgUnjail = "op_weight_msg_unjail"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, keeper slashing.Keeper) simulation.WeightedOperations {
	var weightMsgUnjail int

	appParams.GetOrGenerate(cdc, OpWeightMsgUnjail, &weightMsgUnjail, nil,
		func(_ *rand.Rand) { weightMsgUnjail = 100 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgUnjail,
			SimulateMsgUnjail(keeper),
		),
	}
}
