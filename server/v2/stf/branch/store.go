package branch

import (
	"errors"

	"cosmossdk.io/core/store"
)

var _ store.Writer = (*Store[store.Reader])(nil)

// Store wraps an in-memory cache around an underlying types.KVStore.
type Store[T store.Reader] struct {
	changeSet changeSet // always ascending sorted
	parent    T
}

// NewStore creates a new Store object
func NewStore[T store.Reader](parent T) Store[T] {
	return Store[T]{
		changeSet: newChangeSet(),
		parent:    parent,
	}
}

// Get implements types.KVStore.
func (s Store[T]) Get(key []byte) (value []byte, err error) {
	value, found := s.changeSet.get(key)
	if found {
		return
	}
	return s.parent.Get(key)
}

// Set implements types.KVStore.
func (s Store[T]) Set(key, value []byte) error {
	if value == nil {
		return errors.New("cannot set a nil value")
	}

	s.changeSet.set(key, value)
	return nil
}

// Has implements types.KVStore.
func (s Store[T]) Has(key []byte) (bool, error) {
	tmpValue, found := s.changeSet.get(key)
	if found {
		return tmpValue != nil, nil
	}
	return s.parent.Has(key)
}

// Delete implements types.KVStore.
func (s Store[T]) Delete(key []byte) error {
	s.changeSet.delete(key)
	return nil
}

// ----------------------------------------
// Iteration

// Iterator implements types.KVStore.
func (s Store[T]) Iterator(start, end []byte) (store.Iterator, error) {
	return s.iterator(start, end, true)
}

// ReverseIterator implements types.KVStore.
func (s Store[T]) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return s.iterator(start, end, false)
}

func (s Store[T]) iterator(start, end []byte, ascending bool) (store.Iterator, error) {
	var (
		err           error
		parent, cache store.Iterator
	)

	if ascending {
		parent, err = s.parent.Iterator(start, end)
		if err != nil {
			return nil, err
		}
		cache, err = s.changeSet.iterator(start, end)
		if err != nil {
			return nil, err
		}
		return mergeIterators(parent, cache, ascending), nil
	} else {
		parent, err = s.parent.ReverseIterator(start, end)
		if err != nil {
			return nil, err
		}
		cache, err = s.changeSet.reverseIterator(start, end)
		if err != nil {
			return nil, err
		}
		return mergeIterators(parent, cache, ascending), nil
	}
}

func (s Store[T]) ApplyChangeSets(changes []store.KVPair) error {
	for _, c := range changes {
		if c.Remove {
			err := s.Delete(c.Key)
			if err != nil {
				return err
			}
		} else {
			err := s.Set(c.Key, c.Value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s Store[T]) ChangeSets() (cs []store.KVPair, err error) {
	iter, err := s.changeSet.iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		k, v := iter.Key(), iter.Value()
		cs = append(cs, store.KVPair{
			Key:    k,
			Value:  v,
			Remove: v == nil, // maybe we can optimistically compute size.
		})
	}
	return cs, nil
}
