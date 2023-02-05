package keeper

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type baseFixture struct {
	t   *testing.T
	err error
	ctx context.Context

	// k        Keeper //TODO uncomment this after implementing
	addrs    []sdk.AccAddress
	storeKey *storetypes.KVStoreKey
	sdkCtx   sdk.Context
}

func initFixture(t *testing.T) *baseFixture {
	s := &baseFixture{t: t}

	return s
}
