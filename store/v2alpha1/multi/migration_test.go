package multi

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestMigrationV2(t *testing.T) {
	r := rand.New(rand.NewSource(49872768940))

	// setup a rootmulti store
	db := dbm.NewMemDB()
	v1Store := rootmulti.NewStore(db, log.NewNopLogger())

	// mount the kvStores
	var keys []*types.KVStoreKey
	for i := uint8(0); i < 10; i++ {
		key := types.NewKVStoreKey(fmt.Sprintf("store%v", i))
		v1Store.MountStoreWithDB(key, types.StoreTypeIAVL, nil)
		keys = append(keys, key)
	}

	err := v1Store.LoadLatestVersion()
	require.Nil(t, err)

	// setup a random test data
	for _, key := range keys {
		store := v1Store.GetStore(key).(*iavl.Store)
		store.Set([]byte("temp_data"), []byte("one"))

		for i := 0; i < len(keys); i++ {
			k := make([]byte, 8)
			v := make([]byte, 1024)
			binary.BigEndian.PutUint64(k, uint64(i))
			_, err := r.Read(v)
			if err != nil {
				panic(err)
			}
			store.Set(k, v)
		}
	}

	testCases := []struct {
		testName   string
		emptyStore bool
	}{
		{
			"Migration With Empty Store",
			true,
		},
		{
			"Migration From Root Multi Store (IAVL) to SMT ",
			false,
		},
	}

	for _, testCase := range testCases {
		if !testCase.emptyStore {
			v1Store.Commit()
		}

		// setup a new root store of smt
		db2 := memdb.NewDB()
		storeConfig := DefaultStoreConfig()
		// migrating the iavl store (v1) to smt store (v2)
		v2Store, err := MigrateFromV1(v1Store, db2, storeConfig)
		require.NoError(t, err)

		for _, key := range keys {
			v2StoreKVStore := v2Store.GetKVStore(key)
			if testCase.emptyStore {
				// check the empty store
				require.Nil(t, v2StoreKVStore.Get([]byte("temp_data")))
			} else {
				require.Equal(t, v2StoreKVStore.Get([]byte("temp_data")), []byte("one"))
			}
			require.Equal(t, v2Store.LastCommitID().Version, v1Store.LastCommitID().Version)
		}
		err = v2Store.Close()
		require.NoError(t, err)
	}
}

// TestMigrateV2ForEmptyStore checking empty store migration
func TestMigrateV2ForEmptyStore(t *testing.T) {
	// setup a rootmulti store
	db := dbm.NewMemDB()
	v1Store := rootmulti.NewStore(db, log.NewNopLogger())
	err := v1Store.LoadLatestVersion()
	require.Nil(t, err)
	db2 := memdb.NewDB()
	storeConfig := DefaultStoreConfig()
	// migrating the iavl store (v1) to smt store (v2)
	v2Store, err := MigrateFromV1(v1Store, db2, storeConfig)
	require.NoError(t, err)
	require.Equal(t, v2Store.LastCommitID(), v1Store.LastCommitID())
}
