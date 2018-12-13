package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// InitGenesis sets distribution information for genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetFeePool(ctx, data.FeePool)
	keeper.SetCommunityTax(ctx, data.CommunityTax)
	keeper.SetBaseProposerReward(ctx, data.BaseProposerReward)
	keeper.SetBonusProposerReward(ctx, data.BonusProposerReward)

	keeper.SetPreviousProposerConsAddr(ctx, data.PreviousProposer)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	feePool := keeper.GetFeePool(ctx)
	communityTax := keeper.GetCommunityTax(ctx)
	baseProposerRewards := keeper.GetBaseProposerReward(ctx)
	bonusProposerRewards := keeper.GetBonusProposerReward(ctx)
	pp := keeper.GetPreviousProposerConsAddr(ctx)
	return types.NewGenesisState(feePool, communityTax, baseProposerRewards, bonusProposerRewards, []types.DelegatorWithdrawInfo{}, pp)
}
