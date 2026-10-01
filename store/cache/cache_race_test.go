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

// park is how long the test holds the reader inside the race window before
// letting it continue. It only needs to be long enough for a concurrent Set to
// be attempted.
const park = 250 * time.Millisecond

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

	// The writer runs on its own goroutine: a correct implementation serializes it
	// behind the in-flight read-through, so blocking here would deadlock.
	set := make(chan struct{})
	go func() {
		ckv.Set(key, []byte("v2"))
		close(set)
	}()

	select {
	case <-set:
		// The writer was not serialized against the in-flight read-through, so it
		// published v2 into the cache first. The reader is about to write its
		// stale v1 back over the top of it.
	case <-time.After(park):
		// The writer is correctly blocked behind the reader's read-modify-write.
	}

	close(gated.released)

	require.Equal(t, []byte("v1"), <-read, "reader should observe the value committed before it started")
	<-set

	// Whichever order the two ran in, the last committed value must win. The
	// in-flight read-through may not overwrite the newer v2 with its stale v1.
	require.Equal(t, []byte("v2"), ckv.Get(key), "cache must not serve a value clobbered by a stale read-through")
	require.Equal(t, []byte("v2"), underlying.Get(key))
}
