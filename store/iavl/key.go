package iavl

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.StoreKey = (*StoreKey)(nil)

// StoreKey is used for accessing substores.
// Only the pointer value should ever be used - it functions as a capabilities key.
type StoreKey struct {
	name string
}

// NewStoreKey returns a new pointer to a StoreKey.
// Use a pointer so keys don't collide.
func NewKey(name string) *StoreKey {
	return &StoreKey{
		name: name,
	}
}

// Implements StoreKey
func (key *StoreKey) Name() string {
	return key.name
}

// Implements StoreKey
func (key *StoreKey) String() string {
	return fmt.Sprintf("StoreKey{%p, %s}", key, key.name)
}

// Implements StoreKey
func (key *StoreKey) NewStore() types.CommitStore {
	return &Store{}
}
