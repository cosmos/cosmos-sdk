package client

import (
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/types"
	localhosttypes "github.com/KiraCore/cosmos-sdk/x/ibc/09-localhost/types"
)

// InitGenesis initializes the ibc client submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, client := range gs.Clients {
		k.SetClientState(ctx, client)
		k.SetClientType(ctx, client.GetID(), client.ClientType())
	}
	for _, cs := range gs.ClientsConsensus {
		for _, consState := range cs.ConsensusStates {
			k.SetClientConsensusState(ctx, cs.ClientID, consState.GetHeight(), consState)
		}
	}

	if !gs.CreateLocalhost {
		return
	}

	// NOTE: return if the localhost client was already imported. The chain-id and
	// block height will be overwriten to the correct values during BeginBlock.
	if _, found := k.GetClientState(ctx, exported.ClientTypeLocalHost); found {
		return
	}

	// client id is always "localhost"
	clientState := localhosttypes.NewClientState(ctx.ChainID(), ctx.BlockHeight())

	_, err := k.CreateClient(ctx, clientState, nil)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the ibc client submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Clients:          k.GetAllClients(ctx),
		ClientsConsensus: k.GetAllConsensusStates(ctx),
		CreateLocalhost:  true,
	}
}
