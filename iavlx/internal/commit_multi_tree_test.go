package internal

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/log/v2"
	storeiavl "cosmossdk.io/store/iavl"
	"cosmossdk.io/store/rootmulti"
	pruningtypes "cosmossdk.io/store/pruning/types"
	store "cosmossdk.io/store/types"
)

func TestCommitMultiTree_Reload(t *testing.T) {
	dir := t.TempDir()
	var db *CommitMultiTree

	testStoreKey := store.NewKVStoreKey("test")
	loadDb := func() {
		var err error
		db, err = LoadCommitMultiTree(dir, Options{}, log.NewTestLogger(t))
		require.NoError(t, err)
		db.MountStoreWithDB(testStoreKey, store.StoreTypeIAVL, nil)
		require.NoError(t, db.LoadLatestVersion())
	}

	// open db & create some data
	loadDb()
	cb := db.CommitBranch()
	testStore := cb.GetKVStore(testStoreKey)
	testStore.Set([]byte("key1"), []byte("value1"))
	testStore.Set([]byte("key2"), []byte("value2"))
	committer, err := cb.StartCommit(context.Background(), cmtproto.Header{})
	require.NoError(t, err)
	commitId, err := committer.Finalize()
	require.NoError(t, err)

	// reload the DB
	require.NoError(t, db.Close())
	loadDb()

	// verify data is still there
	cb = db.CommitBranch()
	testStore = cb.GetKVStore(testStoreKey)
	val1 := testStore.Get([]byte("key1"))
	require.Equal(t, []byte("value1"), val1)
	val2 := testStore.Get([]byte("key2"))
	require.Equal(t, []byte("value2"), val2)
	committer, err = cb.StartCommit(context.Background(), cmtproto.Header{})
	require.NoError(t, err)
	commitId, err = committer.Finalize()
	require.NoError(t, err)

	// verify commit ID is the same
	require.Equal(t, commitId, db.LastCommitID())
}

func TestCommitMultiTreeSims(t *testing.T) {
	t.Cleanup(func() {
		if !t.Failed() {
			_ = os.RemoveAll("testdata/iavl-data")
		}
	})
	var iterCount atomic.Int32
	rapid.Check(t, func(t *rapid.T) {
		iter := int(iterCount.Add(1) - 1) // 0-based, matches rapid's "panic after N tests"
		testCommitMultiTreeSims(t, iter, Options{
			// intentionally choose some small sizes to force checkpoint and eviction behavior
			ChangesetRolloverSize:  4096,
			CompactionRolloverSize: 4096,
			BranchEvictDepth:       2,
			LeafEvictDepth:         2,
			CheckpointInterval:     2,
			// use only a small cache for testing
			RootCacheSize:   2,
			RootCacheExpiry: 5 * time.Millisecond,
			// we should never have any checkpoint errors during testing!
			DisableAutoRepair: true,
		}, pruningtypes.PruningOptions{
			KeepRecent: 5,
			Interval:   2,
		})
	})
}

