package transient

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.KVStoreKey = (*StoreKey)(nil)

// StoreKey is used for indexing transient stores in a MultiStore
type StoreKey struct {
	name string
}

// Constructs new StoreKey
// Must return a pointer according to the ocap principle
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
func (key *StoreKey) NewStore() types.CommitKVStore {
	return &Store{}
}
