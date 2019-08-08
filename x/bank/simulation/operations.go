package simulation

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	OpWeightMsgSend                 = "op_weight_msg_send"
	OpWeightSingleInputMsgMultiSend = "op_weight_single_input_msg_multisend"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(cdc *codec.Codec, ak types.AccountKeeper, bk bank.Keeper) simulation.WeightedOperations {
	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgSend, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			SimulateMsgSend(ak, bk),
		),
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSingleInputMsgMultiSend, &v, nil,
					func(_ *rand.Rand) {
						v = 10
					})
				return v
			}(nil),
			SimulateSingleInputMsgMultiSend(ak, bk),
		),
	}
}
