package transfer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

// InitGenesis binds to portid from genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, state types.GenesisState) {
	k.SetPort(ctx, state.PortID)

	// Only try to bind to port if it is not already bound, since we may already own
	// port capability from capability InitGenesis
	if !k.IsBound(ctx, state.PortID) {
		// transfer module binds to the transfer port on InitChain
		// and claims the returned capability
		err := k.BindPort(ctx, state.PortID)
		if err != nil {
			panic(fmt.Sprintf("could not claim port capability: %v", err))
		}
	}

	k.SetParams(ctx, state.Params)

	// check if the module account exists
	moduleAcc := k.GetTransferAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
}

// ExportGenesis exports transfer module's portID into its geneis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		PortID: k.GetPort(ctx),
		Params: k.GetParams(ctx),
	}
}
