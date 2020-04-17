package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
)

// BeginBlocker updates an existing localhost client with the latest block height.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	localhostClient, found := k.GetClientState(ctx, exported.ClientTypeLocalHost)
	if !found {
		return
	}

	// update the localhost client with the latest block height
	_, err := k.UpdateClient(ctx, localhostClient.GetID(), nil)
	if err != nil {
		panic(err)
	}
}
