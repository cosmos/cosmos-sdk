package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

// nolint - reexport
type Store = store.Store
type ReadOnlyStore = store.ReadOnlyStore
type Key = store.Key

// nolint - reexport
func NewKey(keys ...string) Key {
	return store.NewKey(keys...)
}
