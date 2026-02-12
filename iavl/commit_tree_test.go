package iavl

import (
	"bytes"
	"context"
	"math/rand/v2"
	"os"
	"runtime/debug"
	"slices"
	"testing"
	"time"

	corestore "cosmossdk.io/core/store"
	sdklog "cosmossdk.io/log"
	iavl1 "github.com/cosmos/iavl"
	dbm "github.com/cosmos/iavl/db"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func TestCommitTreeSims(t *testing.T) {
	rapid.Check(t, testCommitTreeSims)
}

func FuzzCommitTreeSims(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(testCommitTreeSims))
}

func testCommitTreeSims(t *rapid.T) {
	defer func() {
		if r := recover(); r != nil {
			// "overrun" happens when fuzz input is too short for rapid to generate all needed values
			// This is expected with Go's native fuzzing - just ignore these cases
			if r == "overrun" {
				return
			}
			t.Fatalf("panic recovered: %v\nStack trace:\n%s", r, debug.Stack())
		}
	}()
	// logger := sdklog.NewTestLogger(t)
	logger := sdklog.NewNopLogger()
	dbV1 := dbm.NewMemDB()
	treeV1 := iavl1.NewMutableTree(dbV1, 500000, true, logger)

	tempDir, err := os.MkdirTemp("", "iavlx")
	require.NoError(t, err, "failed to create temp directory")
	defer os.RemoveAll(tempDir)
	simMachine := &SimCommitTree{
		treeV1: treeV1,
		treeV2: nil,
		dirV2:  tempDir,
		keyGen: newKeyGen(t),
	}
	simMachine.openV2Tree(t)

	simMachine.Check(t)

	require.NoError(t, treeV1.Close(), "failed to close iavl tree")
	require.NoError(t, simMachine.treeV2.Close(), "failed to close iavlx tree")
}

type SimCommitTree struct {
	treeV1 *iavl1.MutableTree
	treeV2 *CommitTree
	dirV2  string
	keyGen *keyGen
}

func (s *SimCommitTree) openV2Tree(t interface {
	require.TestingT
	sdklog.TestingT
}) {
	var err error
	s.treeV2, err = NewCommitTree(s.dirV2, Options{
		// intentionally choose some small sizes to force checkpoint and eviction behavior
		ChangesetRolloverSize: 4096,
		BranchEvictDepth:      2,
		LeafEvictDepth:        2,
		CheckpointInterval:    2,
		// disable caches to simplify testing
		RootCacheSize:   -1,
		RootCacheExpiry: -1,
	})
	require.NoError(t, err, "failed to create iavlx tree")
}

func (s *SimCommitTree) Check(t *rapid.T) {
	versions := rapid.IntRange(1, 100).Draw(t, "versions")
	for i := 0; i < versions; i++ {
		s.checkNewVersion(t)
	}
}

func (s *SimCommitTree) checkNewVersion(t *rapid.T) {
	// randomly generate some updates that we'll revert to test rollback capability
	testRollback := rapid.Bool().Draw(t, "testRollback")
	if testRollback {
		tempUpdates := s.genUpdates(t, true)
		committer := s.treeV2.StartCommit(context.Background(), slices.Values(tempUpdates), len(tempUpdates))
		// wait a little bit of time before rolling back
		time.Sleep(5 * time.Millisecond)
		require.NoError(t, committer.Rollback())
	}

	updates := s.genUpdates(t, false)
	// apply updates to v1 tree
	for _, update := range updates {
		if update.Delete {
			_, _, err := s.treeV1.Remove(update.Key)
			require.NoError(t, err, "failed to delete key in V1 tree")
		} else {
			_, err := s.treeV1.Set(update.Key, update.Value)
			require.NoError(t, err, "failed to set key in V1 tree")
		}
	}
	hashV1, versionV1, err := s.treeV1.SaveVersion()
	require.NoError(t, err, "failed to save version in V1 tree")

	// apply updates to v2 tree
	committer := s.treeV2.StartCommit(context.Background(), slices.Values(updates), len(updates))
	commitIdV2, err := committer.Finalize()
	require.NoError(t, err, "failed to finalize commit in V2 tree")

	// check v2 iavl invariants
	latestPtr := s.treeV2.treeStore.Latest()
	if latestPtr != nil {
		latest, pin, err := latestPtr.Resolve()
		defer pin.Unpin()
		require.NoError(t, err, "failed to resolve latest node pointer in V2 tree")
		require.NoError(t, internal.VerifyAVLInvariants(latest))
	}

	// compare versions and hashes
	require.Equal(t, versionV1, commitIdV2.Version, "version mismatch between V1 and V2 trees")
	if !bytes.Equal(hashV1, commitIdV2.Hash) {
		t.Fatalf("hash mismatch between V1 and V2 trees: V1=%X, V2=%X", hashV1, commitIdV2.Hash)
	}

	// compare iterators at this version
	iterV1, err := s.treeV1.Iterator(nil, nil, true)
	require.NoError(t, err, "failed to create iterator for V1 tree")
	iterV2 := s.treeV2.Latest().Iterator(nil, nil)
	compareIteratorsAtVersion(t, iterV1, iterV2)

	// randomly close and reopen the V2 tree to test persistence
	closeReopen := rapid.Bool().Draw(t, "closeReopen")
	if closeReopen {
		require.NoError(t, s.treeV2.Close())
		s.openV2Tree(t)
	}

	// randomly check historical state
	checkHistory := rapid.Bool().Draw(t, "checkHistory")
	if checkHistory && versionV1 > 1 {
		historyVersion := int64(rapid.IntRange(1, int(versionV1-1)).Draw(t, "historyVersion"))
		treeV1, err := s.treeV1.GetImmutable(historyVersion)
		require.NoError(t, err)
		treeV2, err := s.treeV2.GetVersion(historyVersion)
		require.NoError(t, err)
		histIterV1, err := treeV1.Iterator(nil, nil, true)
		require.NoError(t, err, "failed to create historical iterator for V1 tree")
		histIterV2 := treeV2.Iterator(nil, nil)
		compareIteratorsAtVersion(t, histIterV1, histIterV2)
	}
}

