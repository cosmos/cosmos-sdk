package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// InitGenesis initializes the ibc connection submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, connection := range gs.Connections {
		k.SetConnection(ctx, connection.ID, connection)
	}
	for _, connPaths := range gs.ClientConnectionPaths {
		k.SetClientConnectionPaths(ctx, connPaths.ClientID, connPaths.Paths)
	}
}

// ExportGenesis returns the ibc connection submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Connections:           k.GetAllConnections(ctx),
		ClientConnectionPaths: k.GetAllClientConnectionPaths(ctx),
	}
}