func testCommitMultiTreeSims(t *rapid.T, iter int, opts Options, pruningOpts pruningtypes.PruningOptions) {
	dbV1 := dbm.NewMemDB()
	mtV1 := rootmulti.NewStore(dbV1, log.NewNopLogger())

	// NOTE: if we need to debug test failures, we can uncomment this code to save data after test failures:
	// dataDir := fmt.Sprintf("testdata/iavl-data/run-%d", iter)
	//require.NoError(t, os.MkdirAll(dataDir, 0o755), "failed to create data directory")
	//testPassed := false
	//defer func() {
	//	if testPassed {
	//		os.RemoveAll(dataDir)
	//	} else {
	//		t.Logf("keeping iavl data dir for debugging: %s", dataDir)
	//	}
	//}()
	dataDir, err := os.MkdirTemp("", "iavlx")
	require.NoError(t, err, "failed to create temp directory")
	defer os.RemoveAll(dataDir)

	sim := &SimCommitMultiTree{
		mtV1:        mtV1,
		dirV2:       dataDir,
		kv1:         store.NewKVStoreKey("kv1"),
		kv2:         store.NewKVStoreKey("kv2"),
		kv3:         store.NewKVStoreKey("kv3"),
		mem1:        store.NewMemoryStoreKey("mem1"),
		transient1:  store.NewTransientStoreKey("transient1"),
		opts:        opts,
		pruningOpts: pruningOpts,
	}
	sim.storeKeys = []store.StoreKey{sim.kv1, sim.kv2, sim.kv3, sim.mem1, sim.transient1}
	sim.kvStoreKeys = []store.StoreKey{sim.kv1, sim.kv2, sim.kv3}
	for range sim.storeKeys {
		sim.keyGens = append(sim.keyGens, newKeyGen(t))
	}

	// mount stores and load latest version for v1 store
	sim.mountStores(sim.mtV1)
	require.NoError(t, mtV1.LoadLatestVersion())

	// open v2 tree
	sim.openV2Tree(t)

	// defer func() {
	//	if r := recover(); r != nil {
	//		t.Fatalf("panic recovered: %v\nStack trace:\n%s", r, debug.Stack())
	//	}
	//}()

	t.Cleanup(func() {
		// NOTE: if we need to debug test failures, we can uncomment this code to save a debug HTML file with the tree state at the end of the test for inspection:
		//// generate debug HTML file for test inspection
		////if t.Failed() {
		// desc := sim.mtV2.Describe()
		//os.MkdirAll("testdata", 0o755)
		//f, err := os.Create(fmt.Sprintf("testdata/iavl-debug-run-%d.html", iter))
		//if err != nil {
		//	t.Logf("failed to create debug HTML file: %v", err)
		//	return
		//}
		//defer f.Close()
		//if err := RenderHTML(f, desc); err != nil {
		//	t.Logf("failed to render debug HTML: %v", err)
		//}
		////}

		require.NoError(t, sim.mtV2.Close(), "failed to close iavlx commit multi tree")
	})

	sim.Check(t)
	// NOTE: if we need to debug test failures, we can keep the data directory by not setting testPassed to true, which will prevent cleanup of the data directory and allow us to inspect the generated HTML file with the tree state at the end of the test
	// testPassed = true
}

type SimCommitMultiTree struct {
	mtV1  *rootmulti.Store
	mtV2  *CommitMultiTree
	dirV2 string

	kv1, kv2, kv3         *store.KVStoreKey
	mem1                  *store.MemoryStoreKey
	transient1            *store.TransientStoreKey
	storeKeys             []store.StoreKey
	kvStoreKeys           []store.StoreKey
	keyGens               []*keyGen
	opts                  Options
	pruningOpts           pruningtypes.PruningOptions
	lastCompactionVersion int64
}

func (sim *SimCommitMultiTree) openV2Tree(t *rapid.T) {
	var err error
	sim.mtV2, err = LoadCommitMultiTree(sim.dirV2, sim.opts, log.NewTestLogger(t))
	// we explicitly do not set pruning options here because we run pruning synchronously in the test when needed for determinism
	require.NoError(t, err, "failed to create iavlx commit multi tree")
	sim.mountStores(sim.mtV2)
	require.NoError(t, sim.mtV2.LoadLatestVersion())
}

func (sim *SimCommitMultiTree) mountStores(st store.CommitMultiStore) {
	st.MountStoreWithDB(sim.kv1, store.StoreTypeIAVL, nil)
	st.MountStoreWithDB(sim.kv2, store.StoreTypeIAVL, nil)
	st.MountStoreWithDB(sim.kv3, store.StoreTypeIAVL, nil)
	st.MountStoreWithDB(sim.mem1, store.StoreTypeMemory, nil)
	st.MountStoreWithDB(sim.transient1, store.StoreTypeTransient, nil)
}

