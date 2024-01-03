package mock

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

func DB() (readState store.ReadonlyState, branch func(state store.ReadonlyState) store.WritableState) {
	return memdb{kv: map[string][]byte{}}, func(state store.ReadonlyState) store.WritableState {
		return branchdb{
			parent:  state,
			changes: make(map[string][]byte),
		}
	}
}

type memdb struct {
	kv map[string][]byte
}

func (m memdb) Get(key []byte) ([]byte, error) {
	return m.kv[string(key)], nil
}

func (m memdb) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("not supported")
}

func (m memdb) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("not supported")
}

func (m memdb) Has(key []byte) (bool, error) {
	v, err := m.Get(key)
	return v != nil, err
}

type branchdb struct {
	parent  store.ReadonlyState
	changes map[string][]byte
}

func (b branchdb) Has(key []byte) (bool, error) {
	v, err := b.Get(key)
	return v != nil, err
}

func (b branchdb) Get(key []byte) ([]byte, error) {
	dirty, exists := b.changes[string(key)]
	switch {
	case exists && dirty == nil:
		return nil, nil
	case exists && dirty != nil:
		return dirty, nil
	default:
		return b.parent.Get(key)
	}
}

func (b branchdb) Iterator(start, end []byte) (corestore.Iterator, error) {
	// TODO implement me
	panic("implement me")
}

func (b branchdb) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	// TODO implement me
	panic("implement me")
}

func (b branchdb) Set(key, value []byte) error {
	b.changes[string(key)] = value
	return nil
}

func (b branchdb) Delete(key []byte) error {
	b.changes[string(key)] = nil
	return nil
}

func (b branchdb) ApplyChangeSets(changes []store.ChangeSet) error {
	for _, change := range changes {
		if change.Remove {
			err := b.Delete(change.Key)
			if err != nil {
				return err
			}
		} else {
			err := b.Set(change.Key, change.Value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b branchdb) ChangeSets() ([]store.ChangeSet, error) {
	changes := make([]store.ChangeSet, 0, len(b.changes))
	for key, value := range b.changes {
		if value == nil {
			changes = append(changes, store.ChangeSet{
				Key:    []byte(key),
				Value:  nil,
				Remove: true,
			})
		} else {
			changes = append(changes, store.ChangeSet{
				Key:    []byte(key),
				Value:  value,
				Remove: false,
			})
		}
	}
	return changes, nil
}
