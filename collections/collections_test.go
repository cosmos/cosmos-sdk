package collections

import (
	"context"
	"cosmossdk.io/core/store"
	db "github.com/tendermint/tm-db"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

type testStore struct {
	db db.DB
}

func (t testStore) OpenKVStore(ctx context.Context) store.KVStore {
	return t
}

func (t testStore) Get(key []byte) []byte {
	res, err := t.db.Get(key)
	if err != nil {
		panic(err)
	}
	return res
}

func (t testStore) Has(key []byte) bool {
	res, err := t.db.Has(key)
	if err != nil {
		panic(err)
	}
	return res
}

func (t testStore) Set(key, value []byte) {
	err := t.db.Set(key, value)
	if err != nil {
		panic(err)
	}
}

func (t testStore) Delete(key []byte) {
	err := t.db.Delete(key)
	if err != nil {
		panic(err)
	}
}

func (t testStore) Iterator(start, end []byte) store.Iterator {
	res, err := t.db.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return res
}

func (t testStore) ReverseIterator(start, end []byte) store.Iterator {
	res, err := t.db.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return res
}

var _ store.KVStore = testStore{}

func deps() (store.KVStoreService, context.Context) {
	kv := db.NewMemDB()
	return &testStore{kv}, context.Background()
}

// checkKeyCodec asserts the correct behaviour of a KeyCodec over the type T.
func checkKeyCodec[T any](t *testing.T, encoder KeyCodec[T], key T) {
	buffer := make([]byte, encoder.Size(key))
	written, err := encoder.Encode(buffer, key)
	require.NoError(t, err)
	require.Equal(t, len(buffer), written)
	read, decodedKey, err := encoder.Decode(buffer)
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
