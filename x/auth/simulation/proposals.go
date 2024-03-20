package simulation

import (
	"math/rand"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/x/auth/types"

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
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ []simtypes.Account, _ coreaddress.Codec) (sdk.Msg, error) {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	params.MaxMemoCharacters = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.TxSigLimit = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.TxSizeCostPerByte = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.SigVerifyCostED25519 = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.SigVerifyCostSecp256k1 = uint64(simtypes.RandIntBetween(r, 1, 1000))

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}, nil
}
