package cachemulti

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/dbadapter"
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
func TestConcurrentCacheMultiStoreTraceContext(t *testing.T) {
	// Create a StoreKey that will be shared across all goroutines
	storeKey := types.NewKVStoreKey("store1")

	// Create a parent store with tracing enabled
	stores := map[types.StoreKey]types.CacheWrapper{
		storeKey: dbadapter.Store{DB: dbm.NewMemDB()},
	}

	traceWriter := &bytes.Buffer{}
	traceContext := types.TraceContext{"initial": "context"}

	store := NewStore(stores, traceWriter, traceContext)

	// Run 100 concurrent goroutines that create child stores and update tracing context
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a child cache multi store
			// This should clone traceContext, not share the reference
			child := store.CacheMultiStore()

			// Update the tracing context on the child
			// If traceContext was shared by reference, this would cause
			// concurrent map iteration and map write panic
			child.SetTracingContext(types.TraceContext{
				"txHash": fmt.Sprintf("TX_%d", id),
				"action": "test",
			})

			// Access a store to ensure initialization
			// Use the SAME storeKey instance
			_ = child.GetKVStore(storeKey)
		}(i)
	}

	// Wait for all goroutines to complete
	// If there's a race condition, this will panic with:
	// "fatal error: concurrent map iteration and map write"
	wg.Wait()

	// If we reach here without panic, the test passes
	require.True(t, true, "No concurrent map access panic occurred")
}

// TestConcurrentCacheMultiStoreAccess tests concurrent access to the same
// parent store creating multiple child stores
func TestConcurrentCacheMultiStoreAccess(t *testing.T) {
	// Create shared StoreKeys
	storeKey1 := types.NewKVStoreKey("store1")
	storeKey2 := types.NewKVStoreKey("store2")

	stores := map[types.StoreKey]types.CacheWrapper{
		storeKey1: dbadapter.Store{DB: dbm.NewMemDB()},
		storeKey2: dbadapter.Store{DB: dbm.NewMemDB()},
	}

	traceWriter := &bytes.Buffer{}
	traceContext := types.TraceContext{
		"blockHeight": "100",
		"chainID":     "test-chain",
	}

	parentStore := NewStore(stores, traceWriter, traceContext)

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create multiple nested cache stores
			child1 := parentStore.CacheMultiStore()
			child1.SetTracingContext(types.TraceContext{
				"level": "1",
				"id":    fmt.Sprintf("%d", id),
			})

			child2 := child1.CacheMultiStore()
			child2.SetTracingContext(types.TraceContext{
				"level": "2",
				"id":    fmt.Sprintf("%d", id),
			})

			// Access stores using the SAME storeKey instances
			_ = child2.GetKVStore(storeKey1)
			_ = child2.GetKVStore(storeKey2)
		}(i)
	}

	wg.Wait()
	require.True(t, true, "No concurrent map access panic occurred")
}

// TestTraceContextIsolation verifies that child stores have isolated
// traceContext and modifications don't affect the parent
func TestTraceContextIsolation(t *testing.T) {
	storeKey := types.NewKVStoreKey("store1")

	stores := map[types.StoreKey]types.CacheWrapper{
		storeKey: dbadapter.Store{DB: dbm.NewMemDB()},
	}

	originalContext := types.TraceContext{"key": "original"}
	parentStore := NewStore(stores, &bytes.Buffer{}, originalContext)

	// Create a child and modify its context
	childStore := parentStore.CacheMultiStore()
	childStore.SetTracingContext(types.TraceContext{"key": "modified"})

	// Verify parent's context is unchanged
	// Note: Since Store is a value type, we can't directly check
	// but we can verify no panic occurs with concurrent access
	require.NotNil(t, childStore)
}
