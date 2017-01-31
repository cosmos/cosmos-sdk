package trader

import "github.com/tendermint/basecoin/types"

type prefixedStore struct {
	store  types.KVStore
	prefix []byte
}

func PrefixStore(store types.KVStore, prefix []byte) types.KVStore {
	return prefixedStore{
		store:  store,
		prefix: prefix,
	}
}

func (p prefixedStore) Set(key, value []byte) {
	pkey := append(p.prefix, key...)
	p.store.Set(pkey, value)
}

func (p prefixedStore) Get(key []byte) []byte {
	pkey := append(p.prefix, key...)
	return p.store.Get(pkey)
}