func (sim *SimCommitMultiTree) Check(t *rapid.T) {
	versions := rapid.IntRange(1, 50).Draw(t, "versions")
	for i := 0; i < versions; i++ {
		sim.checkNewVersion(t)
	}
}

func (sim *SimCommitMultiTree) checkNewVersion(t *rapid.T) {
	// randomly generate some updates that we'll revert to test rollback capability
	testRollback := rapid.Bool().Draw(t, "testRollback")
	if testRollback {
		cb := sim.mtV2.CommitBranch()
		numUpdates := rapid.IntRange(0, 20).Draw(t, "numRollbackUpdates")
		for i := 0; i < numUpdates; i++ {
			j := rapid.IntRange(0, len(sim.storeKeys)-1).Draw(t, "storeKey")
			storeKey := sim.storeKeys[j]
			st := cb.GetKVStore(storeKey)
			// don't use the key gen here since we don't want to affect the main state!
			isDelete := rapid.Bool().Draw(t, "isDelete")
			key := rapid.SliceOfN(rapid.Byte(), 1, 100).Draw(t, "key")
			if isDelete {
				st.Delete(key)
			} else {
				value := rapid.SliceOfN(rapid.Byte(), 1, 1000).Draw(t, "value")
				st.Set(key, value)
			}
		}
		committer, err := cb.StartCommit(context.Background(), cmtproto.Header{})
		require.NoError(t, err)
		// wait a little bit of time before rolling back
		// to increase chance of overlapping with other async operations
		// inside the commit multi tree
		time.Sleep(5 * time.Millisecond)
		require.NoError(t, committer.Rollback())
	}

	cb1 := sim.mtV1.CommitBranch()
	cb2 := sim.mtV2.CommitBranch()

	sim.applyVersionUpdates(t, cb1, cb2)

	committer1, err := cb1.StartCommit(context.Background(), cmtproto.Header{})
	require.NoError(t, err)
	commitId1, err := committer1.Finalize()
	require.NoError(t, err)

	committer2, err := cb2.StartCommit(context.Background(), cmtproto.Header{})
	require.NoError(t, err)
	commitId2, err := committer2.Finalize()
	require.NoError(t, err)

	// verify commit IDs are the same
	require.Equal(t, commitId1.Version, commitId2.Version, "committed versions do not match")
	require.Equal(t, commitId1.Hash, commitId2.Hash, "commit hashes do not match")

	// prune manually for determinism in testing instead of relying on the async pruner
	if sim.pruningOpts.Interval != 0 && commitId1.Version-sim.lastCompactionVersion >= int64(sim.pruningOpts.Interval) {
		if commitId1.Version > int64(sim.pruningOpts.KeepRecent)+1 {
			retainVersion := commitId1.Version - int64(sim.pruningOpts.KeepRecent)
			t.Logf("compacting at version %d, retainVersion=%d", commitId1.Version, retainVersion)
			sim.lastCompactionVersion = commitId1.Version
			sim.mtV2.compactNow(context.Background(), uint64(retainVersion))
		}
	}

	// randomly close and reopen the V2 tree to test persistence
	closeReopen := rapid.Bool().Draw(t, "closeReopen")
	if closeReopen {
		require.NoError(t, sim.mtV2.Close())
		sim.openV2Tree(t)
	}

	// optionally check history by reopening old versions
	checkHistory := rapid.Bool().Draw(t, "checkHistory")
	if checkHistory && commitId1.Version > 1 {
		latestVersion := int(commitId1.Version)
		oldestVersion := 1
		keepWindow := latestVersion - int(sim.pruningOpts.KeepRecent)
		if keepWindow > oldestVersion {
			oldestVersion = keepWindow
		}
		historyVersion := rapid.IntRange(oldestVersion, latestVersion).Draw(t, "historyVersion")

		historyMs1, err := sim.mtV1.CacheMultiStoreWithVersion(int64(historyVersion))
		require.NoError(t, err, "failed to load historical version from V1 store")
		historyMs2, err := sim.mtV2.CacheMultiStoreWithVersion(int64(historyVersion))
		require.NoError(t, err, "failed to load historical version from V2 store")

		// compare contents of kv trees only
		for _, storeKey := range sim.kvStoreKeys {
			kvStore1 := historyMs1.GetKVStore(storeKey)
			kvStore2 := historyMs2.GetKVStore(storeKey)
			iterV1 := kvStore1.Iterator(nil, nil)
			iterV2 := kvStore2.Iterator(nil, nil)
			t.Logf("comparing store %s at version %d", storeKey.Name(), historyVersion)
			compareIteratorsAtVersion(t, iterV1, iterV2)
		}
	}
}

