package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// InitGenesis initializes the ibc client submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, client := range gs.Clients {
		cs, ok := client.ClientState.GetCachedValue().(exported.ClientState)
		if !ok {
			panic("invalid client state")
		}

		k.SetClientState(ctx, client.ClientId, cs)
		k.SetClientType(ctx, client.ClientId, cs.ClientType())
	}

	for _, cs := range gs.ClientsConsensus {
		for _, consState := range cs.ConsensusStates {
			consensusState, ok := consState.GetCachedValue().(exported.ConsensusState)
			if !ok {
				panic("invalid consensus state")
			}

			k.SetClientConsensusState(ctx, cs.ClientId, consensusState.GetHeight(), consensusState)
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
	clientState := localhosttypes.NewClientState(ctx.ChainID(), ctx.BlockHeight(), ctx.BlockTime())

	_, err := k.CreateClient(ctx, exported.ClientTypeLocalHost, clientState, nil)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the ibc client submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Clients:          k.GetAllGenesisClients(ctx),
		ClientsConsensus: k.GetAllConsensusStates(ctx),
		CreateLocalhost:  false,
	}
}
