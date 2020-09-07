package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// BeginBlocker updates an existing localhost client with the latest block height.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	_, found := k.GetClientState(ctx, exported.ClientTypeLocalHost)
	if !found {
		return
	}

	// update the localhost client with the latest block height
	_, err := k.UpdateClient(ctx, exported.ClientTypeLocalHost, nil)
	if err != nil {
		panic(err)
	}
}