func (sim *SimCommitMultiTree) applyVersionUpdates(t *rapid.T, cacheMs1, cacheMs2 store.MultiStore) {
	n := rapid.IntRange(0, 200).Draw(t, "numUpdates")
	for i := 0; i < n; i++ {
		j := rapid.IntRange(0, len(sim.storeKeys)-1).Draw(t, "storeKey")
		storeKey := sim.storeKeys[j]
		gen := sim.keyGens[j]
		key, isDelete := gen.genOp(t)

		kvStore1 := cacheMs1.GetKVStore(storeKey)
		kvStore2 := cacheMs2.GetKVStore(storeKey)
		if isDelete {
			kvStore1.Delete(key)
			kvStore2.Delete(key)
		} else {
			value := rapid.SliceOfN(rapid.Byte(), 1, 1000).Draw(t, "value")
			kvStore1.Set(key, value)
			kvStore2.Set(key, value)
		}
	}

	for _, storeKey := range sim.kvStoreKeys {
		kvStore1 := cacheMs1.GetKVStore(storeKey)
		kvStore2 := cacheMs2.GetKVStore(storeKey)
		iterV1 := kvStore1.Iterator(nil, nil)
		iterV2 := kvStore2.Iterator(nil, nil)
		compareIteratorsAtVersion(t, iterV1, iterV2)
	}
}

type testKV struct {
	key   []byte
	value []byte
}

func TestCommitMultiTreeQueryKeyAndProof(t *testing.T) {
	mt, key := setupQueryableMultiTree(t)
	cid := commitQueryableTree(t, mt, key, []testKV{
		{key: []byte("alpha"), value: []byte("one")},
		{key: []byte("beta"), value: []byte("two")},
	})

	res, err := mt.Query(&store.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("alpha"),
		Height: cid.Version,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("one"), res.Value)
	require.Equal(t, cid.Version, res.Height)

	res, err = mt.Query(&store.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("missing"),
		Height: cid.Version,
	})
	require.NoError(t, err)
	require.Nil(t, res.Value)

	res, err = mt.Query(&store.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("alpha"),
		Height: cid.Version,
		Prove:  true,
	})
	require.NoError(t, err)
	require.Len(t, res.ProofOps.Ops, 2)
	require.Equal(t, store.ProofOpIAVLCommitment, res.ProofOps.Ops[0].Type)
	require.Equal(t, store.ProofOpSimpleMerkleCommitment, res.ProofOps.Ops[1].Type)

	prt := rootmulti.DefaultProofRuntime()
	require.NoError(t, prt.VerifyValue(res.ProofOps, cid.Hash, "/test/alpha", []byte("one")))

	res, err = mt.Query(&store.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("absent"),
		Height: cid.Version,
		Prove:  true,
	})
	require.NoError(t, err)
	require.Len(t, res.ProofOps.Ops, 2)
	require.NoError(t, prt.VerifyAbsence(res.ProofOps, cid.Hash, "/test/absent"))

	res, err = mt.Query(&store.RequestQuery{
		Path: "/test/key",
		Data: []byte("alpha"),
	})
	require.NoError(t, err)
	require.Equal(t, cid.Version, res.Height)
	require.Equal(t, []byte("one"), res.Value)

	res, err = mt.Query(&store.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("alpha"),
		Height: cid.Version + 10,
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Log)

	_, err = mt.Query(&store.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("alpha"),
		Height: cid.Version + 10,
		Prove:  true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "proof is unexpectedly empty")
}

