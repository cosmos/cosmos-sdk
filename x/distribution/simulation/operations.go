package simulation

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	OpWeightMsgSetWithdrawAddress          = "op_weight_msg_set_withdraw_address"
	OpWeightMsgWithdrawDelegationReward    = "op_weight_msg_withdraw_delegation_reward"
	OpWeightMsgWithdrawValidatorCommission = "op_weight_msg_withdraw_validator_commission"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(cdc *codec.Codec, ak types.AccountKeeper, keeper distr.Keeper) simulation.WeightedOperations {
	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgSetWithdrawAddress, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			SimulateMsgSetWithdrawAddress(ak, keeper),
		),
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgWithdrawDelegationReward, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			SimulateMsgWithdrawDelegatorReward(ak, keeper),
		),
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgWithdrawValidatorCommission, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			SimulateMsgWithdrawValidatorCommission(ak, keeper),
		),
	}
}
