package branch

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

var _ store.WritableState = (*Store[store.ReadonlyState])(nil)

// Store wraps an in-memory cache around an underlying types.KVStore.
type Store[T store.ReadonlyState] struct {
	changeSet changeSet // always ascending sorted
	parent    T
}

// NewStore creates a new Store object
func NewStore[T store.ReadonlyState](parent T) Store[T] {
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
func (s Store[T]) Iterator(start, end []byte) (corestore.Iterator, error) {
	return s.iterator(start, end, true)
}

// ReverseIterator implements types.KVStore.
func (s Store[T]) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return s.iterator(start, end, false)
}

func (s Store[T]) iterator(start, end []byte, ascending bool) (corestore.Iterator, error) {
	var (
		err           error
		parent, cache corestore.Iterator
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

func (s Store[T]) ApplyChangeSets(changes []store.ChangeSet) error {
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

func (s Store[T]) ChangeSets() (cs []store.ChangeSet, err error) {
	iter, err := s.changeSet.iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		k, v := iter.Key(), iter.Value()
		cs = append(cs, store.ChangeSet{
			Key:    k,
			Value:  v,
			Remove: v == nil, // maybe we can optimistically compute size.
		})
	}
	return cs, nil
}
