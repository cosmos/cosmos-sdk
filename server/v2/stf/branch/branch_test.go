package branch

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"

	"cosmossdk.io/core/store"
)

func TestBranch(t *testing.T) {
	set := func(s interface{ Set([]byte, []byte) error }, key, value string) {
		require.NoError(t, s.Set([]byte(key), []byte(value)))
	}
	get := func(s interface{ Get([]byte) ([]byte, error) }, key, wantValue string) {
		value, err := s.Get([]byte(key))
		require.NoError(t, err)
		if wantValue == "" {
			require.Nil(t, value)
		} else {
			require.Equal(t, wantValue, string(value))
		}
	}

	remove := func(s interface{ Delete([]byte) error }, key string) {
		err := s.Delete([]byte(key))
		require.NoError(t, err)
	}

	iter := func(s interface {
		Iterator(start, end []byte) (store.Iterator, error)
	}, start, end string, wantPairs [][2]string,
	) {
		startKey := []byte(start)
		endKey := []byte(end)
		if start == "" {
			startKey = nil
		}
		if end == "" {
			endKey = nil
		}
		iter, err := s.Iterator(startKey, endKey)
		require.NoError(t, err)
		defer iter.Close()
		numPairs := len(wantPairs)
		for i := 0; i < numPairs; i++ {
			require.True(t, iter.Valid(), "expected iterator to be valid")
			gotKey, gotValue := string(iter.Key()), string(iter.Value())
			wantKey, wantValue := wantPairs[i][0], wantPairs[i][1]
			require.Equal(t, wantKey, gotKey)
			require.Equal(t, wantValue, gotValue)
			iter.Next()
		}
	}

	parent := newMemState()

	// populate parent with some state
	set(parent, "1", "a")
	set(parent, "2", "b")
	set(parent, "3", "c")
	set(parent, "4", "d")

	branch := NewStore(parent)

	get(branch, "1", "a") // gets from parent

	set(branch, "1", "z")
	get(branch, "1", "z") // gets updated value from branch

	set(branch, "5", "e")
	get(branch, "5", "e") // gets updated value from branch

	remove(branch, "3")
	get(branch, "3", "") // it's not fetched even if it exists in parent, it's not part of branch changeset currently.

	set(branch, "6", "f")
	remove(branch, "6")
	get(branch, "6", "") // inserted and then removed from branch

	// test iter
	iter(
		branch,
		"", "",
		[][2]string{
			{"1", "z"},
			{"2", "b"},
			{"4", "d"},
			{"5", "e"},
		},
	)

	// test iter in range
	iter(
		branch,
		"2", "4",
		[][2]string{
			{"2", "b"},
		},
	)

	// test reverse iter
}

func newMemState() memStore {
	return memStore{btree.NewBTreeGOptions(byKeys, btree.Options{Degree: bTreeDegree, NoLocks: true})}
}

var _ store.Writer = memStore{}

type memStore struct {
	t *btree.BTreeG[item]
}

func (m memStore) Set(key, value []byte) error {
	m.t.Set(item{key: key, value: value})
	return nil
}

func (m memStore) Delete(key []byte) error {
	m.t.Delete(item{key: key})
	return nil
}

func (m memStore) ApplyChangeSets(changes []store.KVPair) error {
	panic("not callable")
}

func (m memStore) ChangeSets() ([]store.KVPair, error) { panic("not callable") }

func (m memStore) Has(key []byte) (bool, error) {
	_, found := m.t.Get(item{key: key})
	return found, nil
}

func (m memStore) Get(bytes []byte) ([]byte, error) {
	v, found := m.t.Get(item{key: bytes})
	if !found {
		return nil, nil
	}
	return v.value, nil
}

func (m memStore) Iterator(start, end []byte) (store.Iterator, error) {
	return newMemIterator(start, end, m.t, true), nil
}

func (m memStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return newMemIterator(start, end, m.t, false), nil
}
