package connection

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// InitGenesis initializes the ibc connection submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, gs GenesisState) {
	for _, connection := range gs.Connections {
		k.SetConnection(ctx, connection.ID, connection)
	}
	for _, connPaths := range gs.ClientConnectionPaths {
		k.SetClientConnectionPaths(ctx, connPaths.ClientID, connPaths.Paths)
	}
}

// ExportGenesis returns the ibc connection submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	var (
		openConnections []ConnectionEnd
		clientConnPaths []types.ConnectionPaths
	)

	connections := k.GetAllConnections(ctx)
	// filter OPEN connections
	for _, connection := range connections {
		if connection.State != exported.OPEN {
			continue
		}

		paths, found := k.GetClientConnectionPaths(ctx, connection.ClientID)
		if !found {
			panic(fmt.Sprintf("connection %s is not in client %s's paths", connection.ID, connection.ClientID))
		}
		// TODO: filter OPEN connections from slice
		connPath := types.ConnectionPaths{
			ClientID: connection.ClientID,
			Paths:    paths,
		}
		openConnections = append(openConnections, connection)
		clientConnPaths = append(clientConnPaths, connPath)
	}

	return GenesisState{
		Connections:           openConnections,
		ClientConnectionPaths: clientConnPaths,
	}
}
