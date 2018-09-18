package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// InitGenesis sets the pool and parameters for the provided keeper and
// initializes the IntraTxCounter. For each validator in data, it sets that
// validator in the keeper along with manually setting the indexes. In
// addition, it also sets any delegations found in data. Finally, it updates
// the bonded validators.
// Returns final validator set after applying all declaration and delegations
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetFeePool(ctx, data.FeePool)

	for _, vdi := range data.ValidatorDistInfos {
		keeper.SetValidatorDistInfo(ctx, vdi)
	}
	for _, ddi := range data.DelegatorDistInfos {
		keeper.SetDelegatorDistInfo(ctx, ddi)
	}
	for _, dw := range data.DelegatorWithdrawAddrs {
		keeper.SetDelegatorWithdrawAddr(ctx, dw.DelegatorAddr, dw.WithdrawAddr)
	}
}

// WriteGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, and validator/delegator distribution info's
func WriteGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	feePool := keeper.GetFeePool(ctx)
	vdis := keeper.GetAllVDIs(ctx)
	ddis := keeper.GetAllDDIs(ctx)
	dws := keeper.GetAllDWs(ctx)

	return GenesisState{
		FeePool:                feePool,
		ValidatorDistInfos:     vdis,
		DelegatorDistInfos:     ddis,
		DelegatorWithdrawAddrs: dws,
	}
}
