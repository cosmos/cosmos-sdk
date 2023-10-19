package simulation

import (
	"math/rand"

	pooltypes "cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	OpWeightMsgCommunityPoolSpend = "op_weight_msg_community_pool_spend"

	DefaultWeightMsgCommunityPoolSpend int = 50
)

func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgCommunityPoolSpend,
			DefaultWeightMsgCommunityPoolSpend,
			SimulateMsgCommunityPoolSpend,
		),
	}
}

func SimulateMsgCommunityPoolSpend(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	accs := simtypes.RandomAccounts(r, 5)
	acc, _ := simtypes.RandomAcc(r, accs)

	coins, err := sdk.ParseCoinsNormalized("100stake,2testtoken")
	if err != nil {
		panic(err)
	}

	return &pooltypes.MsgCommunityPoolSpend{
		Authority: authority.String(),
		Recipient: acc.Address.String(),
		Amount:    coins,
	}
}
