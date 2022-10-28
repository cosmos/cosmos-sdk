package store_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetAndDecodeHelpers(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	cacheKVStore := cachekv.NewStore(mem)

	key1 := []byte("valid_key")                    // key1 is a key with value set in store
	value1 := sdk.AccAddress("valid_test_address") // value to be set for key1
	invalid_key := []byte("invalid_key")           // key with no value set in store

	cacheKVStore.Set(key1, value1) // set value1 to key1

	// testing GetAndDecode
	accAddr, _ := store.GetAndDecode(cacheKVStore, decodeAcc, key1) // test with valid key
	require.NotNil(t, accAddr)
	require.Equal(t, []byte(value1), cacheKVStore.Get(key1))

	accAddr, _ = store.GetAndDecode(cacheKVStore, decodeAcc, invalid_key) // test with invalid key
	require.Equal(t, accAddr, "")
	require.Equal(t, []byte(nil), cacheKVStore.Get(invalid_key))

	// testing GetAndDecodeWithBool
	accAddr, ok := store.GetAndDecodeWithBool(cacheKVStore, decodeAccWithBool, key1) // test with valid key
	require.NotNil(t, accAddr)
	require.Equal(t, []byte(value1), cacheKVStore.Get(key1))
	require.True(t, ok)

	accAddr, ok = store.GetAndDecodeWithBool(cacheKVStore, decodeAccWithBool, invalid_key) // test with invalid key
	require.Equal(t, accAddr, "")
	require.Equal(t, []byte(nil), cacheKVStore.Get(invalid_key))
	require.False(t, ok)
}

func decodeAccWithBool(bz []byte) (string, bool) {
	if bz == nil {
		return "", false
	}
	res := sdk.AccAddress(bz)
	return res.String(), true
}

func decodeAcc(bz []byte) (string, error) {
	if bz == nil {
		return "", nil
	}
	res := sdk.AccAddress(bz)
	return res.String(), nil
}
