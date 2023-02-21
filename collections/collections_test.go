package collections

import (
	"context"
	"errors"
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
	itr, err := t.db.Iterator(start, end)
	return coreIterator{itr}, err
}

func (t testStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	itr, err := t.db.ReverseIterator(start, end)
	return coreIterator{itr}, err
}

var _ store.KVStore = testStore{}

func deps() (store.KVStoreService, context.Context) {
	kv := db.NewMemDB()
	return &testStore{kv}, context.Background()
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

var _ store.Iterator = coreIterator{}

type coreIterator struct {
	iterator db.Iterator
}

func NewCoreIterator(iterator db.Iterator) (coreIterator, error) {
	return coreIterator{iterator}, nil
}

// Domain implements Iterator.
func (itr coreIterator) Domain() ([]byte, []byte) {
	return itr.iterator.Domain()
}

// Valid implements Iterator.
func (itr coreIterator) Valid() bool {
	return itr.iterator.Valid()
}

// Key implements Iterator.
func (itr coreIterator) Key() ([]byte, error) {
	// Key returns a copy of the current key.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	if !itr.Valid() {
		return []byte{}, errors.New("iterator is invalid")
	}

	return itr.iterator.Key(), nil
}

// Value implements Iterator.
func (itr coreIterator) Value() ([]byte, error) {
	// Value returns a copy of the current value.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	if !itr.Valid() {
		return []byte{}, errors.New("iterator is invalid")
	}

	return itr.iterator.Value(), nil
}

// Next implements Iterator.
func (itr coreIterator) Next() error {
	if !itr.Valid() {
		return errors.New("iterator is invalid")
	}
	itr.iterator.Next()

	return nil
}

// Error implements Iterator.
func (itr coreIterator) Error() error {
	return itr.iterator.Error()
}

// Close implements Iterator.
func (itr coreIterator) Close() error {
	return itr.iterator.Close()
}
