package collections

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
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

type testValueCodec[T any] struct{}

func (t testValueCodec[T]) EncodeJSON(value T) ([]byte, error) {
	return t.Encode(value)
}

func (t testValueCodec[T]) DecodeJSON(b []byte) (T, error) {
	return t.Decode(b)
}

func (testValueCodec[T]) Encode(value T) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return b, nil
}
func (testValueCodec[T]) Decode(b []byte) (T, error) {
	t := new(T)
	err := json.Unmarshal(b, t)
	if err != nil {
		return *t, err
	}
	return *t, nil
}
func (testValueCodec[T]) Stringify(value T) string {
	return fmt.Sprintf("%#v", value)
}

func (testValueCodec[T]) ValueType() string { return reflect.TypeOf(*new(T)).Name() }

func newTestValueCodec[T any]() ValueCodec[T] { return testValueCodec[T]{} }

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
