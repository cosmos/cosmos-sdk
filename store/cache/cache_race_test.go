package cache_test

import (
	"sync"
	"testing"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
	"cosmossdk.io/log/v2"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/v2/cache"
	iavlstore "github.com/cosmos/cosmos-sdk/store/v2/iavl"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/store/v2/wrapper"
)

// park is how long the test waits, with the reader still gated, to observe
// whether a concurrent Set is blocked behind the in-flight read-through. It
// only needs to be long enough for the writer to actually reach the lock.
const park = 500 * time.Millisecond

// settle is the bounded wait used for the writer handshakes, so a broken
// implementation fails the test instead of hanging it.
const settle = 5 * time.Second

// gatedCommitKVStore wraps a CommitKVStore so a test can pin a reader inside
// the window between CommitKVStoreCache.Get's read of the underlying store and
// its write-back into the ARC cache. That window is exactly where a concurrent
// Set can be clobbered by a stale value.
//
// Ref: https://github.com/cosmos/cosmos-sdk/issues/23891
type gatedCommitKVStore struct {
	types.CommitKVStore

	mtx      sync.Mutex
	gate     bool
	entered  chan struct{}
	released chan struct{}
}

func newGatedStore(inner types.CommitKVStore) *gatedCommitKVStore {
	return &gatedCommitKVStore{
		CommitKVStore: inner,
		entered:       make(chan struct{}),
		released:      make(chan struct{}),
	}
}

// arm makes the next Get block after it has read the underlying store.
func (s *gatedCommitKVStore) arm() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.gate = true
}

func (s *gatedCommitKVStore) Get(key []byte) []byte {
	value := s.CommitKVStore.Get(key)

	s.mtx.Lock()
	gated := s.gate
	s.gate = false
	s.mtx.Unlock()

	if gated {
		close(s.entered)
		<-s.released
	}

	return value
}

func TestCommitKVStoreCacheSetDuringReadThrough(t *testing.T) {
	db := wrapper.NewDBWrapper(dbm.NewMemDB())
	tree := iavl.NewMutableTree(db, 100, false, log.NewNopLogger())

	underlying := iavlstore.UnsafeNewStore(tree)
	gated := newGatedStore(underlying)
	ckv := cache.NewCommitKVStoreCache(gated, cache.DefaultCommitKVStoreCacheSize)

	key := []byte("key")

	// Seed the underlying store so the reader observes a stale value, then let a
	// concurrent commit publish a newer one for the same key.
	underlying.Set(key, []byte("v1"))

	read := make(chan []byte, 1)
	gated.arm()

	go func() {
		read <- ckv.Get(key)
	}()

	// Wait until the reader has read the underlying store but has not yet written
	// its result into the cache.
	<-gated.entered

	// The writer runs on its own goroutine. It signals setStarted before calling
	// Set so the check below cannot pass vacuously: without the handshake, a
	// writer that never ran would look identical to a correctly blocked one.
	setStarted := make(chan struct{})
	set := make(chan struct{})
	go func() {
		close(setStarted)
		ckv.Set(key, []byte("v2"))
		close(set)
	}()
	<-setStarted

	// With the reader still gated inside its read-modify-write, a correct
	// implementation must not let the writer finish. If the write lock is not
	// held across the read-through, Set completes here and the test fails.
	select {
	case <-set:
		t.Fatal("Set completed while a read-through was in flight: cache is not serialized against concurrent writers")
	case <-time.After(park):
	}

	// Release the reader and require that both goroutines actually finish, so a
	// broken lock shows up as a failure rather than a hang.
	close(gated.released)

	select {
	case v := <-read:
		require.Equal(t, []byte("v1"), v, "reader should observe the value committed before it started")
	case <-time.After(settle):
		t.Fatal("reader did not return after the gate was released")
	}

	select {
	case <-set:
	case <-time.After(settle):
		t.Fatal("writer did not return after the read-through completed")
	}

	// The newer committed value must win. The in-flight read-through may not
	// overwrite v2 with the stale v1 it read before the writer published.
	require.Equal(t, []byte("v2"), ckv.Get(key), "cache must not serve a value clobbered by a stale read-through")
	require.Equal(t, []byte("v2"), underlying.Get(key))
}
