package multi

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	protoio "github.com/gogo/protobuf/io"

	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Snapshot implements snapshottypes.Snapshotter.
func (rs *Store) Snapshot(height uint64, protoWriter protoio.Writer) error {
	if height == 0 {
		return snapshottypes.ErrInvalidSnapshotVersion
	}
	if height > uint64(rs.LastCommitID().Version) {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot snapshot future height %v", height)
	}

	// get the saved snapshot at height
	vs, err := rs.getView(int64(height))
	if err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("error while get the version at height %d", height))
	}

	// sending the snapshot store schema
	var storeByteKeys [][]byte
	for sKey := range vs.schema {
		if vs.schema[sKey] == storetypes.StoreTypePersistent {
			storeByteKeys = append(storeByteKeys, []byte(sKey))
		}
	}

	sort.Slice(storeByteKeys, func(i, j int) bool {
		return bytes.Compare(storeByteKeys[i], storeByteKeys[j]) == -1
	})

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

	for _, sKey := range storeByteKeys {
		subStore, err := vs.getSubstore(string(sKey))
		if err != nil {
			return err
		}

		err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
			Item: &snapshottypes.SnapshotItem_Store{
				Store: &snapshottypes.SnapshotStoreItem{
					Name: string(sKey),
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

	if rs.LastCommitID().Version != 0 {
		return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot restore snapshot for non empty store at height %v", height)
	}

	var subStore *substore
	storeSchemaReceived := false

	var snapshotItem snapshottypes.SnapshotItem

loop:
	for {
		snapshotItem = snapshottypes.SnapshotItem{}
		err := protoReader.ReadMsg(&snapshotItem)
		if err == io.EOF {
			break
		} else if err != nil {
			return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}

		switch item := snapshotItem.Item.(type) {
		case *snapshottypes.SnapshotItem_Schema:
			receivedStoreSchema := make(StoreSchema, len(item.Schema.GetKeys()))
			storeSchemaReceived = true
			for _, sKey := range item.Schema.GetKeys() {
				receivedStoreSchema[string(sKey)] = types.StoreTypePersistent
			}

			if !rs.schema.equal(receivedStoreSchema) {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(sdkerrors.ErrLogic, "received schema does not match app schema")
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
			break loop
		}
	}

	// commit the all key/values to store
	_, err := rs.commit(height)
	if err != nil {
		return snapshotItem, sdkerrors.Wrap(err, fmt.Sprintf("error during commit the store at height %d", height))
	}

	return snapshotItem, nil
}
