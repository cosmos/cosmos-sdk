package stack

import "github.com/tendermint/basecoin/state"

type prefixStore struct {
	prefix []byte
	store  state.KVStore
}

var _ state.KVStore = prefixStore{}

func (p prefixStore) Set(key, value []byte) {
	key = append(p.prefix, key...)
	p.store.Set(key, value)
}

func (p prefixStore) Get(key []byte) (value []byte) {
	key = append(p.prefix, key...)
	return p.store.Get(key)
}

// stateSpace will unwrap any prefixStore and then add the prefix
//
// this can be used by the middleware and dispatcher to isolate one space,
// then unwrap and isolate another space
func stateSpace(store state.KVStore, app string) state.KVStore {
	// unwrap one-level if wrapped
	if pstore, ok := store.(prefixStore); ok {
		store = pstore.store
	}
	return PrefixedStore(app, store)
}

func unwrap(store state.KVStore) state.KVStore {
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
func PrefixedStore(app string, store state.KVStore) state.KVStore {
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
