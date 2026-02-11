package blockstm

import (
	"bytes"
	"testing"

	"github.com/test-go/testify/require"

	"cosmossdk.io/store/cachemulti"
	storetypes "cosmossdk.io/store/types"
)

func TestMsWrapper_CacheWrapWithTrace(t *testing.T) {
	t.Parallel()
	storeKey := storetypes.NewKVStoreKey("test")
	mmdb := NewMultiMemDB(map[storetypes.StoreKey]int{storeKey: 0})
	ms := msWrapper{mmdb}

	var buf bytes.Buffer
	tc := storetypes.TraceContext(map[string]interface{}{"blockHeight": 1})

	cacheWrappedWithTrace := ms.CacheWrapWithTrace(&buf, tc)
	require.NotNil(t, cacheWrappedWithTrace)
	require.IsType(t, cachemulti.Store{}, cacheWrappedWithTrace)
}

func TestMsWrapper_CacheMultiStoreWithVersion(t *testing.T) {
	t.Parallel()
	storeKey := storetypes.NewKVStoreKey("test")
	mmdb := NewMultiMemDB(map[storetypes.StoreKey]int{storeKey: 0})
	ms := msWrapper{mmdb}

	require.Panics(t, func() {
		_, _ = ms.CacheMultiStoreWithVersion(1)
	}, "should panic when trying to branch cached multi-store with a version")
}

func TestMsWrapper_LatestVersion(t *testing.T) {
	t.Parallel()
	storeKey := storetypes.NewKVStoreKey("test")
	mmdb := NewMultiMemDB(map[storetypes.StoreKey]int{storeKey: 0})
	ms := msWrapper{mmdb}

	require.Panics(t, func() {
		_ = ms.LatestVersion()
	}, "should panic when trying to get latest version from branch cached multi-store")
}

