package simulation

import (
	"context"
	"math/rand"
	"time"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
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
	// changes to MinCommissionRate or BondDenom create issues for in flight messages or state operations

	addr, err := addressCodec.BytesToString(authority)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParams{
		Authority: addr,
		Params:    params,
	}, nil
}
