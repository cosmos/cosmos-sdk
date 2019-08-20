package operations

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation parameter constants
const (
	OpWeightMsgCreateValidator = "op_weight_msg_create_validator"
	OpWeightMsgEditValidator   = "op_weight_msg_edit_validator"
	OpWeightMsgDelegate        = "op_weight_msg_delegate"
	OpWeightMsgUndelegate      = "op_weight_msg_undelegate"
	OpWeightMsgBeginRedelegate = "op_weight_msg_begin_redelegate"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper, keeper staking.Keeper) simulation.WeightedOperations {

	var (
		weightMsgCreateValidator int
		weightMsgEditValidator   int
		weightMsgDelegate        int
		weightMsgUndelegate      int
		weightMsgBeginRedelegate int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil,
		func(_ *rand.Rand) { weightMsgCreateValidator = 100 })

	appParams.GetOrGenerate(cdc, OpWeightMsgEditValidator, &weightMsgEditValidator, nil,
		func(_ *rand.Rand) { weightMsgEditValidator = 5 })

	appParams.GetOrGenerate(cdc, OpWeightMsgDelegate, &weightMsgDelegate, nil,
		func(_ *rand.Rand) { weightMsgDelegate = 100 })

	appParams.GetOrGenerate(cdc, OpWeightMsgUndelegate, &weightMsgUndelegate, nil,
		func(_ *rand.Rand) { weightMsgUndelegate = 100 })

	appParams.GetOrGenerate(cdc, OpWeightMsgBeginRedelegate, &weightMsgBeginRedelegate, nil,
		func(_ *rand.Rand) { weightMsgBeginRedelegate = 100 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateValidator(ak, keeper),
		),
		simulation.NewWeigthedOperation(
			weightMsgEditValidator,
			SimulateMsgEditValidator(keeper),
		),
		simulation.NewWeigthedOperation(
			weightMsgDelegate,
			SimulateMsgDelegate(ak, keeper),
		),
		simulation.NewWeigthedOperation(
			weightMsgUndelegate,
			SimulateMsgUndelegate(ak, keeper),
		),
		simulation.NewWeigthedOperation(
			weightMsgBeginRedelegate,
			SimulateMsgBeginRedelegate(ak, keeper),
		),
	}
}
