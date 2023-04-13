package keeper

import (
	"testing"
)

type baseFixture struct {
	t   *testing.T
	err error

	// TODO: uncomment these after implementing.
	// ctx context.Context

	// k        Keeper
	// addrs    []sdk.AccAddress
	// storeKey *storetypes.KVStoreKey
	// sdkCtx   sdk.Context
}

func initFixture(t *testing.T) *baseFixture {
	s := &baseFixture{t: t}

	return s
}
