package iavlx

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
	"testing"

	"github.com/cosmos/iavl"
	dbm "github.com/cosmos/iavl/db"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	corestore "cosmossdk.io/core/store"
	sdklog "cosmossdk.io/log"

	storetypes "cosmossdk.io/store/types"
)

func TestTree_MembershipProof(t *testing.T) {
	rapid.Check(t, testTreeMembershipProof)
}

func testTreeMembershipProof(t *rapid.T) {
	dir, err := os.MkdirTemp("", "iavlx-membership-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	commitTree, err := NewCommitTree(dir, Options{}, sdklog.NewNopLogger())
	require.NoError(t, err)
	defer commitTree.Close()

	// Generate a random number of key-value pairs
	numKeys := rapid.IntRange(1, 200).Draw(t, "numKeys")
	keys := make([][]byte, 0, numKeys)
	keySet := make(map[string]bool) // Track keys to avoid duplicates

	for i := range numKeys {
		// Generate random key and value with varying lengths (non-empty)
		keyLen := rapid.IntRange(1, 50).Draw(t, fmt.Sprintf("keyLen-%d", i))
		valueLen := rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("valueLen-%d", i))
		key := rapid.SliceOfN(rapid.Byte(), keyLen, keyLen).Draw(t, fmt.Sprintf("key-%d", i))
		value := rapid.SliceOfN(rapid.Byte(), valueLen, valueLen).Draw(t, fmt.Sprintf("value-%d", i))

		// Skip duplicates
		if keySet[string(key)] {
			continue
		}
		keySet[string(key)] = true

		commitTree.Set(key, value)
		keys = append(keys, key)
	}

	// Commit the tree
	commitTree.Commit()
	itree, err := commitTree.GetImmutable(1)
	require.NoError(t, err)

	// Property: For all keys that were inserted, membership proofs should be valid
	for _, key := range keys {
		proof, err := itree.GetMembershipProof(key)
		require.NoError(t, err, "failed to get membership proof for key %X", key)

		ok, err := itree.VerifyMembership(proof, key)
		require.NoError(t, err, "failed to verify membership for key %X", key)
		require.True(t, ok, "membership verification failed for key %X", key)
	}
}

func TestTree_NonMembershipProof(t *testing.T) {
	rapid.Check(t, testTreeNonMembershipProof)
}

func testTreeNonMembershipProof(t *rapid.T) {
	dir, err := os.MkdirTemp("", "iavlx-nonmembership-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	commitTree, err := NewCommitTree(dir, Options{}, sdklog.NewNopLogger())
	require.NoError(t, err)
	defer commitTree.Close()

	// Generate a random number of key-value pairs and track them
	numKeys := rapid.IntRange(1, 200).Draw(t, "numKeys")
	keySet := make(map[string]bool)

	for i := range numKeys {
		// Generate random key and value with varying lengths (non-empty values)
		keyLen := rapid.IntRange(1, 50).Draw(t, fmt.Sprintf("keyLen-%d", i))
		valueLen := rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("valueLen-%d", i))
		key := rapid.SliceOfN(rapid.Byte(), keyLen, keyLen).Draw(t, fmt.Sprintf("key-%d", i))
		value := rapid.SliceOfN(rapid.Byte(), valueLen, valueLen).Draw(t, fmt.Sprintf("value-%d", i))

		keySet[string(key)] = true
		commitTree.Set(key, value)
	}

	// Commit the tree
	commitTree.Commit()
	itree, err := commitTree.GetImmutable(1)
	require.NoError(t, err)

	// Property: For keys that are NOT in the tree, non-membership proofs should be valid
	numNonMembershipTests := rapid.IntRange(1, 100).Draw(t, "numNonMembershipTests")
	for i := range numNonMembershipTests {
		// Generate random keys that are not in the keySet
		var nonExistentKey []byte
		maxAttempts := 100
		for attempt := 0; attempt < maxAttempts; attempt++ {
			keyLen := rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("nonMemberKeyLen-%d-%d", i, attempt))
			candidate := rapid.SliceOfN(rapid.Byte(), keyLen, keyLen).Draw(t, fmt.Sprintf("nonMemberKey-%d-%d", i, attempt))

			// Check if this key is not in the tree
			if !keySet[string(candidate)] {
				nonExistentKey = candidate
				break
			}
		}

		// If we couldn't find a non-existent key after max attempts, skip this iteration
		if nonExistentKey == nil {
			continue
		}

		proof, err := itree.GetNonMembershipProof(nonExistentKey)
		require.NoError(t, err, "failed to get non-membership proof for key %X", nonExistentKey)
		require.NotNil(t, proof, "non-membership proof is nil for key %X", nonExistentKey)

		valid, err := itree.VerifyNonMembership(proof, nonExistentKey)
		require.NoError(t, err, "failed to verify non-membership for key %X", nonExistentKey)
		require.True(t, valid, "non-membership verification failed for key %X", nonExistentKey)
	}
}

