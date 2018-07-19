package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

type Store = store.Store
type ReadOnlyStore = store.ReadOnlyStore
type Key = store.Key

func NewKey(keys ...string) Key {
	return store.NewKey(keys...)
}
