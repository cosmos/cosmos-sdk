package root

import (
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/mem"
	v1Store "github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/transient"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MigrationFromIAVLStoreToSMTStore will migrate the complete state from iavl to smt
func MigrationFromIAVLStoreToSMTStore(rs *v1Store.Store, rootStore *Store) error {
	// Collect stores to snapshot (only IAVL stores are supported)
	type namedStore struct {
		*iavl.Store
		name string
	}
	var stores []namedStore
	for key := range rs.GetStores() {
		switch store := rs.GetCommitKVStore(key).(type) {
		case *iavl.Store:
			stores = append(stores, namedStore{name: key.Name(), Store: store})
		case *transient.Store, *mem.Store:
			continue
		default:
			continue
		}
	}

	// make new smt store schema
	if len(rootStore.schema) != 0 {
		// schema is already exists
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "smt store already have schema")
	}

	// set the schema to smt store
	schemaWriter := prefixdb.NewPrefixWriter(rootStore.stateTxn, schemaPrefix)
	for _, store := range stores {
		rootStore.schema[store.name] = types.StoreTypePersistent
		err := schemaWriter.Set([]byte(store.name), []byte{byte(types.StoreTypePersistent)})
		if err != nil {
			return sdkerrors.Wrap(err, "error at set the store schema key values")
		}
	}

	// iterate through all iavl stores
	for _, store := range stores {
		subStore, err := rootStore.getSubstore(store.name)
		if err != nil {
			return err
		}
		// iterate all iavl tree node key/values
		iterator := store.Iterator(nil, nil)
		for ; iterator.Valid(); iterator.Next() {
			// set the iavl key,values into smt node
			subStore.Set(iterator.Key(), iterator.Value())
		}
	}
	// commit the all key/values from iavl to smt tree (SMT Store)
	rootStore.Commit()
	return nil
}
