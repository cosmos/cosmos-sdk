package stack

import (
	"bytes"
	"errors"

	"github.com/cosmos/cosmos-sdk/state"
)

type prefixStore struct {
	prefix []byte
	store  state.SimpleDB
}

var _ state.SimpleDB = prefixStore{}

func (p prefixStore) Set(key, value []byte) {
	key = append(p.prefix, key...)
	p.store.Set(key, value)
}

func (p prefixStore) Get(key []byte) (value []byte) {
	key = append(p.prefix, key...)
	return p.store.Get(key)
}

func (p prefixStore) Has(key []byte) bool {
	key = append(p.prefix, key...)
	return p.store.Has(key)
}

func (p prefixStore) Remove(key []byte) (value []byte) {
	key = append(p.prefix, key...)
	return p.store.Remove(key)
}

func (p prefixStore) List(start, end []byte, limit int) []state.Model {
	start = append(p.prefix, start...)
	end = append(p.prefix, end...)
	res := p.store.List(start, end, limit)

	trim := len(p.prefix)
	for i := range res {
		res[i].Key = res[i].Key[trim:]
	}
	return res
}

func (p prefixStore) First(start, end []byte) state.Model {
	start = append(p.prefix, start...)
	end = append(p.prefix, end...)
	res := p.store.First(start, end)
	if len(res.Key) > 0 {
		res.Key = res.Key[len(p.prefix):]
	}
	return res
}

func (p prefixStore) Last(start, end []byte) state.Model {
	start = append(p.prefix, start...)
	end = append(p.prefix, end...)
	res := p.store.Last(start, end)
	if len(res.Key) > 0 {
		res.Key = res.Key[len(p.prefix):]
	}
	return res
}

func (p prefixStore) Checkpoint() state.SimpleDB {
	return prefixStore{
		prefix: p.prefix,
		store:  p.store.Checkpoint(),
	}
}

func (p prefixStore) Commit(sub state.SimpleDB) error {
	ps, ok := sub.(prefixStore)
	if !ok {
		return errors.New("Must commit prefixStore")
	}
	if !bytes.Equal(ps.prefix, p.prefix) {
		return errors.New("Cannot commit sub-tx with different prefix")
	}

	// commit the wrapped data, don't worry about the prefix here
	p.store.Commit(ps.store)
	return nil
}

func (p prefixStore) Discard() {
	p.store.Discard()
}

// stateSpace will unwrap any prefixStore and then add the prefix
//
// this can be used by the middleware and dispatcher to isolate one space,
// then unwrap and isolate another space
func stateSpace(store state.SimpleDB, app string) state.SimpleDB {
	// unwrap one-level if wrapped
	if pstore, ok := store.(prefixStore); ok {
		store = pstore.store
	}
	return PrefixedStore(app, store)
}

func unwrap(store state.SimpleDB) state.SimpleDB {
	// unwrap one-level if wrapped
	if pstore, ok := store.(prefixStore); ok {
		store = pstore.store
	}
	return store
}

// PrefixedStore allows one to create an isolated state-space for a given
// app prefix, but it cannot easily be unwrapped
//
// This is useful for tests or utilities that have access to the global
// state to check individual app spaces.  Individual apps should not be able
// to use this to read each other's space
func PrefixedStore(app string, store state.SimpleDB) state.SimpleDB {
	prefix := append([]byte(app), byte(0))
	return prefixStore{prefix, store}
}

// PrefixedKey returns the absolute path to a given key in a particular
// app's state-space
//
// This is useful for tests or utilities that have access to the global
// state to check individual app spaces.  Individual apps should not be able
// to use this to read each other's space
func PrefixedKey(app string, key []byte) []byte {
	prefix := append([]byte(app), byte(0))
	return append(prefix, key...)
}
