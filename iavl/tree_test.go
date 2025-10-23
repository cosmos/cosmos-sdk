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
)

func TestBasicTest(t *testing.T) {
	dir, err := os.MkdirTemp("", "iavlx")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	commitTree, err := NewCommitTree(dir, Options{}, sdklog.NewNopLogger())
	require.NoError(t, err)
	tree := commitTree.CacheWrap().(*Tree)
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
}, tree *Tree,
) {
	graph := &bytes.Buffer{}
	require.NoError(t, RenderDotGraph(graph, tree.root))
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
	treeV2, err := NewCommitTree(tempDir, Options{
		WriteWAL:              true,
		CompactWAL:            true,
		DisableCompaction:     true,
		ZeroCopy:              false,
		EvictDepth:            0,
		WalSyncBuffer:         0,
		CompactionOrphanRatio: 0,
		CompactionOrphanAge:   0,
		RetainVersions:        0,
		MinCompactionSeconds:  0,
		ChangesetMaxTarget:    1,
		CompactAfterVersions:  0,
		ReaderUpdateInterval:  1,
	}, sdklog.NewNopLogger())
	require.NoError(t, err, "failed to create iavlx tree")
	simMachine := &SimMachine{
		treeV1:       treeV1,
		treeV2:       treeV2,
		existingKeys: map[string][]byte{},
	}

	// TODO switch from StateMachineActions to manually setting up the actions map, this is going to be too magical for other maintainers otherwise
	t.Repeat(map[string]func(*rapid.T){
		"":        simMachine.Check,
		"UpdateN": simMachine.UpdateN,
		"GetN":    simMachine.GetN,
		"Iterate": simMachine.Iterate,
		"Commit":  simMachine.Commit,
	})

	require.NoError(t, treeV1.Close(), "failed to close iavl tree")
	require.NoError(t, treeV2.Close(), "failed to close iavlx tree")
}

type SimMachine struct {
	treeV1 *iavl.MutableTree
	treeV2 *CommitTree
	// existingKeys keeps track of keys that have been set in the tree or deleted. Deleted keys are retained as nil values.
	existingKeys map[string][]byte
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
	branch := s.treeV2.CacheWrap().(*Tree)
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
	valueV2 := s.treeV2.CacheWrap().(*Tree).Get(key)
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
		return rapid.SliceOfN(rapid.Byte(), 0, 10).Draw(t, "key")
	}
}

func (s *SimMachine) delete(t *rapid.T) {
	key := s.selectKey(t)
	existingValue, found := s.existingKeys[string(key)]
	exists := found && existingValue != nil
	// delete in both trees
	_, removedV1, errV1 := s.treeV1.Remove(key)
	require.NoError(t, errV1, "failed to remove key from V1 tree")
	branch := s.treeV2.CacheWrap().(*Tree)
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
	// s.debugDump(t)
	err = VerifyTree(s.treeV2)
	// if err != nil {
	//	branches := s.treeV2.rollingDiff.branchData
	//	n := branches.Count()
	//	buf := &bytes.Buffer{}
	//	for i := uint64(0); i < n; i++ {
	//		branch, err := branches.Branch(i)
	//		require.NoError(t, err, "failed to read branch")
	//		buf.WriteString(fmt.Sprintf("%d: %s\n", i+1, branch))
	//	}
	//	require.NoError(t, os.WriteFile("branches_dump.txt", buf.Bytes(), 0o644))
	//
	//	buf = &bytes.Buffer{}
	//	require.NoError(t, s.treeV2.wal.DebugDump(buf))
	//	require.NoError(t, os.WriteFile("wal_dump.txt", buf.Bytes(), 0o644))
	//}
	require.NoError(t, err, "failed to verify V2 tree")
	require.Equal(t, hash1, commitId2.Hash, "hash mismatch between V1 and V2 trees")
	// require.Equal(t, v1, v2, "version mismatch between V1 and V2 trees")
}

func (s *SimMachine) debugDump(t *rapid.T) {
	version := s.treeV1.Version()
	t.Logf("Dumping trees at version %d", version)
	graph1 := &bytes.Buffer{}
	iavl.WriteDOTGraph(graph1, s.treeV1.ImmutableTree, nil)
	t.Logf("V1 tree:\n%s", graph1.String())
	// renderTree(t, s.treeV2.Branch())
	iter2 := s.treeV2.CacheWrap().(*Tree).Iterator(nil, nil)
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
	iter2 := s.treeV2.CacheWrap().(*Tree).Iterator(start, end)
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
