package cache_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/cache"
	"cosmossdk.io/store/cachekv"
	iavlstore "cosmossdk.io/store/iavl"
	"cosmossdk.io/store/types"
	"cosmossdk.io/store/wrapper"
)

// Note: cachekv is still imported for TestCacheWrap which tests CacheWrap() returns *cachekv.Store

func TestGetOrSetStoreCache(t *testing.T) {
	db := wrapper.NewDBWrapper(dbm.NewMemDB())
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree := iavl.NewMutableTree(db, 100, false, log.NewNopLogger())
	store := iavlstore.UnsafeNewStore(tree)
	store2 := mngr.GetStoreCache(sKey, store)

	require.NotNil(t, store2)
	require.Equal(t, store2, mngr.GetStoreCache(sKey, store))
}

func TestUnwrap(t *testing.T) {
	db := wrapper.NewDBWrapper(dbm.NewMemDB())
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree := iavl.NewMutableTree(db, 100, false, log.NewNopLogger())
	store := iavlstore.UnsafeNewStore(tree)
	_ = mngr.GetStoreCache(sKey, store)

	require.Equal(t, store, mngr.Unwrap(sKey))
	require.Nil(t, mngr.Unwrap(types.NewKVStoreKey("test2")))
}

func TestStoreCache(t *testing.T) {
	db := wrapper.NewDBWrapper(dbm.NewMemDB())
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree := iavl.NewMutableTree(db, 100, false, log.NewNopLogger())
	store := iavlstore.UnsafeNewStore(tree)
	kvStore := mngr.GetStoreCache(sKey, store)

	for i := uint(0); i < cache.DefaultCommitKVStoreCacheSize*2; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d", i))

		kvStore.Set(key, value)

		res := kvStore.Get(key)
		require.Equal(t, res, value)
		require.Equal(t, res, store.Get(key))

		kvStore.Delete(key)

		require.Nil(t, kvStore.Get(key))
		require.Nil(t, store.Get(key))
	}
}

func TestReset(t *testing.T) {
	db := wrapper.NewDBWrapper(dbm.NewMemDB())
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree := iavl.NewMutableTree(db, 100, false, log.NewNopLogger())
	store := iavlstore.UnsafeNewStore(tree)
	store2 := mngr.GetStoreCache(sKey, store)

	require.NotNil(t, store2)
	require.Equal(t, store2, mngr.GetStoreCache(sKey, store))

	// reset and check if the cache is gone
	mngr.Reset()
	require.Nil(t, mngr.Unwrap(sKey))

	// check if the cache is recreated
	require.Equal(t, store2, mngr.GetStoreCache(sKey, store))
}

func TestCacheWrap(t *testing.T) {
	db := wrapper.NewDBWrapper(dbm.NewMemDB())
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree := iavl.NewMutableTree(db, 100, false, log.NewNopLogger())
	store := iavlstore.UnsafeNewStore(tree)

	cacheWrapper := mngr.GetStoreCache(sKey, store).CacheWrap()
	require.IsType(t, &cachekv.Store{}, cacheWrapper)
}

// TestInterBlockCacheRaceConcurrent reproduces a race condition in CommitKVStoreCache
// where a concurrent gRPC query can re-populate the cache with a stale value after
// Delete() removes it from cache but before the underlying store deletion completes.
//
// This is the root cause of app hash divergence on sentry nodes in Cosmos SDK chains.
// Sentry nodes handle heavy gRPC query traffic during block processing, making them
// susceptible to this race condition.
//
// Store hierarchy:
//
//	CacheKVStore (per-block) -> CommitKVStoreCache (inter-block) -> IAVL
//
// The bug is in CommitKVStoreCache.Delete() which performs two non-atomic operations:
//
//	func (ckv *CommitKVStoreCache) Delete(key []byte) {
//	    ckv.cache.Remove(key)           // Step 1: Remove from LRU cache
//	    ckv.CommitKVStore.Delete(key)   // Step 2: Delete from underlying store
//	}
//
// Race condition - a concurrent Get() can slip between steps 1 and 2:
//
//	1. Delete: cache.Remove(key)           <- Cache now empty for this key
//	2. Get:    cache.Get(key) -> miss      <- Concurrent query sees cache miss
//	3. Get:    CommitKVStore.Get(key)      <- Returns OLD value (delete not yet applied!)
//	4. Delete: CommitKVStore.Delete(key)   <- Now deleted from underlying store
//	5. Get:    cache.Add(key, stale_value) <- BUG: Stale value added AFTER delete!
//
// Result: Cache has stale value, IAVL has nil. Next block reads stale value from cache,
// computes wrong state transitions (e.g., double-distributing fees), and diverges from
// validators who didn't hit the race.
func TestInterBlockCacheRaceConcurrent(t *testing.T) {
	db := wrapper.NewDBWrapper(dbm.NewMemDB())
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("bank")
	tree := iavl.NewMutableTree(db, 100, false, log.NewNopLogger())
	store := iavlstore.UnsafeNewStore(tree)

	interBlockCache := mngr.GetStoreCache(sKey, store)

	feeCollectorKey := []byte("fee_collector_balance")
	initialBalance := []byte("100")

	raceDetected := atomic.Bool{}

	// Run many iterations to catch the race
	for i := 0; i < 1000; i++ {
		// Reset state: set value and commit
		interBlockCache.Set(feeCollectorKey, initialBalance)
		_, _, err := tree.SaveVersion()
		require.NoError(t, err)

		require.Equal(t, initialBalance, interBlockCache.Get(feeCollectorKey))

		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine 1: Delete (simulates block processing)
		go func() {
			defer wg.Done()
			interBlockCache.Delete(feeCollectorKey)
		}()

		// Goroutine 2: Get (simulates concurrent query)
		go func() {
			defer wg.Done()
			_ = interBlockCache.Get(feeCollectorKey)
		}()

		wg.Wait()

		// Commit the deletion
		_, _, err = tree.SaveVersion()
		require.NoError(t, err)

		// IAVL should have nil
		require.Nil(t, store.Get(feeCollectorKey), "IAVL should have nil")

		// Check if cache has stale value (the bug)
		cachedValue := interBlockCache.Get(feeCollectorKey)
		if cachedValue != nil {
			raceDetected.Store(true)
			t.Logf("RACE DETECTED at iteration %d: cache has stale value '%s'", i, string(cachedValue))
			t.Logf("IAVL correctly has nil, but cache returned stale data")
			t.Logf("This causes app hash divergence on sentry nodes!")
			break
		}
	}

	if raceDetected.Load() {
		t.Log("BUG CONFIRMED: Inter-block cache race condition exists")
		t.Log("Fix: Make Delete() atomic or track pending deletions")
		t.FailNow()
	} else {
		t.Log("Race not triggered in 1000 iterations (may need more runs or -race flag)")
	}
}
