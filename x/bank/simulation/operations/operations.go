package operations

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
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper, bk bank.Keeper) simulation.WeightedOperations {

	var weightMsgSend, weightSingleInputMsgMultiSend int

	appParams.GetOrGenerate(cdc, OpWeightMsgSend, &weightMsgSend, nil,
		func(_ *rand.Rand) { weightMsgSend = 100 })

	appParams.GetOrGenerate(cdc, OpWeightSingleInputMsgMultiSend, &weightSingleInputMsgMultiSend, nil,
		func(_ *rand.Rand) { weightSingleInputMsgMultiSend = 10 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgSend,
			SimulateMsgSend(ak, bk),
		),
		simulation.NewWeigthedOperation(
			weightSingleInputMsgMultiSend,
			SimulateSingleInputMsgMultiSend(ak, bk),
		),
	}
}
