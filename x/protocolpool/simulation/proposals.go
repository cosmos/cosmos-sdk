package simulation

import (
	"math/rand"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 50

	OpWeightMsgUpdateParams = "op_weight_msg_update_params"

	DefaultWeightCreateContinuousFund int = 50

	OpWeightMsgCreateContinuousFund = "op_weight_msg_create_continuous_fund"
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
		simulation.NewWeightedProposalMsg(
			OpWeightMsgCreateContinuousFund,
			DefaultWeightCreateContinuousFund,
			SimulateMsgCreateContinuousFund,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    GenParams(r),
	}
}

// SimulateMsgCreateContinuousFund returns a random MsgCreateContinuousFund
func SimulateMsgCreateContinuousFund(r *rand.Rand, _ sdk.Context, accs []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")
	simAccount, _ := simtypes.RandomAcc(r, accs)
	percentage := math.LegacyNewDec(int64(r.Intn(10) + 1)).Quo(math.LegacyNewDec(100))

	return &types.MsgCreateContinuousFund{
		Authority:  authority.String(),
		Recipient:  simAccount.Address.String(),
		Percentage: percentage,
	}
}
