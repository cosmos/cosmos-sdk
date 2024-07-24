package simulation

import (
	"context"
	"math/rand"
	"time"

	coreaddress "cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams = "op_weight_msg_update_params"
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsgX(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(_ context.Context, r *rand.Rand, _ []simtypes.Account, addressCodec coreaddress.Codec) (sdk.Msg, error) {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	params.HistoricalEntries = uint32(simtypes.RandIntBetween(r, 0, 1000))
	params.MaxEntries = uint32(simtypes.RandIntBetween(r, 1, 1000))
	params.MaxValidators = uint32(simtypes.RandIntBetween(r, 1, 1000))
	params.UnbondingTime = time.Duration(simtypes.RandTimestamp(r).UnixNano())

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}, nil
}
