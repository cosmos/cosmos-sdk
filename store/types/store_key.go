package types

import (
	"fmt"
	"sort"
	"strings"
)

// StoreKey defines an interface for types that provide store keys.
type StoreKey interface {
	Name() string
	String() string
}

// KVStoreKey is used for providing permissioned access to module stores.
type KVStoreKey struct {
	name string
}

func NewKVStoreKey(name string) *KVStoreKey {
	if name == "" {
		panic("empty key name not allowed")
	}

	return &KVStoreKey{
		name: name,
	}
}

func NewKVStoreKeys(names ...string) map[string]*KVStoreKey {
	assertNoCommonPrefix(names)
	keys := make(map[string]*KVStoreKey, len(names))
	for _, n := range names {
		keys[n] = NewKVStoreKey(n)
	}

	return keys
}

func (key *KVStoreKey) Name() string {
	return key.name
}

func (key *KVStoreKey) String() string {
	return fmt.Sprintf("KVStoreKey{%p, %s}", key, key.name)
}

func assertNoCommonPrefix(keys []string) {
	sorted := make([]string, len(keys))

	copy(sorted, keys)
	sort.Strings(sorted)

	for i := 1; i < len(sorted); i++ {
		if strings.HasPrefix(sorted[i], sorted[i-1]) {
			panic(fmt.Errorf("potential key collision between KVStores: %s, %s", sorted[i], sorted[i-1]))
		}
	}
}
