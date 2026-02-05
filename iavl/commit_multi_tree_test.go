package iavl

import (
	"context"
	"os"
	"testing"

	sdklog "cosmossdk.io/log"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	storemetrics "cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	store "cosmossdk.io/store/types"
)

func TestCommitMultiTree_Reload(t *testing.T) {
	dir := t.TempDir()
	var db *CommitMultiTree

	testStoreKey := store.NewKVStoreKey("test")
	loadDb := func() {
		var err error
		db, err = LoadCommitMultiTree(dir, Options{})
		require.NoError(t, err)
		db.MountStoreWithDB(testStoreKey, store.StoreTypeIAVL, nil)
		require.NoError(t, db.LoadLatestVersion())
	}

	// open db & create some data
	loadDb()
	cacheMs := db.CacheMultiStore()
	testStore := cacheMs.GetKVStore(testStoreKey)
	testStore.Set([]byte("key1"), []byte("value1"))
	testStore.Set([]byte("key2"), []byte("value2"))
	committer, err := db.StartCommit(context.Background(), cacheMs, cmtproto.Header{})
	require.NoError(t, err)
	commitId, err := committer.Finalize()
	require.NoError(t, err)

	// reload the DB
	require.NoError(t, db.Close())
	loadDb()

	// verify data is still there
	cacheMs = db.CacheMultiStore()
	testStore = cacheMs.GetKVStore(testStoreKey)
	val1 := testStore.Get([]byte("key1"))
	require.Equal(t, []byte("value1"), val1)
	val2 := testStore.Get([]byte("key2"))
	require.Equal(t, []byte("value2"), val2)
	committer, err = db.StartCommit(context.Background(), cacheMs, cmtproto.Header{})
	require.NoError(t, err)
	commitId, err = committer.Finalize()
	require.NoError(t, err)

	// verify commit ID is the same
	require.Equal(t, commitId, db.LastCommitID())
}

func TestCommitMultiTreeSims(t *testing.T) {
	rapid.Check(t, testCommitMultiTreeSims)
}

func testCommitMultiTreeSims(t *rapid.T) {
	logger := sdklog.NewNopLogger()
	dbV1 := dbm.NewMemDB()
	mtV1 := rootmulti.NewStore(dbV1, logger, storemetrics.NewNoOpMetrics())

	tempDir, err := os.MkdirTemp("", "iavlx-mt")
	require.NoError(t, err, "failed to create temp directory")
	defer os.RemoveAll(tempDir)

	sim := &SimCommitMultiTree{
		mtV1:       mtV1,
		dirV2:      tempDir,
		kv1:        store.NewKVStoreKey("kv1"),
		kv2:        store.NewKVStoreKey("kv2"),
		kv3:        store.NewKVStoreKey("kv3"),
		mem1:       store.NewMemoryStoreKey("mem1"),
		transient1: store.NewTransientStoreKey("transient1"),
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

	sim.Check(t)

	require.NoError(t, sim.mtV2.Close(), "failed to close iavlx commit multi tree")
}

type SimCommitMultiTree struct {
	mtV1  *rootmulti.Store
	mtV2  *CommitMultiTree
	dirV2 string

	kv1, kv2, kv3 *store.KVStoreKey
	mem1          *store.MemoryStoreKey
	transient1    *store.TransientStoreKey
	storeKeys     []store.StoreKey
	kvStoreKeys   []store.StoreKey
	keyGens       []*keyGen
}

func (sim *SimCommitMultiTree) openV2Tree(t *rapid.T) {
	var err error
	sim.mtV2, err = LoadCommitMultiTree(sim.dirV2, Options{
		// intentionally choose some small sizes to force checkpoint and eviction behavior
		ChangesetRolloverSize: 4096,
		EvictDepth:            2,
		CheckpointInterval:    2,
	})
	require.NoError(t, err, "failed to create iavlx commit multi tree")
	sim.mountStores(sim.mtV2)
	require.NoError(t, sim.mtV2.LoadLatestVersion())
}

func (sim *SimCommitMultiTree) mountStores(st store.CommitMultiStore2) {
	st.MountStoreWithDB(sim.kv1, store.StoreTypeIAVL, nil)
	st.MountStoreWithDB(sim.kv2, store.StoreTypeIAVL, nil)
	st.MountStoreWithDB(sim.kv3, store.StoreTypeIAVL, nil)
	st.MountStoreWithDB(sim.mem1, store.StoreTypeMemory, nil)
	st.MountStoreWithDB(sim.transient1, store.StoreTypeTransient, nil)
}

func (sim *SimCommitMultiTree) Check(t *rapid.T) {
	versions := rapid.IntRange(1, 100).Draw(t, "versions")
	for i := 0; i < versions; i++ {
		sim.checkNewVersion(t)
	}
}

func (sim *SimCommitMultiTree) checkNewVersion(t *rapid.T) {
	cacheMs1 := sim.mtV1.CacheMultiStore()
	cacheMs2 := sim.mtV2.CacheMultiStore()

	sim.applyVersionUpdates(t, cacheMs1, cacheMs2)

	// follow old workflow with store v1
	cacheMs1.Write()
	commitId1 := sim.mtV1.Commit()

	// follow new workflow with store v2
	committer2, err := sim.mtV2.StartCommit(context.Background(), cacheMs2, cmtproto.Header{})
	require.NoError(t, err)
	commitId2, err := committer2.Finalize()
	require.NoError(t, err)

	// verify commit IDs are the same
	require.Equal(t, commitId1.Version, commitId2.Version, "committed versions do not match")
	require.Equal(t, commitId1.Hash, commitId2.Hash, "commit hashes do not match")

	// randomly close and reopen the V2 tree to test persistence
	closeReopen := rapid.Bool().Draw(t, "closeReopen")
	if closeReopen {
		require.NoError(t, sim.mtV2.Close())
		sim.openV2Tree(t)
	}
}

func (sim *SimCommitMultiTree) applyVersionUpdates(t *rapid.T, cacheMs1, cacheMs2 store.MultiStore) {
	n := rapid.IntRange(1, 200).Draw(t, "numUpdates")
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

	// compare contents of kv trees only
	for _, storeKey := range sim.kvStoreKeys {
		kvStore1 := cacheMs1.GetKVStore(storeKey)
		kvStore2 := cacheMs2.GetKVStore(storeKey)
		iterV1 := kvStore1.Iterator(nil, nil)
		iterV2 := kvStore2.Iterator(nil, nil)
		compareIteratorsAtVersion(t, iterV1, iterV2)
	}
}
