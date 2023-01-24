package collections

import (
	"context"
	"math"
	"testing"

	"cosmossdk.io/core/store"
	db "github.com/cosmos/cosmos-db"

	"github.com/stretchr/testify/require"
)

type testStore struct {
	db db.DB
}

func (t testStore) OpenKVStore(ctx context.Context) store.KVStore {
	return t
}

func (t testStore) Get(key []byte) ([]byte, error) {

	return t.db.Get(key)
}

func (t testStore) Has(key []byte) (bool, error) {
	return t.db.Has(key)
}

func (t testStore) Set(key, value []byte) error {
	return t.db.Set(key, value)

}

func (t testStore) Delete(key []byte) error {
	return t.db.Delete(key)
}

func (t testStore) Iterator(start, end []byte) (store.Iterator, error) {
	return t.db.Iterator(start, end)
}

func (t testStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return t.db.ReverseIterator(start, end)
}

var _ store.KVStore = testStore{}

func deps() (store.KVStoreService, context.Context) {
	kv := db.NewMemDB()
	return &testStore{kv}, context.Background()
}

// checkKeyCodec asserts the correct behaviour of a KeyCodec over the type T.
func checkKeyCodec[T any](t *testing.T, keyCodec KeyCodec[T], key T) {
	buffer := make([]byte, keyCodec.Size(key))
	written, err := keyCodec.Encode(buffer, key)
	require.NoError(t, err)
	require.Equal(t, len(buffer), written)
	read, decodedKey, err := keyCodec.Decode(buffer)
	require.NoError(t, err)
	require.Equal(t, len(buffer), read, "encoded key and read bytes must have same size")
	require.Equal(t, key, decodedKey, "encoding and decoding produces different keys")
	// test if terminality is correctly applied
	pairCodec := PairKeyCodec(keyCodec, StringKey)
	pairKey := Join(key, "TEST")
	buffer = make([]byte, pairCodec.Size(pairKey))
	written, err = pairCodec.Encode(buffer, pairKey)
	require.NoError(t, err)
	read, decodedPairKey, err := pairCodec.Decode(buffer)
	require.NoError(t, err)
	require.Equal(t, len(buffer), read, "encoded non terminal key and pair key read bytes must have same size")
	require.Equal(t, pairKey, decodedPairKey, "encoding and decoding produces different keys with non terminal encoding")

	// check JSON
	keyJSON, err := keyCodec.EncodeJSON(key)
	require.NoError(t, err)
	decoded, err := keyCodec.DecodeJSON(keyJSON)
	require.NoError(t, err)
	require.Equal(t, key, decoded, "json encoding and decoding did not produce the same results")
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
