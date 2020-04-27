package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the ibc client submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, gs GenesisState) {
	for _, client := range gs.Clients {
		k.SetClientState(ctx, client)
		k.SetClientType(ctx, client.GetID(), client.ClientType())
	}
	for _, cs := range gs.ClientsConsensus {
		for _, consState := range cs.ConsensusStates {
			k.SetClientConsensusState(ctx, cs.ClientID, consState.GetHeight(), consState)
		}
	}
}

// ExportGenesis returns the ibc client submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{
		Clients:          k.GetAllClients(ctx),
		ClientsConsensus: k.GetAllConsensusStates(ctx),
	}
}
