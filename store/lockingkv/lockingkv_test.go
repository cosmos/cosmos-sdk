package lockingkv_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/lockingkv"
	"cosmossdk.io/store/transient"
	storetypes "cosmossdk.io/store/types"
)

var (
	a   = []byte("a")
	b   = []byte("b")
	key = []byte("key")
)

func TestLockingKV_LinearizeReadsAndWrites(t *testing.T) {
	parent := transient.NewStore()
	locking := lockingkv.NewStore(parent)

	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			locked := locking.CacheWrapWithLocks([][]byte{a})
			defer locked.(storetypes.LockingStore).Unlock()
			v := locked.(storetypes.KVStore).Get(key)
			if v == nil {
				locked.(storetypes.KVStore).Set(key, []byte{1})
			} else {
				v[0]++
				locked.(storetypes.KVStore).Set(key, v)
			}
			locked.Write()
		}()
	}

	wg.Wait()
	require.Equal(t, []byte{100}, locking.Get(key))
}

func TestLockingKV_LockOrderToPreventDeadlock(t *testing.T) {
	parent := transient.NewStore()
	locking := lockingkv.NewStore(parent)

	// Acquire keys in two different orders ensuring that we don't reach deadlock.
	wg := sync.WaitGroup{}
	wg.Add(200)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			locked := locking.CacheWrapWithLocks([][]byte{a, b})
			defer locked.(storetypes.LockingStore).Unlock()
			v := locked.(storetypes.KVStore).Get(key)
			if v == nil {
				locked.(storetypes.KVStore).Set(key, []byte{1})
			} else {
				v[0]++
				locked.(storetypes.KVStore).Set(key, v)
			}
			locked.Write()
		}()

		go func() {
			defer wg.Done()

			locked := locking.CacheWrapWithLocks([][]byte{b, a})
			defer locked.(storetypes.LockingStore).Unlock()
			v := locked.(storetypes.KVStore).Get(key)
			if v == nil {
				locked.(storetypes.KVStore).Set(key, []byte{1})
			} else {
				v[0]++
				locked.(storetypes.KVStore).Set(key, v)
			}
			locked.Write()
		}()
	}

	wg.Wait()
	require.Equal(t, []byte{200}, locking.Get(key))
}

func TestLockingKV_AllowForParallelUpdates(t *testing.T) {
	parent := transient.NewStore()
	locking := lockingkv.NewStore(parent)

	wg := sync.WaitGroup{}
	wg.Add(100)

	lockeds := make([]storetypes.LockingStore, 100)
	for i := byte(0); i < 100; i++ {
		k := []byte{i}
		// We specifically don't unlock the keys during processing so that we can show that we must process all
		// of these in parallel before the wait group is done.
		locked := locking.CacheWrapWithLocks([][]byte{k})
		lockeds[i] = locked.(storetypes.LockingStore)
		go func() {
			// The defer order is from last to first so we mark that we are done and then exit.
			defer wg.Done()

			locked.(storetypes.KVStore).Set(k, k)
			locked.Write()
		}()
	}

	wg.Wait()
	for _, locked := range lockeds {
		locked.Unlock()
	}
	for i := byte(0); i < 100; i++ {
		require.Equal(t, []byte{i}, locking.Get([]byte{i}))
	}
}

func TestLockingKV_SetGetHas(t *testing.T) {
	parent := transient.NewStore()
	parent.Set(a, b)
	locking := lockingkv.NewStore(parent)

	// Check that Get is transitive to the parent.
	require.Equal(t, b, locking.Get(a))
	require.Nil(t, locking.Get(b))

	// Check that Has is transitive to the parent.
	require.True(t, locking.Has(a))
	require.False(t, locking.Has(b))

	// Check that Set isn't transitive to the parent.
	locking.Set(key, a)
	require.False(t, parent.Has(key))

	// Check that we can read our writes.
	require.True(t, locking.Has(key))
	require.Equal(t, a, locking.Get(key))

	// Check that committing the writes to the parent.
	locking.Write()
	require.True(t, parent.Has(key))
	require.Equal(t, a, parent.Get(key))
}

func TestLockedKV_SetGetHas(t *testing.T) {
	parent := transient.NewStore()
	parent.Set(a, b)
	locking := lockingkv.NewStore(parent)
	locked := locking.CacheWrapWithLocks([][]byte{key}).(storetypes.CacheKVStore)

	// Check that Get is transitive to the parent.
	require.Equal(t, b, locked.Get(a))
	require.Nil(t, locked.Get(b))

	// Check that Has is transitive to the parent.
	require.True(t, locked.Has(a))
	require.False(t, locked.Has(b))

	// Check that Set isn't transitive to the parent.
	locked.Set(key, a)
	require.False(t, locking.Has(key))

	// Check that we can read our writes.
	require.True(t, locked.Has(key))
	require.Equal(t, a, locked.Get(key))

	// Check that committing the writes to the parent and not the parent's parent.
	locked.Write()
	require.True(t, locking.Has(key))
	require.Equal(t, a, locking.Get(key))
	require.False(t, parent.Has(key))
	require.Nil(t, parent.Get(key))

	// Unlock and get another instance of the store to see that the mutations in the locking store are visible.
	locked.(storetypes.LockingStore).Unlock()
	locked = locking.CacheWrapWithLocks([][]byte{key}).(storetypes.CacheKVStore)
	require.True(t, locked.Has(key))
	require.Equal(t, a, locked.Get(key))
	locked.(storetypes.LockingStore).Unlock()
}
