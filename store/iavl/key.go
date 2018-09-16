package iavl

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.StoreKey = (*KVStoreKey)(nil)

// KVStoreKey is used for accessing substores.
// Only the pointer value should ever be used - it functions as a capabilities key.
type KVStoreKey struct {
	name string
}

// NewKVStoreKey returns a new pointer to a KVStoreKey.
// Use a pointer so keys don't collide.
func NewKey(name string) *KVStoreKey {
	return &KVStoreKey{
		name: name,
	}
}

// Implements StoreKey
func (key *KVStoreKey) Name() string {
	return key.name
}

// Implements StoreKey
func (key *KVStoreKey) String() string {
	return fmt.Sprintf("KVStoreKey{%p, %s}", key, key.name)
}

// Implements StoreKey
func (key *KVStoreKey) NewStore() types.CommitStore {
	return &Store{}
}
