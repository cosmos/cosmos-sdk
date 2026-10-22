package simulation

import (
	"context"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants.
const (
	DefaultWeightMsgUpdateParams int = 100
	OpWeightMsgUpdateParams          = "op_weight_msg_update_params"
)

// scopeChoices are the patterns a random scope draws from.
var scopeChoices = []string{
	types.Wildcard,
	"/cosmos.*",
	"/cosmos.bank.*",
	"/cosmos.staking.*",
	"/cosmos.bank.v1beta1.MsgSend",
}

// ProposalMsgs returns the x/safety msgs used for simulated gov proposals.
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a MsgUpdateParams with random params.
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, accs []simtypes.Account) sdk.Msg {
	gov := sdk.AccAddress(address.Module("gov")).String()
	return &types.MsgUpdateParams{Authority: gov, Params: RandomParams(r, gov, accs)}
}

// MsgUpdateParamsFactory returns a factory for MsgUpdateParams with random
// params, submitted through gov.
func MsgUpdateParamsFactory() simsx.SimMsgFactoryFn[*types.MsgUpdateParams] {
	return func(_ context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *types.MsgUpdateParams) {
		gov := testData.ModuleAccountAddress(reporter, "gov")
		return nil, &types.MsgUpdateParams{
			Authority: gov,
			Params:    RandomParams(testData.Rand().Rand, gov, testData.AllAccounts()),
		}
	}
}

// RandomParams returns valid params that keep authority, with a random
// operator from accs or none, and a random scope. Only MsgTrip disables msgs,
// so no params change stalls other modules.
func RandomParams(r *rand.Rand, authority string, accs []simtypes.Account) types.Params {
	params := types.DefaultParams(authority)
	if len(accs) > 0 && r.Intn(2) == 0 {
		params.Operator = accs[r.Intn(len(accs))].Address.String()
	}
	for _, p := range scopeChoices {
		if r.Intn(2) == 0 {
			params.Scope = append(params.Scope, p)
		}
	}
	return params
}
