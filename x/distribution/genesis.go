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

	for _, vdi := range data.ValidatorDistInfos {
		keeper.SetValidatorDistInfo(ctx, vdi)
	}
	for _, ddi := range data.DelegationDistInfos {
		keeper.SetDelegationDistInfo(ctx, ddi)
	}
	for _, dw := range data.DelegatorWithdrawInfos {
		keeper.SetDelegatorWithdrawAddr(ctx, dw.DelegatorAddr, dw.WithdrawAddr)
	}
	keeper.SetPreviousProposerConsAddr(ctx, data.PreviousProposer)
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, and validator/delegator distribution info's
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	feePool := keeper.GetFeePool(ctx)
	communityTax := keeper.GetCommunityTax(ctx)
	baseProposerRewards := keeper.GetBaseProposerReward(ctx)
	bonusProposerRewards := keeper.GetBonusProposerReward(ctx)
	vdis := keeper.GetAllValidatorDistInfos(ctx)
	ddis := keeper.GetAllDelegationDistInfos(ctx)
	dwis := keeper.GetAllDelegatorWithdrawInfos(ctx)
	pp := keeper.GetPreviousProposerConsAddr(ctx)
	return NewGenesisState(feePool, communityTax, baseProposerRewards,
		bonusProposerRewards, vdis, ddis, dwis, pp)
}
