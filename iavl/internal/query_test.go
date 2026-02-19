package internal

import (
	"context"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	sdklog "cosmossdk.io/log/v2"
	storeiavl "cosmossdk.io/store/iavl"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
)

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

	res, err := mt.Query(&storetypes.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("alpha"),
		Height: cid.Version,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("one"), res.Value)
	require.Equal(t, cid.Version, res.Height)

	res, err = mt.Query(&storetypes.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("missing"),
		Height: cid.Version,
	})
	require.NoError(t, err)
	require.Nil(t, res.Value)

	res, err = mt.Query(&storetypes.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("alpha"),
		Height: cid.Version,
		Prove:  true,
	})
	require.NoError(t, err)
	require.Len(t, res.ProofOps.Ops, 2)
	require.Equal(t, storetypes.ProofOpIAVLCommitment, res.ProofOps.Ops[0].Type)
	require.Equal(t, storetypes.ProofOpSimpleMerkleCommitment, res.ProofOps.Ops[1].Type)

	prt := rootmulti.DefaultProofRuntime()
	require.NoError(t, prt.VerifyValue(res.ProofOps, cid.Hash, "/test/alpha", []byte("one")))

	res, err = mt.Query(&storetypes.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("absent"),
		Height: cid.Version,
		Prove:  true,
	})
	require.NoError(t, err)
	require.Len(t, res.ProofOps.Ops, 2)
	require.NoError(t, prt.VerifyAbsence(res.ProofOps, cid.Hash, "/test/absent"))

	res, err = mt.Query(&storetypes.RequestQuery{
		Path: "/test/key",
		Data: []byte("alpha"),
	})
	require.NoError(t, err)
	require.Equal(t, cid.Version, res.Height)
	require.Equal(t, []byte("one"), res.Value)

	res, err = mt.Query(&storetypes.RequestQuery{
		Path:   "/test/key",
		Data:   []byte("alpha"),
		Height: cid.Version + 10,
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Log)

	_, err = mt.Query(&storetypes.RequestQuery{
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

	res, err := mt.Query(&storetypes.RequestQuery{
		Path:   "/test/subspace",
		Data:   []byte("a/"),
		Height: cid.Version,
		Prove:  true, // prove should be ignored for /subspace
	})
	require.NoError(t, err)
	require.Nil(t, res.ProofOps)
	require.Equal(t, legacySubspaceResponse(t, updates, []byte("a/")), res.Value)

	res, err = mt.Query(&storetypes.RequestQuery{
		Path:   "/test/subspace",
		Data:   nil,
		Height: cid.Version,
	})
	require.NoError(t, err)
}

func TestCommitMultiTreeQueryInvalidPaths(t *testing.T) {
	mt, key := setupQueryableMultiTree(t)
	_ = commitQueryableTree(t, mt, key, []testKV{{key: []byte("k"), value: []byte("v")}})

	_, err := mt.Query(&storetypes.RequestQuery{
		Path: "/does-not-exist/key",
		Data: []byte("k"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such store")

	_, err = mt.Query(&storetypes.RequestQuery{
		Path: "/test/nope",
		Data: []byte("k"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected query path")

	_, err = mt.Query(&storetypes.RequestQuery{
		Path: "/test/key",
		Data: nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "query cannot be zero length")
}

func setupQueryableMultiTree(t *testing.T) (*CommitMultiTree, *storetypes.KVStoreKey) {
	t.Helper()

	mt, err := LoadCommitMultiTree(t.TempDir(), Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = mt.Close() })

	key := storetypes.NewKVStoreKey("test")
	mt.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	require.NoError(t, mt.LoadLatestVersion())

	return mt, key
}

func commitQueryableTree(t *testing.T, mt *CommitMultiTree, key storetypes.StoreKey, updates []testKV) storetypes.CommitID {
	t.Helper()

	cacheMs := mt.CacheMultiStore()
	kvStore := cacheMs.GetKVStore(key)
	for _, update := range updates {
		kvStore.Set(update.key, update.value)
	}

	committer, err := mt.StartCommit(context.Background(), cacheMs, cmtproto.Header{Time: time.Now()})
	require.NoError(t, err)

	cid, err := committer.Finalize()
	require.NoError(t, err)

	return cid
}

func legacySubspaceResponse(t *testing.T, updates []testKV, prefix []byte) []byte {
	t.Helper()

	db := dbm.NewMemDB()
	store, err := storeiavl.LoadStore(
		db,
		sdklog.NewNopLogger(),
		storetypes.NewKVStoreKey("legacy"),
		storetypes.CommitID{},
		storeiavl.DefaultIAVLCacheSize,
		false,
		metrics.NewNoOpMetrics(),
	)
	require.NoError(t, err)
	legacy := store.(*storeiavl.Store)

	for _, update := range updates {
		legacy.Set(update.key, update.value)
	}
	cid := legacy.Commit()

	res, err := legacy.Query(&storetypes.RequestQuery{
		Path:   "/subspace",
		Data:   prefix,
		Height: cid.Version,
	})
	require.NoError(t, err)

	return res.Value
}
