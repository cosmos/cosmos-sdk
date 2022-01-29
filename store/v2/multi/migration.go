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

// MigrateV2 will migrate the state from iavl to smt
func MigrateV2(rs *v1Store.Store, rootStore *Store) error {
	if len(rootStore.schema) != 0 {
		// schema already exists
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "smt store already have schema")
	}

	// set the schema to smt store
	schemaWriter := prefixdb.NewPrefixWriter(rootStore.stateTxn, schemaPrefix)

	for key := range rs.GetStores() {
		switch store := rs.GetCommitKVStore(key).(type) {
		case *iavl.Store:
			rootStore.schema[key.Name()] = types.StoreTypePersistent
			err := schemaWriter.Set([]byte(key.Name()), []byte{byte(types.StoreTypePersistent)})
			if err != nil {
				return sdkerrors.Wrap(err, "error at set the store schema key values")
			}

			subStore, err := rootStore.getSubstore(key.Name())
			if err != nil {
				return err
			}
			// iterate all iavl tree node key/values
			iterator := store.Iterator(nil, nil)
			for ; iterator.Valid(); iterator.Next() {
				// set the iavl key,values into smt node
				subStore.Set(iterator.Key(), iterator.Value())
			}

		case *transient.Store, *mem.Store:
			continue
		default:
			continue
		}
	}

	// commit the all key/values from iavl to smt tree (SMT Store)
	rootStore.Commit()

	return nil
}