func TestCommitMultiTreeQuerySubspaceCompat(t *testing.T) {
	mt, key := setupQueryableMultiTree(t)
	updates := []testKV{
		{key: []byte("a/1"), value: []byte("one")},
		{key: []byte("a/2"), value: []byte("two")},
		{key: []byte("b/1"), value: []byte("three")},
	}
	cid := commitQueryableTree(t, mt, key, updates)

	res, err := mt.Query(&store.RequestQuery{
		Path:   "/test/subspace",
		Data:   []byte("a/"),
		Height: cid.Version,
		Prove:  true, // prove should be ignored for /subspace
	})
	require.NoError(t, err)
	require.Nil(t, res.ProofOps)
	require.Equal(t, legacySubspaceResponse(t, updates, []byte("a/")), res.Value)

	res, err = mt.Query(&store.RequestQuery{
		Path:   "/test/subspace",
		Data:   nil,
		Height: cid.Version,
	})
	require.NoError(t, err)
}

func TestCommitMultiTreeQueryInvalidPaths(t *testing.T) {
	mt, key := setupQueryableMultiTree(t)
	_ = commitQueryableTree(t, mt, key, []testKV{{key: []byte("k"), value: []byte("v")}})

	_, err := mt.Query(&store.RequestQuery{
		Path: "/does-not-exist/key",
		Data: []byte("k"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such store")

	_, err = mt.Query(&store.RequestQuery{
		Path: "/test/nope",
		Data: []byte("k"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected query path")

	_, err = mt.Query(&store.RequestQuery{
		Path: "/test/key",
		Data: nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "query cannot be zero length")
}

func setupQueryableMultiTree(t *testing.T) (*CommitMultiTree, *store.KVStoreKey) {
	t.Helper()

	mt, err := LoadCommitMultiTree(t.TempDir(), Options{}, log.NewTestLogger(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = mt.Close() })

	key := store.NewKVStoreKey("test")
	mt.MountStoreWithDB(key, store.StoreTypeIAVL, nil)
	require.NoError(t, mt.LoadLatestVersion())

	return mt, key
}

func commitQueryableTree(t *testing.T, mt *CommitMultiTree, key store.StoreKey, updates []testKV) store.CommitID {
	t.Helper()

	cb := mt.CommitBranch()
	kvStore := cb.GetKVStore(key)
	for _, update := range updates {
		kvStore.Set(update.key, update.value)
	}

	committer, err := cb.StartCommit(context.Background(), cmtproto.Header{Time: time.Now()})
	require.NoError(t, err)

	cid, err := committer.Finalize()
	require.NoError(t, err)

	return cid
}

func legacySubspaceResponse(t *testing.T, updates []testKV, prefix []byte) []byte {
	t.Helper()

	db := dbm.NewMemDB()
	iavlStore, err := storeiavl.LoadStore(
		db,
		log.NewNopLogger(),
		store.NewKVStoreKey("legacy"),
		store.CommitID{},
		storeiavl.DefaultIAVLCacheSize,
		false,
	)
	require.NoError(t, err)
	legacy := iavlStore.(*storeiavl.Store)

	for _, update := range updates {
		legacy.Set(update.key, update.value)
	}
	cid := legacy.Commit()

	res, err := legacy.Query(&store.RequestQuery{
		Path:   "/subspace",
		Data:   prefix,
		Height: cid.Version,
	})
	require.NoError(t, err)

	return res.Value
}
