package iavl

import (
	"context"
	"os"
	"testing"

	sdklog "cosmossdk.io/log"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
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
	simMachine := &SimCommitMultiTree{
		mtV1:         mtV1,
		mtV2:         nil,
		dirV2:        tempDir,
		existingKeys: map[store.StoreKey]map[string][]byte{},
		kv1:          store.NewKVStoreKey("kv1"),
		kv2:          store.NewKVStoreKey("kv2"),
		kv3:          store.NewKVStoreKey("kv3"),
		mem1:         store.NewMemoryStoreKey("mem1"),
		transient1:   store.NewTransientStoreKey("transient1"),
	}
	simMachine.storeKeys = []store.StoreKey{simMachine.kv1, simMachine.kv2, simMachine.kv3, simMachine.mem1, simMachine.transient1}
	for _, sk := range simMachine.storeKeys {
		simMachine.existingKeys[sk] = map[string][]byte{}
	}

	// mount stores and load latest version for v1 store
	simMachine.mountStores(t, simMachine.mtV1)
	require.NoError(t, mtV1.LoadLatestVersion())

	// open v2 tree
	simMachine.openV2Tree(t)

	simMachine.Check(t)

	require.NoError(t, simMachine.mtV2.Close(), "failed to close iavlx commit multi tree")
}

type SimCommitMultiTree struct {
	mtV1  *rootmulti.Store
	mtV2  *CommitMultiTree
	dirV2 string
	// existingKeys keeps track of keys that have been set in the tree or deleted. Deleted keys are retained as nil values.
	existingKeys  map[store.StoreKey]map[string][]byte
	kv1, kv2, kv3 *store.KVStoreKey
	mem1          *store.MemoryStoreKey
	transient1    *store.TransientStoreKey
	// TODO maybe add an object store key later
	storeKeys []store.StoreKey
}

func (sim *SimCommitMultiTree) openV2Tree(t *rapid.T) {
	var err error
	sim.mtV2, err = LoadCommitMultiTree(sim.dirV2, Options{
		// intentionally choose some small sizes to force checkpoint and eviction behavior
		ChangesetRolloverSize: 4096,
		EvictDepth:            2,
		CheckpointInterval:    1,
	})
	require.NoError(t, err, "failed to create iavlx commit multi tree")
	sim.mountStores(t, sim.mtV2)
	require.NoError(t, sim.mtV2.LoadLatestVersion())
}

func (sim *SimCommitMultiTree) mountStores(t *rapid.T, st store.CommitMultiStore2) {
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
	//// randomly generate some updates that we'll revert to test rollback capability
	//testRollback := rapid.Bool().Draw(t, "testRollback")
	//if testRollback {
	//	tempUpdates := sim.genUpdates(t)
	//	committer := sim..
	//	StartCommit(context.Background(), slices.Values(tempUpdates), len(tempUpdates))
	//	// wait a little bit of time before rolling back
	//	time.Sleep(5 * time.Millisecond)
	//	require.NoError(t, committer.Rollback())
	//}

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
		storeKey, key := sim.selectKey(t)
		isDelete := rapid.Bool().Draw(t, "isDelete")
		kvStore1 := cacheMs1.GetKVStore(storeKey)
		kvStore2 := cacheMs2.GetKVStore(storeKey)
		if isDelete {
			kvStore1.Delete(key)
			kvStore2.Delete(key)
			sim.existingKeys[storeKey][string(key)] = nil
		} else {
			value := rapid.SliceOfN(rapid.Byte(), 1, 1000).Draw(t, "value")
			kvStore1.Set(key, value)
			kvStore2.Set(key, value)
			sim.existingKeys[storeKey][string(key)] = value
		}
	}

	// compare contents of kv trees only!
	for _, storeKey := range []store.StoreKey{sim.kv1, sim.kv2, sim.kv3} {
		kvStore1 := cacheMs1.GetKVStore(storeKey)
		kvStore2 := cacheMs2.GetKVStore(storeKey)
		iterV1 := kvStore1.Iterator(nil, nil)
		iterV2 := kvStore2.Iterator(nil, nil)
		compareIteratorsAtVersion(t, iterV1, iterV2)
	}
}

func (sim *SimCommitMultiTree) selectKey(t *rapid.T) (store.StoreKey, []byte) {
	storeKey := rapid.SampledFrom(sim.storeKeys).Draw(t, "storeKey")
	existingKeys := sim.existingKeys[storeKey]
	if len(existingKeys) > 0 && rapid.Bool().Draw(t, "existingKey") {
		return storeKey, []byte(rapid.SampledFrom(maps.Keys(existingKeys)).Draw(t, "key"))
	} else {
		return storeKey, rapid.SliceOfN(rapid.Byte(), 1, 500).Draw(t, "key")
	}
}
