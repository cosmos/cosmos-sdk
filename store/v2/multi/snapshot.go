package multi

import (
	"fmt"
	"io"
	"sort"
	"strings"

	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Snapshot implements snapshottypes.Snapshotter.
func (rs *Store) Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error) {
	if format != snapshottypes.CurrentFormat {
		return nil, sdkerrors.Wrapf(snapshottypes.ErrUnknownFormat, "format %v", format)
	}

	if height == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "cannot snapshot height 0")
	}
	if height > uint64(rs.LastCommitID().Version) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot snapshot future height %v", height)
	}
	versions, err := rs.stateDB.Versions()
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "error while getting the snapshot versions at height %v", height)
	}
	if !versions.Exists(height) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "cannot find snapshot at height %v", height)
	}

	// get the saved snapshot at height
	vs, err := rs.getView(int64(height))
	if err != nil {
		return nil, sdkerrors.Wrap(err, fmt.Sprintf("error while get the version at height %d", height))
	}

	// Spawn goroutine to generate snapshot chunks and pass their io.ReadClosers through a channel
	ch := make(chan io.ReadCloser)
	go func() {
		// setup snapshot export stream
		chunkWriter, bufWriter, zWriter, protoWriter, err := snapshots.SetupExportStreamPipeline(ch)
		if err != nil {
			return
		}
		defer chunkWriter.Close()
		defer func() {
			if err := bufWriter.Flush(); err != nil {
				chunkWriter.CloseWithError(err)
			}
		}()
		defer func() {
			if err := zWriter.Close(); err != nil {
				chunkWriter.CloseWithError(err)
			}
		}()
		defer func() {
			if err := protoWriter.Close(); err != nil {
				chunkWriter.CloseWithError(err)
			}
		}()

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

		err = protoWriter.WriteMsg(&storetypes.SnapshotItem{
			Item: &storetypes.SnapshotItem_Schema{
				Schema: &storetypes.SnapshotSchema{
					Keys: storeByteKeys,
				},
			},
		})
		if err != nil {
			chunkWriter.CloseWithError(err)
			return
		}

		for _, sKey := range sKeys {
			subStore, err := vs.getSubstore(sKey)
			if err != nil {
				chunkWriter.CloseWithError(err)
				return
			}

			err = protoWriter.WriteMsg(&storetypes.SnapshotItem{
				Item: &storetypes.SnapshotItem_Store{
					Store: &storetypes.SnapshotStoreItem{
						Name: sKey,
					},
				},
			})
			if err != nil {
				chunkWriter.CloseWithError(err)
				return
			}

			iter := subStore.Iterator(nil, nil)
			for ; iter.Valid(); iter.Next() {
				err = protoWriter.WriteMsg(&storetypes.SnapshotItem{
					Item: &storetypes.SnapshotItem_KV{
						KV: &storetypes.SnapshotKVItem{
							Key:   iter.Key(),
							Value: iter.Value(),
						},
					},
				})
				if err != nil {
					chunkWriter.CloseWithError(err)
					return
				}
			}
			iter.Close()
		}
	}()

	return ch, nil
}

// Restore implements snapshottypes.Snapshotter.
func (rs *Store) Restore(height uint64, format uint32, chunks <-chan io.ReadCloser, ready chan<- struct{}) error {
	if err := snapshots.ValidRestoreHeight(format, height); err != nil {
		return err
	}

	versions, err := rs.stateDB.Versions()
	if err != nil {
		return sdkerrors.Wrapf(err, "error while getting the snapshot versions at height %v", height)
	}
	if versions.Count() != 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot restore snapshot for non empty store at height %v", height)
	}

	// Signal readiness. Must be done before the readers below are set up, since the zlib
	// reader reads from the stream on initialization, potentially causing deadlocks.
	if ready != nil {
		close(ready)
	}

	// Set up a restore stream pipeline
	chunkReader, zReader, protoReader, err := snapshots.SetupRestoreStreamPipeline(chunks)
	if err != nil {
		return sdkerrors.Wrap(err, "zlib failure")
	}
	defer chunkReader.Close()
	defer zReader.Close()
	defer protoReader.Close()

	var subStore *substore
	var storeSchemaReceived = false

	for {
		item := &storetypes.SnapshotItem{}
		err := protoReader.ReadMsg(item)
		if err == io.EOF {
			break
		} else if err != nil {
			return sdkerrors.Wrap(err, "invalid protobuf message")
		}

		switch item := item.Item.(type) {
		case *storetypes.SnapshotItem_Schema:
			if len(rs.schema) != 0 {
				return sdkerrors.Wrap(sdkerrors.ErrLogic, "store schema is not empty")
			}

			storeSchemaReceived = true
			schemaWriter := prefixdb.NewPrefixWriter(rs.stateTxn, schemaPrefix)
			sKeys := item.Schema.GetKeys()
			for _, sKey := range sKeys {
				rs.schema[string(sKey)] = types.StoreTypePersistent
				err := schemaWriter.Set(sKey, []byte{byte(types.StoreTypePersistent)})
				if err != nil {
					return sdkerrors.Wrap(err, "error at set the store schema key values")
				}
			}

		case *storetypes.SnapshotItem_Store:
			storeName := item.Store.GetName()
			// checking the store schema is received or not
			if !storeSchemaReceived {
				return sdkerrors.Wrapf(sdkerrors.ErrLogic, "received store name before store schema %s", storeName)
			}
			// checking the store schema exists or not
			if _, has := rs.schema[storeName]; !has {
				return sdkerrors.Wrapf(sdkerrors.ErrLogic, "store is missing from schema %s", storeName)
			}

			// get the substore
			subStore, err = rs.getSubstore(storeName)
			if err != nil {
				return sdkerrors.Wrap(err, fmt.Sprintf("error while getting the substore for key %s", storeName))
			}

		case *storetypes.SnapshotItem_KV:
			if subStore == nil {
				return sdkerrors.Wrap(sdkerrors.ErrLogic, "received KV Item before store item")
			}
			// update the key/value SMT.Store
			subStore.Set(item.KV.Key, item.KV.Value)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrLogic, "unknown snapshot item %T", item)
		}
	}

	// commit the all key/values to store
	_, err = rs.commit(height)
	if err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("error during commit the store at height %d", height))
	}

	return nil
}