// TODO: this test isn't for expected behavior.
// It should eventually be updated such that having a default ReaderUpdateInterval shouldn't error on old version queries.
func TestTree_ErrorsOnOldVersion(t *testing.T) {
	testCases := []struct {
		name     string
		getTree  func() *CommitTree
		expError error
	}{
		{
			name: "should error",
			getTree: func() *CommitTree {
				dir := t.TempDir()
				commitTree, err := NewCommitTree(dir, Options{}, sdklog.NewNopLogger())
				require.NoError(t, err)
				return commitTree
			},
			// TODO: this shouldn't error!
			expError: fmt.Errorf("no changeset found for version 2"),
		},
		{
			name: "should NOT error",
			getTree: func() *CommitTree {
				dir := t.TempDir()
				commitTree, err := NewCommitTree(dir, Options{ReaderUpdateInterval: 1}, sdklog.NewNopLogger())
				require.NoError(t, err)
				return commitTree
			},
			expError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commitTree := tc.getTree()
			for range 7 {
				tree := commitTree.CacheWrap().(storetypes.CacheKVStore)
				tree.Set([]byte{0}, []byte{1})
				tree.Write()
				commitTree.Commit()
			}
			_, err := commitTree.GetImmutable(2)
			require.Equal(t, tc.expError, err)
		})
	}

}

func TestTree_NonExistentChangeset(t *testing.T) {
	dir := t.TempDir()
	commitTree, err := NewCommitTree(dir, Options{ReaderUpdateInterval: 1}, sdklog.NewNopLogger())
	require.NoError(t, err)

	for range 7 {
		tree := commitTree.CacheWrap().(storetypes.CacheKVStore)
		tree.Set([]byte{0}, []byte{1})
		tree.Write()
		commitTree.Commit()
	}

	_, err = commitTree.GetImmutable(2)
	require.NoError(t, err)
}

func TestBasicTest(t *testing.T) {
	dir, err := os.MkdirTemp("", "iavlx")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	commitTree, err := NewCommitTree(dir, Options{}, sdklog.NewNopLogger())
	require.NoError(t, err)
	tree := commitTree.CacheWrap().(storetypes.CacheKVStore)
	tree.Set([]byte{0}, []byte{1})
	// renderTree(t, tree)

	val := tree.Get([]byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{1}, val)

	tree.Set([]byte{1}, []byte{2})
	// renderTree(t, tree)

	val = tree.Get([]byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{1}, val)
	val = tree.Get([]byte{1})
	require.NoError(t, err)
	require.Equal(t, []byte{2}, val)

	tree.Set([]byte{2}, []byte{3})
	// renderTree(t, tree)

	val = tree.Get([]byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{1}, val)
	val = tree.Get([]byte{1})
	require.NoError(t, err)
	require.Equal(t, []byte{2}, val)
	val = tree.Get([]byte{2})
	require.NoError(t, err)
	require.Equal(t, []byte{3}, val)

	val = tree.Get([]byte{3})
	require.NoError(t, err)
	require.Nil(t, val)

	tree.Delete([]byte{1})
	// renderTree(t, tree)

	val = tree.Get([]byte{1})
	require.NoError(t, err)
	require.Nil(t, val)

	tree.Write()
	commitId := commitTree.Commit()
	require.NoError(t, err)
	require.NotNil(t, commitId)
	t.Logf("committed with root commitId: %X", commitId)
	require.NoError(t, commitTree.Close())
}

