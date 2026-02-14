package cachemulti

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/types"
)

func TestStoreGetKVStore(t *testing.T) {
	require := require.New(t)

	s := Store{stores: map[types.StoreKey]types.CacheWrap{}}
	key := types.NewKVStoreKey("abc")
	errMsg := fmt.Sprintf("kv store with key %v has not been registered in stores", key)

	require.PanicsWithValue(errMsg,
		func() { s.GetStore(key) })

	require.PanicsWithValue(errMsg,
		func() { s.GetKVStore(key) })
}

func TestCacheMultiStoreTraceContextClone(t *testing.T) {
	tc := types.TraceContext(map[string]interface{}{"blockHeight": 64})
	buf := &bytes.Buffer{}

	cms := NewFromKVStore(map[types.StoreKey]types.CacheWrapper{}, buf, tc)
	require.NotNil(t, cms)

	childCms := cms.CacheMultiStore().(Store)

	cms.SetTracingContext(types.TraceContext{"newKey": "newValue"})

	_, hasNewKey := cms.traceContext["newKey"]
	require.True(t, hasNewKey, "parent should have newKey after SetTracingContext")

	_, childHasNewKey := childCms.traceContext["newKey"]
	require.False(t, childHasNewKey, "child should not have newKey - traceContext should be cloned")
}

// TestCacheMultiStoreTraceConcurrency tests that concurrent access to child stores'
// traceContext does not cause "concurrent map iteration and map write" panic.
func TestCacheMultiStoreTraceConcurrency(t *testing.T) {
	require.NotPanics(t, func() {
		tc := types.TraceContext(map[string]interface{}{"blockHeight": 64})
		buf := &bytes.Buffer{}

		cms := NewFromKVStore(map[types.StoreKey]types.CacheWrapper{}, buf, tc)

		const numGoroutines = 10
		const numIterations = 100

		childStores := make([]types.CacheMultiStore, numGoroutines)
		for i := range numGoroutines {
			childStores[i] = cms.CacheMultiStore()
		}

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := range numGoroutines {
			go func(idx int) {
				defer wg.Done()
				child := childStores[idx]
				for j := range numIterations {
					child.SetTracingContext(types.TraceContext{
						fmt.Sprintf("key-%d", idx): fmt.Sprintf("value-%d", j),
					})
				}
			}(i)
		}

		wg.Wait()
	})
}
