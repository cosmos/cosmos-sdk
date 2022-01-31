package root

import (
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/mem"
	v1Store "github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/transient"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MigrateV2 will migrate the state from iavl to smt
func MigrateV2(rs *v1Store.Store, db dbm.DBConnection, storeConfig StoreConfig) (*Store, error) {
	type namedStore struct {
		*iavl.Store
		name string
	}
	var stores []namedStore
	for key := range rs.GetStores() {
		switch store := rs.GetCommitKVStore(key).(type) {
		case *iavl.Store:
			storeConfig.prefixRegistry.StoreSchema[key.Name()] = types.StoreTypePersistent
			stores = append(stores, namedStore{name: key.Name(), Store: store})
		case *transient.Store, *mem.Store:
			// Non-persisted stores shouldn't be snapshotted
			continue
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "don't know how to snapshot store %q of type %T", key.Name(), store)
		}
	}

	// creating the new store of smt tree
	rootStore, err := NewStore(db, storeConfig)
	if err != nil {
		return nil, err
	}

	// iterate through the rootmulti stores and save the key/values into smt tree
	for _, store := range stores {
		subStore, err := rootStore.getSubstore(store.name)
		if err != nil {
			return nil, err
		}
		// iterate all iavl tree node key/values
		iterator := store.Iterator(nil, nil)
		for ; iterator.Valid(); iterator.Next() {
			// set the iavl key,values into smt node
			subStore.Set(iterator.Key(), iterator.Value())
		}
	}

	// commit the all key/values from iavl to smt tree (SMT Store)
	_, err = rootStore.commit(uint64(rs.LastCommitID().Version))
	if err != nil {
		return nil, err
	}

	return rootStore, nil
}
