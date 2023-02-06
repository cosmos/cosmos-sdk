package simulation

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams = "op_weight_msg_update_params" //nolint:gosec
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
	params.DowntimeJailDuration = time.Duration(simtypes.RandTimestamp(r).UnixNano())
	params.SignedBlocksWindow = int64(simtypes.RandIntBetween(r, 1, 1000))
	params.MinSignedPerWindow = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 100)), 2)
	params.SlashFractionDoubleSign = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 100)), 2)
	params.SlashFractionDowntime = sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 100)), 2)

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}