func (s *SimCommitTree) genUpdates(t *rapid.T, forRollback bool) []KVUpdate {
	n := rapid.IntRange(0, 100).Draw(t, "n")
	updates := make([]KVUpdate, 0, n)
	for i := 0; i < n; i++ {
		var key []byte
		var isDelete bool
		if forRollback {
			key = rapid.SliceOfN(rapid.Byte(), 1, 500).Draw(t, "rollbackKey")
			isDelete = rapid.Bool().Draw(t, "rollbackIsDelete")
		} else {
			key, isDelete = s.keyGen.genOp(t)
		}
		if isDelete {
			updates = append(updates, KVUpdate{Key: key, Delete: true})
		} else {
			value := rapid.SliceOfN(rapid.Byte(), 0, 5000).Draw(t, "value")
			updates = append(updates, KVUpdate{Key: key, Value: value})
		}
	}
	return updates
}

func compareIteratorsAtVersion(t *rapid.T, iterV1, iterV2 corestore.Iterator) {
	defer func() {
		require.NoError(t, iterV1.Close(), "failed to close iterator for V1 tree")
	}()
	defer func() {
		require.NoError(t, iterV2.Close(), "failed to close iterator for V2 tree")
	}()

	for {
		hasNextV1 := iterV1.Valid()
		hasNextV2 := iterV2.Valid()
		require.Equal(t, hasNextV1, hasNextV2, "iterator validity mismatch between V1 and V2 trees")
		if !hasNextV1 {
			break
		}
		keyV1 := iterV1.Key()
		valueV1 := iterV1.Value()
		keyV2 := iterV2.Key()
		valueV2 := iterV2.Value()
		require.Equal(t, keyV1, keyV2, "key mismatch between V1 and V2 trees")
		require.Equal(t, valueV1, valueV2, "value mismatch between V1 and V2 trees")
		iterV1.Next()
		iterV2.Next()
	}
}

// keyGen tracks key generation state per store using index-based deterministic key generation.
// Keys are generated from (index, seed) so no map iteration is needed, making tests fully reproducible.
type keyGen struct {
	insertIndex uint64 // next index for new key insertion
	deleteIndex uint64 // next index for sequential deletion (keys below this are deleted)
	seed        uint64
}

func newKeyGen(t *rapid.T) *keyGen {
	seed := rapid.Uint64().Draw(t, "keyGenSeed")
	return &keyGen{
		insertIndex: 0,
		deleteIndex: 0,
		seed:        seed,
	}
}

func (gen *keyGen) genOp(t *rapid.T) (key []byte, isDelete bool) {
	hasExisting := gen.insertIndex > gen.deleteIndex

	if hasExisting {
		// 0=insert, 1=update existing, 2=delete existing
		op := rapid.IntRange(0, 2).Draw(t, "op")
		switch op {
		case 0: // insert new key
			key = genKey(gen.insertIndex, gen.seed)
			gen.insertIndex++
		case 1: // update existing key
			idx := uint64(rapid.IntRange(int(gen.deleteIndex), int(gen.insertIndex-1)).Draw(t, "keyIdx"))
			key = genKey(idx, gen.seed)
		case 2: // delete (sequential from bottom)
			key = genKey(gen.deleteIndex, gen.seed)
			gen.deleteIndex++
			isDelete = true
		}
	} else {
		// no existing keys, must insert
		key = genKey(gen.insertIndex, gen.seed)
		gen.insertIndex++
	}

	return key, isDelete
}

// genKey deterministically generates a key from an index and seed.
func genKey(index, seed uint64) []byte {
	rng := rand.New(rand.NewPCG(index, seed))
	length := rng.IntN(500) + 1
	key := make([]byte, length)
	for i := range key {
		key[i] = byte(rng.IntN(256))
	}
	return key
}
