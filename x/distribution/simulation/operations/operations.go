package operations

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	OpWeightMsgSetWithdrawAddress          = "op_weight_msg_set_withdraw_address"
	OpWeightMsgWithdrawDelegationReward    = "op_weight_msg_withdraw_delegation_reward"
	OpWeightMsgWithdrawValidatorCommission = "op_weight_msg_withdraw_validator_commission"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, k keeper.Keeper) simulation.WeightedOperations {

	var (
		weightMsgSetWithdrawAddress          int
		weightMsgWithdrawDelegationReward    int
		weightMsgWithdrawValidatorCommission int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgSetWithdrawAddress, &weightMsgSetWithdrawAddress, nil,
		func(_ *rand.Rand) { weightMsgSetWithdrawAddress = 50 })

	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawDelegationReward, &weightMsgWithdrawDelegationReward, nil,
		func(_ *rand.Rand) { weightMsgWithdrawDelegationReward = 50 })

	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawValidatorCommission, &weightMsgWithdrawValidatorCommission, nil,
		func(_ *rand.Rand) { weightMsgWithdrawValidatorCommission = 50 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgSetWithdrawAddress,
			SimulateMsgSetWithdrawAddress(k),
		),
		simulation.NewWeigthedOperation(
			weightMsgWithdrawDelegationReward,
			SimulateMsgWithdrawDelegatorReward(k),
		),
		simulation.NewWeigthedOperation(
			weightMsgWithdrawValidatorCommission,
			SimulateMsgWithdrawValidatorCommission(k),
		),
	}
}
