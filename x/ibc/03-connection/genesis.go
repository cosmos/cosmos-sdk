package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
)

// InitGenesis initializes the ibc connection submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, gs GenesisState) {
	for _, connection := range gs.Connections {
		k.SetConnection(ctx, connection.ID, connection)
	}
}

// ExportGenesis returns the ibc connection submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {

	var openConnections []Conne
	connections := k.GetAllConnections(ctx)
	// filter OPEN connections
	for _, connection := range connections {
		if conn.State != exported.OPEN {
			continue
		}

		openConnections = append(openConnections, connection)
	}

	return GenesisState{
		Connections: openConnections,
	}
}
