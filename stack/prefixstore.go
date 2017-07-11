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
func stateSpace(store state.KVStore, app string) state.KVStore {
	// unwrap one-level if wrapped
	if pstore, ok := store.(prefixStore); ok {
		store = pstore.store
	}
	// wrap it with the prefix
	prefix := append([]byte(app), byte(0))
	return prefixStore{prefix, store}
}
