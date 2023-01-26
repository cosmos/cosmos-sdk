package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
//
//nolint:gosec // these are not hardcoded credentials.
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams = "op_weight_msg_update_params"
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	params.DefaultSendEnabled = r.Intn(2) == 0
	if r.Intn(2) == 0 {
		params.SendEnabled = nil
	} else {
		params.SendEnabled = make([]*types.SendEnabled, 10)
		for i := 0; i < r.Intn(10); i++ {
			params.SendEnabled[i] = types.NewSendEnabled(
				simtypes.RandStringOfLength(r, 10),
				r.Intn(2) == 0,
			)
		}
	}

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}
