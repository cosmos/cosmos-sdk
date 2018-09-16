package transient

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.StoreKey = (*TransientStoreKey)(nil)

// TransientStoreKey is used for indexing transient stores in a MultiStore
type TransientStoreKey struct {
	name string
}

// Constructs new TransientStoreKey
// Must return a pointer according to the ocap principle
func NewKey(name string) *TransientStoreKey {
	return &TransientStoreKey{
		name: name,
	}
}

// Implements StoreKey
func (key *TransientStoreKey) Name() string {
	return key.name
}

// Implements StoreKey
func (key *TransientStoreKey) String() string {
	return fmt.Sprintf("TransientStoreKey{%p, %s}", key, key.name)
}

// Implements StoreKey
func (key *TransientStoreKey) NewStore() types.CommitStore {
	return &Store{}
}
