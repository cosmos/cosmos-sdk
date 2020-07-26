package connection

import (
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/x/ibc/03-connection/keeper"
	"github.com/KiraCore/cosmos-sdk/x/ibc/03-connection/types"
)

// InitGenesis initializes the ibc connection submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, connection := range gs.Connections {
		conn := types.NewConnectionEnd(connection.State, connection.ClientID, connection.Counterparty, connection.Versions)
		k.SetConnection(ctx, connection.ID, conn)
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
