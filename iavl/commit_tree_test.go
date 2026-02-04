package iavl

import (
	"bytes"
	"context"
	"os"
	"runtime/debug"
	"slices"
	"testing"

	corestore "cosmossdk.io/core/store"
	sdklog "cosmossdk.io/log"
	iavl1 "github.com/cosmos/iavl"
	dbm "github.com/cosmos/iavl/db"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func TestIAVLXSims(t *testing.T) {
	rapid.Check(t, testIAVLXSims)
}

func FuzzIAVLX(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(testIAVLXSims))
}

func testIAVLXSims(t *rapid.T) {
	defer func() {
		if r := recover(); r != nil {
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
	simMachine := &SimMachine{
		treeV1:       treeV1,
		treeV2:       nil,
		dirV2:        tempDir,
		existingKeys: map[string][]byte{},
	}
	simMachine.openV2Tree(t)

	simMachine.Check(t)

	require.NoError(t, treeV1.Close(), "failed to close iavl tree")
	require.NoError(t, simMachine.treeV2.Close(), "failed to close iavlx tree")
}

type SimMachine struct {
	treeV1 *iavl1.MutableTree
	treeV2 *CommitTree
	dirV2  string
	// existingKeys keeps track of keys that have been set in the tree or deleted. Deleted keys are retained as nil values.
	existingKeys map[string][]byte
}

func (s *SimMachine) openV2Tree(t interface {
	require.TestingT
	sdklog.TestingT
}) {
	var err error
	s.treeV2, err = NewCommitTree(s.dirV2, Options{
		// intentionally choose some small sizes to force checkpoint and eviction behavior
		ChangesetRolloverSize: 4096,
		EvictDepth:            2,
		CheckpointInterval:    1,
	})
	require.NoError(t, err, "failed to create iavlx tree")
}

func (s *SimMachine) Check(t *rapid.T) {
	versions := rapid.IntRange(1, 100).Draw(t, "versions")
	for i := 0; i < versions; i++ {
		s.checkNewVersion(t)
	}
}

func (s *SimMachine) checkNewVersion(t *rapid.T) {
	// randomly generate some updates that we'll revert to test rollback capability
	testRollback := rapid.Bool().Draw(t, "testRollback")
	if testRollback {
		tempUpdates := s.genUpdates(t)
		committer := s.treeV2.StartCommit(context.Background(), slices.Values(tempUpdates), len(tempUpdates))
		// get the hash so we actually wait a bit before aborting
		_, err := committer.PrepareFinalize()
		require.NoError(t, err)
		require.NoError(t, committer.Rollback())
	}

	updates := s.genUpdates(t)
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
}

func (s *SimMachine) genUpdates(t *rapid.T) []KVUpdate {
	n := rapid.IntRange(1, 100).Draw(t, "n")
	updates := make([]KVUpdate, 0, n)
	for i := 0; i < n; i++ {
		key := s.selectKey(t)
		isDelete := rapid.Bool().Draw(t, "isDelete")
		if isDelete {
			updates = append(updates, KVUpdate{Key: key, Delete: true})
		} else {
			value := rapid.SliceOfN(rapid.Byte(), 0, 5000).Draw(t, "value")
			updates = append(updates, KVUpdate{Key: key, Value: value})
		}
	}
	return updates
}

func (s *SimMachine) selectKey(t *rapid.T) []byte {
	if len(s.existingKeys) > 0 && rapid.Bool().Draw(t, "existingKey") {
		return []byte(rapid.SampledFrom(maps.Keys(s.existingKeys)).Draw(t, "key"))
	} else {
		return rapid.SliceOfN(rapid.Byte(), 1, 500).Draw(t, "key")
	}
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
