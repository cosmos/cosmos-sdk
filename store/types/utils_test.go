package types_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/types"
)

func initTestStores(t *testing.T) (types.KVStore, types.KVStore) {
	db := dbm.NewMemDB()
	ms := rootmulti.NewStore(db, log.NewNopLogger())

	key1 := types.NewKVStoreKey("store1")
	key2 := types.NewKVStoreKey("store2")
	require.NotPanics(t, func() { ms.MountStoreWithDB(key1, types.StoreTypeIAVL, db) })
	require.NotPanics(t, func() { ms.MountStoreWithDB(key2, types.StoreTypeIAVL, db) })
	require.NoError(t, ms.LoadLatestVersion())
	return ms.GetKVStore(key1), ms.GetKVStore(key2)
}

func TestPrefixEndBytes(t *testing.T) {
	t.Parallel()
	bs1 := []byte{0x23, 0xA5, 0x06}
	require.True(t, bytes.Equal([]byte{0x23, 0xA5, 0x07}, types.PrefixEndBytes(bs1)))
	bs2 := []byte{0x23, 0xA5, 0xFF}
	require.True(t, bytes.Equal([]byte{0x23, 0xA6}, types.PrefixEndBytes(bs2)))
	require.Nil(t, types.PrefixEndBytes([]byte{0xFF}))
	require.Nil(t, types.PrefixEndBytes(nil))
}

func TestInclusiveEndBytes(t *testing.T) {
	t.Parallel()
	require.True(t, bytes.Equal([]byte{0x00}, types.InclusiveEndBytes(nil)))
	bs := []byte("test")
	require.True(t, bytes.Equal(append(bs, byte(0x00)), types.InclusiveEndBytes(bs)))
}
