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
	prefix := makePrefix(app)
	return prefixStore{prefix, store}
}

func makePrefix(app string) []byte {
	return append([]byte(app), byte(0))
}

// PrefixedKey gives us the absolute path to a key that is embedded in an
// application-specific state-space.
//
// This is useful for tests or utilities that have access to the global
// state to check individual app spaces.  Individual apps should not be able
// to use this to read each other's space
func PrefixedKey(app string, key []byte) []byte {
	return append(makePrefix(app), key...)
}
