package rootmulti

import (
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
)

// storeCounts mirrors the order of magnitude of a real app's mounted modules.
// CommitInfo.Hash() rebuilds a SHA256 merkle root over one leaf per store, so the
// recompute cost scales with this count while the memoized read stays flat.
// The SDK alone mounts ~20 modules (the realistic low end); top chains add up to
// ~20 custom modules, so ~50 is roughly the high end. 1 is the degenerate baseline.
var storeCounts = []int{1, 20, 35, 50}

// newCommittedStore mounts n IAVL stores, writes a key into each so every
// StoreInfo carries a non-empty hash, and commits once. The returned store has a
// populated lastCommit snapshot, matching the state queries hit at runtime.
func newCommittedStore(b *testing.B, n int) *Store {
	b.Helper()

	rs := NewStore(dbm.NewMemDB(), log.NewNopLogger())
	rs.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))

	keys := make([]*types.KVStoreKey, n)
	for i := range keys {
		keys[i] = types.NewKVStoreKey(fmt.Sprintf("store%d", i))
		rs.MountStoreWithDB(keys[i], types.StoreTypeIAVL, nil)
	}
	if err := rs.LoadLatestVersion(); err != nil {
		b.Fatalf("load latest version: %v", err)
	}

	for _, k := range keys {
		rs.GetStoreByName(k.Name()).(types.KVStore).Set([]byte("k"), []byte("v"))
	}
	rs.Commit()

	return rs
}

// BenchmarkStore_LastCommitID measures the hot query-path call after the memoize
// change: it should be a flat field read regardless of store count.
func BenchmarkStore_LastCommitID(b *testing.B) {
	for _, n := range storeCounts {
		rs := newCommittedStore(b, n)
		b.Run(fmt.Sprintf("stores=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = rs.LastCommitID()
			}
		})
	}
}

// BenchmarkStore_LastCommitID_Recompute reproduces the pre-change cost: calling
// CommitInfo.CommitID() on every query rebuilt the SHA256 merkle root. Compare its
// ns/op and allocs against BenchmarkStore_LastCommitID to see what the memoize saves.
func BenchmarkStore_LastCommitID_Recompute(b *testing.B) {
	for _, n := range storeCounts {
		rs := newCommittedStore(b, n)
		b.Run(fmt.Sprintf("stores=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = rs.lastCommitInfo().CommitID()
			}
		})
	}
}
