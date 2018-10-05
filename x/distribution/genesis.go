package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// InitGenesis sets distribution information for genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetFeePool(ctx, data.FeePool)
	keeper.SetCommunityTax(ctx, data.CommunityTax)

	for _, vdi := range data.ValidatorDistInfos {
		keeper.SetValidatorDistInfo(ctx, vdi)
	}
	for _, ddi := range data.DelegatorDistInfos {
		keeper.SetDelegatorDistInfo(ctx, ddi)
	}
	for _, dw := range data.DelegatorWithdrawInfos {
		keeper.SetDelegatorWithdrawAddr(ctx, dw.DelegatorAddr, dw.WithdrawAddr)
	}
}

// WriteGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, and validator/delegator distribution info's
func WriteGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	feePool := keeper.GetFeePool(ctx)
	communityTax := keeper.GetCommunityTax(ctx)
	vdis := keeper.GetAllVDIs(ctx)
	ddis := keeper.GetAllDDIs(ctx)
	dwis := keeper.GetAllDWIs(ctx)
	return NewGenesisState(feePool, communityTax, vdis, ddis, dwis)
}