func renderTree(t interface {
	require.TestingT
	Logf(format string, args ...any)
}, tree *ImmutableTree,
) {
	graph := &bytes.Buffer{}
	require.NoError(t, RenderNodeDotGraph(graph, tree.root))
	t.Logf("tree graph:\n%s", graph.String())
}

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
	treeV1 := iavl.NewMutableTree(dbV1, 500000, true, logger)

	tempDir, err := os.MkdirTemp("", "iavlx")
	require.NoError(t, err, "failed to create temp directory")
	defer os.RemoveAll(tempDir)
	simMachine := &SimMachine{
		treeV1:       treeV1,
		dirV2:        tempDir,
		existingKeys: map[string][]byte{},
	}
	simMachine.openV2Tree(t)

	// TODO switch from StateMachineActions to manually setting up the actions map, this is going to be too magical for other maintainers otherwise
	t.Repeat(map[string]func(*rapid.T){
		"":        simMachine.Check,
		"UpdateN": simMachine.UpdateN,
		"GetN":    simMachine.GetN,
		"Iterate": simMachine.Iterate,
		"Commit":  simMachine.Commit,
	})

	require.NoError(t, treeV1.Close(), "failed to close iavl tree")
	require.NoError(t, simMachine.treeV2.Close(), "failed to close iavlx tree")
}

type SimMachine struct {
	treeV1 *iavl.MutableTree
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
		WriteWAL:              true,
		CompactWAL:            true,
		DisableCompaction:     true,
		ZeroCopy:              false,
		EvictDepth:            0,
		CompactionOrphanRatio: 0,
		CompactionOrphanAge:   0,
		RetainVersions:        0,
		MinCompactionSeconds:  0,
		ChangesetMaxTarget:    1,
		CompactAfterVersions:  0,
		ReaderUpdateInterval:  1,
	}, sdklog.NewTestLogger(t))
	require.NoError(t, err, "failed to create iavlx tree")
}

func (s *SimMachine) Check(t *rapid.T) {
	// after every operation verify the iavlx tree
	// after every operation we check that both trees are identical
	s.compareIterators(t, nil, nil, true)
}

func (s *SimMachine) UpdateN(t *rapid.T) {
	n := rapid.IntRange(1, 5000).Draw(t, "n")
	for i := 0; i < n; i++ {
		del := rapid.Bool().Draw(t, "del")
		if del {
			s.delete(t)
		} else {
			s.set(t)
		}
	}
}

func (s *SimMachine) GetN(t *rapid.T) {
	n := rapid.IntRange(1, 5000).Draw(t, "n")
	for i := 0; i < n; i++ {
		s.get(t)
	}
}

func (s *SimMachine) set(t *rapid.T) {
	// choose either a new or an existing key
	key := s.selectKey(t)
	value := rapid.SliceOfN(rapid.Byte(), 0, 10).Draw(t, "value")
	// set in both trees
	updated, errV1 := s.treeV1.Set(key, value)
	require.NoError(t, errV1, "failed to set key in V1 tree")
	branch := s.treeV2.CacheWrap().(storetypes.CacheKVStore)
	branch.Set(key, value)
	branch.Write()
	// require.Equal(t, updated, updatedV2, "update status mismatch between V1 and V2 trees")
	if updated {
		require.NotNil(t, s.existingKeys[string(key)], "key shouldn't have been marked as updated")
	} else {
		existing, found := s.existingKeys[string(key)]
		if found {
			require.Nil(t, existing, value, "marked as not an update but existin key is nil")
		}
	}
	s.existingKeys[string(key)] = value // mark as existing
}

func (s *SimMachine) get(t *rapid.T) {
	key := s.selectKey(t)
	valueV1, errV1 := s.treeV1.Get(key)
	require.NoError(t, errV1, "failed to get key from V1 tree")
	valueV2 := s.treeV2.CacheWrap().(storetypes.CacheKVStore).Get(key)
	require.Equal(t, valueV1, valueV2, "value mismatch between V1 and V2 trees")
	expectedValue, found := s.existingKeys[string(key)]
	if found {
		require.Equal(t, expectedValue, valueV1, "expected value mismatch for key %s", key)
	} else {
		require.Nil(t, valueV1, "expected nil value for non-existing key %s", key)
	}
}

func (s *SimMachine) selectKey(t *rapid.T) []byte {
	if len(s.existingKeys) > 0 && rapid.Bool().Draw(t, "existingKey") {
		return []byte(rapid.SampledFrom(maps.Keys(s.existingKeys)).Draw(t, "key"))
	} else {
		// TODO consider testing longer keys
		return rapid.SliceOfN(rapid.Byte(), 1, 10).Draw(t, "key")
	}
}

