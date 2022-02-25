package multi

import (
	"fmt"
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	protoio "github.com/gogo/protobuf/io"
	"io"
	"sort"
	"strings"
)

// Snapshot implements snapshottypes.Snapshotter.
func (rs *Store) Snapshot(height uint64, protoWriter protoio.Writer) error {
	if height == 0 {
		return snapshottypes.ErrInvalidSnapshotVersion
	}
	if height > uint64(rs.LastCommitID().Version) {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot snapshot future height %v", height)
	}
	versions, err := rs.stateDB.Versions()
	if err != nil {
		return sdkerrors.Wrapf(err, "error while getting the snapshot versions at height %v", height)
	}
	if !versions.Exists(height) {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "cannot find snapshot at height %v", height)
	}

	// get the saved snapshot at height
	vs, err := rs.getView(int64(height))
	if err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("error while get the version at height %d", height))
	}

	// schema keys
	var sKeys []string
	// sending the snapshot store schema
	for sKey := range vs.schema {
		if vs.schema[sKey] == storetypes.StoreTypePersistent {
			sKeys = append(sKeys, sKey)
		}
	}

	sort.Slice(sKeys, func(i, j int) bool {
		return strings.Compare(sKeys[i], sKeys[j]) == -1
	})

	var storeByteKeys [][]byte
	for _, sKey := range sKeys {
		storeByteKeys = append(storeByteKeys, []byte(sKey))
	}

	err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
		Item: &snapshottypes.SnapshotItem_Schema{
			Schema: &snapshottypes.SnapshotSchema{
				Keys: storeByteKeys,
			},
		},
	})
	if err != nil {
		return err
	}

	for _, sKey := range sKeys {
		subStore, err := vs.getSubstore(sKey)
		if err != nil {
			return err
		}

		err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
			Item: &snapshottypes.SnapshotItem_Store{
				Store: &snapshottypes.SnapshotStoreItem{
					Name: sKey,
				},
			},
		})
		if err != nil {
			return err
		}

		iter := subStore.Iterator(nil, nil)
		for ; iter.Valid(); iter.Next() {
			err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
				Item: &snapshottypes.SnapshotItem_KV{
					KV: &snapshottypes.SnapshotKVItem{
						Key:   iter.Key(),
						Value: iter.Value(),
					},
				},
			})
			if err != nil {
				return err
			}
		}

		err = iter.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Restore implements snapshottypes.Snapshotter.
func (rs *Store) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshottypes.SnapshotItem, error) {
	if err := snapshots.ValidRestoreHeight(format, height); err != nil {
		return snapshottypes.SnapshotItem{}, err
	}

	versions, err := rs.stateDB.Versions()
	if err != nil {
		return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(err, "error while getting the snapshot versions at height %v", height)
	}
	if versions.Count() != 0 {
		return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot restore snapshot for non empty store at height %v", height)
	}

	var subStore *substore
	var storeSchemaReceived = false

	for {
		snapshotItem := &snapshottypes.SnapshotItem{}
		err := protoReader.ReadMsg(snapshotItem)
		if err == io.EOF {
			break
		} else if err != nil {
			return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}

		switch item := snapshotItem.Item.(type) {
		case *snapshottypes.SnapshotItem_Schema:
			if len(rs.schema) != 0 {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(sdkerrors.ErrLogic, "store schema is not empty")
			}

			storeSchemaReceived = true
			schemaWriter := prefixdb.NewPrefixWriter(rs.stateTxn, schemaPrefix)
			sKeys := item.Schema.GetKeys()
			for _, sKey := range sKeys {
				rs.schema[string(sKey)] = types.StoreTypePersistent
				err := schemaWriter.Set(sKey, []byte{byte(types.StoreTypePersistent)})
				if err != nil {
					return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "error at set the store schema key values")
				}
			}

		case *snapshottypes.SnapshotItem_Store:
			storeName := item.Store.GetName()
			// checking the store schema is received or not
			if !storeSchemaReceived {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(sdkerrors.ErrLogic, "received store name before store schema %s", storeName)
			}
			// checking the store schema exists or not
			if _, has := rs.schema[storeName]; !has {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(sdkerrors.ErrLogic, "store is missing from schema %s", storeName)
			}

			// get the substore
			subStore, err = rs.getSubstore(storeName)
			if err != nil {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, fmt.Sprintf("error while getting the substore for key %s", storeName))
			}

		case *snapshottypes.SnapshotItem_KV:
			if subStore == nil {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(sdkerrors.ErrLogic, "received KV Item before store item")
			}
			// update the key/value SMT.Store
			subStore.Set(item.KV.Key, item.KV.Value)

		default:
			return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(sdkerrors.ErrLogic, "unknown snapshot item %T", item)
		}
	}

	// commit the all key/values to store
	_, err = rs.commit(height)
	if err != nil {
		return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, fmt.Sprintf("error during commit the store at height %d", height))
	}

	return snapshottypes.SnapshotItem{}, nil
}
