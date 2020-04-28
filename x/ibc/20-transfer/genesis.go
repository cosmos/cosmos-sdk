package transfer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// InitGenesis binds to portid from genesis state
func InitGenesis(ctx sdk.Context, keeper Keeper, state types.GenesisState) {
	// transfer module binds to the transfer port on InitChain
	// and claims the returned capability
	err := keeper.BindPort(ctx, state.PortID)
	if err != nil {
		panic(fmt.Sprintf("could not claim port capability: %v", err))
	}
	// check if the module account exists
	moduleAcc := keeper.GetTransferAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.GetModuleAccountName()))
	}
}

// ExportGenesis exports transfer module's portID into its geneis state
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	portID := keeper.GetPort(ctx)

	return types.GenesisState{
		PortID: portID,
	}
}
