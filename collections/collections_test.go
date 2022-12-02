package collections

import (
	"github.com/cosmos/cosmos-sdk/store/mem"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ StorageProvider = (*mockStorageProvider)(nil)

type mockStorageProvider struct {
	store types.KVStore
}

func (m mockStorageProvider) KVStore(key types.StoreKey) types.KVStore {
	return m.store
}

func deps() (types.StoreKey, StorageProvider) {
	kv := mem.NewStore()
	key := types.NewKVStoreKey("test")
	return key, mockStorageProvider{store: kv}
}

func assertKeyBijective[T any](t *testing.T, encoder KeyEncoder[T], key T) {
	encodedKey, err := encoder.Encode(key)
	require.NoError(t, err)
	read, decodedKey, err := encoder.Decode(encodedKey)
	require.NoError(t, err)
	require.Equal(t, len(encodedKey), read, "encoded key and read bytes must have same size")
	require.Equal(t, key, decodedKey, "encoding and decoding produces different keys")
}

func assertValueBijective[T any](t *testing.T, encoder ValueEncoder[T], value T) {
	encodedValue, err := encoder.Encode(value)
	require.NoError(t, err)
	decodedValue, err := encoder.Decode(encodedValue)
	require.NoError(t, err)
	require.Equal(t, value, decodedValue, "encoding and decoding produces different values")
}
