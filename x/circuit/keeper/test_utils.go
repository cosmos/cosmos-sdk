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

type fixture struct {
	Ctx        sdk.Context
	Keeper     Keeper
	MockPerms  types.Permissions
	MockMsgURL string
}

func setupFixture(t *testing.T) *fixture {
	mockStoreKey := storetypes.NewKVStoreKey("circuit")
	keeperX := NewKeeper(mockStoreKey, addresses[0], addresscodec.NewBech32Codec("cosmos"))
	mockMsgURL := "mock_url"
	mockCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := mockCtx.Ctx.WithBlockHeader(cmproto.Header{})
	mockPerms := types.Permissions{
		Level:         3,
		LimitTypeUrls: []string{"test"},
	}

	return &fixture{
		Ctx:        ctx,
		Keeper:     keeperX,
		MockPerms:  mockPerms,
		MockMsgURL: mockMsgURL,
	}
}