func (s *SimMachine) delete(t *rapid.T) {
	key := s.selectKey(t)
	existingValue, found := s.existingKeys[string(key)]
	exists := found && existingValue != nil
	// delete in both trees
	_, removedV1, errV1 := s.treeV1.Remove(key)
	require.NoError(t, errV1, "failed to remove key from V1 tree")
	branch := s.treeV2.CacheWrap().(storetypes.CacheKVStore)
	branch.Delete(key)
	branch.Write()
	// require.Equal(t, removedV1, removedV2, "removed status mismatch between V1 and V2 trees")
	// TODO v1 & v2 have slightly different behaviors for the value returned on removal. We should re-enable this and check.
	//if valueV1 == nil || len(valueV1) == 0 {
	//	require.Empty(t, valueV2, "value should be empty for removed key in V2 tree")
	//} else {
	//	require.Equal(t, valueV1, valueV2, "value mismatch between V1 and V2 trees")
	//}
	require.Equal(t, exists, removedV1, "removed status should match existence of key")
	s.existingKeys[string(key)] = nil // mark as deleted
}

func (s *SimMachine) Iterate(t *rapid.T) {
	start := s.selectKey(t)
	end := s.selectKey(t)
	// make sure end is after start
	if string(end) <= string(start) {
		temp := start
		start = end
		end = temp
	}

	// TODO add cases where we nudge start or end up or down a little

	// ascending := rapid.Bool().Draw(t, "ascending")

	// s.compareIterators(t, start, end, ascending)
}

func (s *SimMachine) Commit(t *rapid.T) {
	hash1, _, err := s.treeV1.SaveVersion()
	require.NoError(t, err, "failed to save version in V1 tree")
	commitId2 := s.treeV2.Commit()
	require.NoError(t, err, "failed to save version in V2 tree")
	err = VerifyTree(s.treeV2)
	require.NoError(t, err, "failed to verify V2 tree")
	require.Equal(t, hash1, commitId2.Hash, "hash mismatch between V1 and V2 trees")
	closeReopen := rapid.Bool().Draw(t, "closeReopen")
	if closeReopen {
		require.NoError(t, s.treeV2.Close())
		s.openV2Tree(t)
	}
}

func (s *SimMachine) debugDump(t *rapid.T) {
	version := s.treeV1.Version()
	t.Logf("Dumping trees at version %d", version)
	graph1 := &bytes.Buffer{}
	iavl.WriteDOTGraph(graph1, s.treeV1.ImmutableTree, nil)
	t.Logf("V1 tree:\n%s", graph1.String())
	// renderTree(t, s.treeV2.Branch())
	iter2 := s.treeV2.CacheWrap().(storetypes.CacheKVStore).Iterator(nil, nil)
	s.debugDumpTree(t, iter2)
}

func (s *SimMachine) debugDumpTree(t *rapid.T, iter corestore.Iterator) {
	dumpStr := "Tree dump:"
	defer func() {
		require.NoError(t, iter.Close(), "failed to close iterator")
	}()
	for iter.Valid() {
		key := iter.Key()
		value := iter.Value()
		dumpStr += fmt.Sprintf("\n\tKey: %X, Value: %X", key, value)
		iter.Next()
	}
	t.Log(dumpStr)
}

// func (s *SimMachine) CheckoutVersion(t *rapid.T) {
//	if s.treeV1.Version() <= 1 {
//		// cannot checkout version 1 or lower
//		return
//	}
//	s.Commit(t) // make sure we've committed the current version before checking out a previous one
//	curVersion := s.treeV1.Version()
//	version := rapid.Int64Range(1, curVersion-1).Draw(t, "version")
//	itreeV1, err := s.treeV1.GetImmutable(version)
//	require.NoError(t, err, "failed to get immutable tree for V1 tree")
//	err = s.treeV2.LoadVersion(version)
//	require.NoError(t, err, "failed to load version in V2 tree")
//	defer require.NoError(t, s.treeV2.LoadVersion(curVersion), "failed to reload current version in V2 tree")
//
//	s.debugDumpTree(t)
//
//	s.compareIterators(t, nil, nil, true)
//	compareIteratorsAtVersion(t, itreeV1, s.treeV2, nil, nil, true)
//}

func (s *SimMachine) compareIterators(t *rapid.T, start, end []byte, ascending bool) {
	iter1, err1 := s.treeV1.Iterator(start, end, ascending)
	require.NoError(t, err1, "failed to create iterator for V1 tree")
	iter2 := s.treeV2.CacheWrap().(storetypes.CacheKVStore).Iterator(start, end)
	compareIteratorsAtVersion(t, iter1, iter2)
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
