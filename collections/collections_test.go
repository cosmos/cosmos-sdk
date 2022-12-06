package collections

import (
	"math"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/mem"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
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

// checkKeyCodec asserts the correct behaviour of a KeyCodec over the type T.
func checkKeyCodec[T any](t *testing.T, encoder KeyCodec[T], key T) {
	buffer := make([]byte, encoder.Size(key))
	written, err := encoder.EncodeKey(buffer, key)
	require.NoError(t, err)
	require.Equal(t, len(buffer), written)
	read, decodedKey, err := encoder.DecodeKey(buffer)
	require.NoError(t, err)
	require.Equal(t, len(buffer), read, "encoded key and read bytes must have same size")
	require.Equal(t, key, decodedKey, "encoding and decoding produces different keys")
}

// checkValueCodec asserts the correct behaviour of a ValueCodec over the type T.
func checkValueCodec[T any](t *testing.T, encoder ValueCodec[T], value T) {
	encodedValue, err := encoder.Encode(value)
	require.NoError(t, err)
	decodedValue, err := encoder.Decode(encodedValue)
	require.NoError(t, err)
	require.Equal(t, value, decodedValue, "encoding and decoding produces different values")
}

func TestPrefix(t *testing.T) {
	t.Run("panics on invalid int", func(t *testing.T) {
		require.Panics(t, func() {
			NewPrefix(math.MaxUint8 + 1)
		})
	})

	t.Run("string", func(t *testing.T) {
		require.Equal(t, []byte("prefix"), NewPrefix("prefix").Bytes())
	})

	t.Run("int", func(t *testing.T) {
		require.Equal(t, []byte{0x1}, NewPrefix(1).Bytes())
	})

	t.Run("[]byte", func(t *testing.T) {
		bytes := []byte("prefix")
		prefix := NewPrefix(bytes)
		require.Equal(t, bytes, prefix.Bytes())
		// assert if modification happen they do not propagate to prefix
		bytes[0] = 0x0
		require.Equal(t, []byte("prefix"), prefix.Bytes())
	})
}
