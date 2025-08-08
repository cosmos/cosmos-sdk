//go:build objstore
// +build objstore

package rootmulti

import (
	"fmt"

	"github.com/crypto-org-chain/cronos/memiavl"

	"cosmossdk.io/store/transient"
	"cosmossdk.io/store/types"
)

// GetObjKVStore Implements interface MultiStore
func (rs *Store) GetObjKVStore(key types.StoreKey) types.ObjKVStore {
	s, ok := rs.stores[key].(types.ObjKVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not ObjKVStore", key))
	}
	return s
}

func (rs *Store) loadExtraStore(db *memiavl.DB, key types.StoreKey, params storeParams) (types.CommitStore, error) {
	if params.typ == types.StoreTypeObject {
		if _, ok := key.(*types.ObjectStoreKey); !ok {
			return nil, fmt.Errorf("unexpected key type for a ObjectStoreKey; got: %s, %T", key.String(), key)
		}

		return transient.NewObjStore(), nil
	}

	panic(fmt.Sprintf("unrecognized store type %v", params.typ))
}
