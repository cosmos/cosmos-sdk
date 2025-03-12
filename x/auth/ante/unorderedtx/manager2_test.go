package unorderedtx_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/ante/unorderedtx"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func TestRemove(t *testing.T) {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	storeService := runtime.NewKVStoreService(mockStoreKey)
	testCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	mgr := unorderedtx.NewManager(storeService)

	sender := "cosmos1"
	timeout := time.Unix(100, 0)
	err := mgr.Add(testCtx.Ctx, sender, uint64(timeout.Unix()))
	require.NoError(t, err)

	ok, err := mgr.Contains(testCtx.Ctx, sender, uint64(timeout.Unix()))
	require.NoError(t, err)
	require.True(t, ok)

	blockTime := time.Unix(101, 0)
	testCtx.Ctx = testCtx.Ctx.WithBlockTime(blockTime)

	err = mgr.RemoveExpired(testCtx.Ctx)
	require.NoError(t, err)

	ok, err = mgr.Contains(testCtx.Ctx, sender, uint64(timeout.Unix()))
	fmt.Println(err)
	fmt.Println(ok)
}
