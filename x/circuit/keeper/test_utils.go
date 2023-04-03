package keeper

import (
	"testing"
)

type baseFixture struct {
	t   *testing.T
	err error
	// ctx context.Context //TODO: uncomment this after implementing

	// k        Keeper //TODO: uncomment this after implementing
	// addrs    []sdk.AccAddress //TODO: uncomment this after implementing
	// storeKey *storetypes.KVStoreKey //TODO: uncomment this after implementing
	// sdkCtx   sdk.Context //TODO: uncomment this after implementing
}

func initFixture(t *testing.T) *baseFixture {
	s := &baseFixture{t: t}

	return s
}
