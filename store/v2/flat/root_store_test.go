package flat

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	types "github.com/cosmos/cosmos-sdk/store/v2"
)

var (
	skey_1  = types.NewKVStoreKey("store1")
	skey_2  = types.NewKVStoreKey("store2")
	skey_3  = types.NewKVStoreKey("store3")
	skey_1b = types.NewKVStoreKey("store1b")
	skey_2b = types.NewKVStoreKey("store2b")
	skey_3b = types.NewKVStoreKey("store3b")
)

func storeConfig123(t *testing.T) RootStoreConfig {
	opts := DefaultRootStoreConfig()
	opts.Pruning = types.PruneNothing
	err := opts.ReservePrefix(skey_1, types.StoreTypePersistent)
	require.NoError(t, err)
	err = opts.ReservePrefix(skey_2, types.StoreTypePersistent)
	require.NoError(t, err)
	err = opts.ReservePrefix(skey_3, types.StoreTypePersistent)
	require.NoError(t, err)
	return opts
}

func TestRootStoreConfig(t *testing.T) {
	opts := DefaultRootStoreConfig()
	// Ensure that no prefix conflicts are allowed
	err := opts.ReservePrefix(skey_1, types.StoreTypePersistent)
	require.NoError(t, err)
	err = opts.ReservePrefix(skey_2, types.StoreTypePersistent)
	require.NoError(t, err)
	err = opts.ReservePrefix(skey_3b, types.StoreTypePersistent)
	require.NoError(t, err)
	err = opts.ReservePrefix(skey_1b, types.StoreTypePersistent)
	require.Error(t, err)
	err = opts.ReservePrefix(skey_2b, types.StoreTypePersistent)
	require.Error(t, err)
	err = opts.ReservePrefix(skey_3, types.StoreTypePersistent)
	require.Error(t, err)
}

func TestRootStoreBasic(t *testing.T) {
	opts := DefaultRootStoreConfig()
	err := opts.ReservePrefix(skey_1, types.StoreTypePersistent)
	require.NoError(t, err)
	db := memdb.NewDB()
	store, err := NewRootStore(db, opts)
	require.NoError(t, err)

	store_1 := store.GetKVStore(skey_1)
	require.NotNil(t, store_1)
	store_1.Set([]byte{0}, []byte{0})
	val := store_1.Get([]byte{0})
	require.Equal(t, []byte{0}, val)
	store_1.Delete([]byte{0})
	val = store_1.Get([]byte{0})
	require.Equal(t, []byte(nil), val)
}

func TestRootStoreMigration(t *testing.T) {
	skey_2b := types.NewKVStoreKey("store2b")
	skey_4 := types.NewKVStoreKey("store4")

	db := memdb.NewDB()
	opts := storeConfig123(t)
	store, err := NewRootStore(db, opts)
	require.NoError(t, err)

	// write some data in all stores
	k1, v1 := []byte("first"), []byte("store")
	s1 := store.GetKVStore(skey_1)
	require.NotNil(t, s1)
	s1.Set(k1, v1)

	k2, v2 := []byte("second"), []byte("restore")
	s2 := store.GetKVStore(skey_2)
	require.NotNil(t, s2)
	s2.Set(k2, v2)

	k3, v3 := []byte("third"), []byte("dropped")
	s3 := store.GetKVStore(skey_3)
	require.NotNil(t, s3)
	s3.Set(k3, v3)

	s4 := store.GetKVStore(skey_4)
	require.Nil(t, s4)

	_ = store.Commit()
	require.NoError(t, store.Close())

	// Load without changes and make sure it is sensible
	store, err = NewRootStore(db, opts)
	require.NoError(t, err)

	// let's query data to see it was saved properly
	s2 = store.GetKVStore(skey_2)
	require.NotNil(t, s2)
	require.Equal(t, v2, s2.Get(k2))
	require.NoError(t, store.Close())

	// now, let's load with upgrades...
	opts.Upgrades = []types.StoreUpgrades{
		types.StoreUpgrades{
			Added: []string{skey_4.Name()},
			Renamed: []types.StoreRename{{
				OldKey: skey_2.Name(),
				NewKey: skey_2b.Name(),
			}},
			Deleted: []string{skey_3.Name()},
		},
	}
	restore, err := NewRootStore(db, opts)
	require.Nil(t, err)

	// s1 was not changed
	s1 = restore.GetKVStore(skey_1)
	require.NotNil(t, s1)
	require.Equal(t, v1, s1.Get(k1))

	// store3 is gone
	// TODO: breaking change? verify
	s3 = restore.GetKVStore(skey_3)
	require.Nil(t, s3)

	// store4 is mounted, with empty data
	s4 = restore.GetKVStore(skey_4)
	require.NotNil(t, s4)

	values := 0
	it := s4.Iterator(nil, nil)
	for ; it.Valid(); it.Next() {
		values += 1
	}
	require.Zero(t, values)
	require.NoError(t, it.Close())

	// write something inside store4
	k4, v4 := []byte("fourth"), []byte("created")
	s4.Set(k4, v4)

	// store2 is no longer mounted
	st2 := restore.GetKVStore(skey_2)
	require.Nil(t, st2)

	// restore2 has the old data
	rs2 := restore.GetKVStore(skey_2b)
	require.NotNil(t, rs2)
	require.Equal(t, v2, rs2.Get(k2))

	// store this migrated data, and load it again without migrations
	migratedID := restore.Commit()
	require.Equal(t, migratedID.Version, int64(2))
	require.NoError(t, restore.Close())

	// fail to load the migrated store with the old schema
	reload, err := NewRootStore(db, storeConfig123(t))
	require.Error(t, err)

	// pass a schema update with the migrations
	migratedOpts := DefaultRootStoreConfig()
	err = migratedOpts.ReservePrefix(skey_1, types.StoreTypePersistent)
	require.NoError(t, err)
	err = migratedOpts.ReservePrefix(skey_2b, types.StoreTypePersistent)
	require.NoError(t, err)
	err = migratedOpts.ReservePrefix(skey_4, types.StoreTypePersistent)
	require.NoError(t, err)
	reload, err = NewRootStore(db, migratedOpts)
	require.Nil(t, err)
	require.Equal(t, migratedID, reload.LastCommitID())

	// query this new store
	rl1 := reload.GetKVStore(skey_1)
	require.NotNil(t, rl1)
	require.Equal(t, v1, rl1.Get(k1))

	rl2 := reload.GetKVStore(skey_2b)
	require.NotNil(t, rl2)
	require.Equal(t, v2, rl2.Get(k2))

	rl4 := reload.GetKVStore(skey_4)
	require.NotNil(t, rl4)
	require.Equal(t, v4, rl4.Get(k4))
}

func TestGetVersion(t *testing.T) {
	db := memdb.NewDB()
	opts := storeConfig123(t)
	store, err := NewRootStore(db, opts)
	require.NoError(t, err)

	cid := store.Commit()
	// opts := DefaultRootStoreConfig()

	view, err := store.GetVersion(cid.Version)
	require.NoError(t, err)

	subview := view.GetKVStore(skey_1)
	require.NotNil(t, subview)

	view, err = store.GetVersion(cid.Version + 1)
	require.Equal(t, ErrVersionDoesNotExist, err)

	// ...todo
}
