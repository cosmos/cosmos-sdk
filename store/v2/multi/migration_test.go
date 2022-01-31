package root

import (
	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

func TestMigrationV2(t *testing.T) {
	storeName := "store1"
	kvStoreKey := types.NewKVStoreKey(storeName)

	// setup a rootmulti store
	db := dbm.NewMemDB()
	v1Store := rootmulti.NewStore(db)
	v1Store.MountStoreWithDB(kvStoreKey, types.StoreTypeIAVL, db)
	err := v1Store.LoadLatestVersion()
	require.Nil(t, err)

	// setup a temporary test data
	v1StoreKVStore := v1Store.GetKVStore(kvStoreKey)
	v1StoreKVStore.Set([]byte("temp_data"), []byte("one"))
	v1Store.Commit()

	// setup a new root store of smt
	db2 := memdb.NewDB()
	// migrating the iavl store (v1) to smt store (v2)
	v2Store, err := MigrateV2(v1Store, db2, DefaultStoreConfig())
	require.NoError(t, err)

	v2StoreKVStore := v2Store.GetKVStore(kvStoreKey)
	require.Equal(t, v2StoreKVStore.Get([]byte("temp_data")), []byte("one"))
	require.Equal(t, v2Store.LastCommitID().Version,v1Store.LastCommitID().Version)
}
