package root

import (
	"bufio"
	"compress/zlib"
	"fmt"
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	protoio "github.com/gogo/protobuf/io"
	"io"
	"math"
)

// Restore implements snapshottypes.Snapshotter.
func (rs *Store) Restore(height uint64, format uint32, chunks <-chan io.ReadCloser, ready chan<- struct{}) error {
	if height == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "cannot restore snapshot at height 0")
	}
	if height > uint64(math.MaxInt64) {
		return sdkerrors.Wrapf(snapshottypes.ErrInvalidMetadata,
			"snapshot height %v cannot exceed %v", height, int64(math.MaxInt64))
	}

	versions, err := rs.stateDB.Versions()
	if versions.Count() != 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot restore snapshot for non empty store at height %v", height)
	}

	// Signal readiness. Must be done before the readers below are set up, since the zlib
	// reader reads from the stream on initialization, potentially causing deadlocks.
	if ready != nil {
		close(ready)
	}

	// Set up a restore stream pipeline
	// chan io.ReadCloser -> chunkReader -> zlib -> delimited Protobuf -> ExportNode
	chunkReader := snapshots.NewChunkReader(chunks)
	defer chunkReader.Close()
	zReader, err := zlib.NewReader(chunkReader)
	if err != nil {
		return sdkerrors.Wrap(err, "zlib failure")
	}
	defer zReader.Close()
	protoReader := protoio.NewDelimitedReader(zReader, snapshotMaxItemSize)
	defer protoReader.Close()

	var subStore *substore
	// initialisation empty store-schema for snapshot
	preg := prefixRegistry{StoreSchema: StoreSchema{}}

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
			schemaWriter := prefixdb.NewPrefixWriter(rs.stateTxn, schemaPrefix)
			sKeys := item.Schema.GetKeys()
			for _, sKey := range sKeys {
				preg.StoreSchema[string(sKey)] = types.StoreTypePersistent
				preg.reserved = append(preg.reserved, string(sKey))
				err := schemaWriter.Set(sKey, []byte{byte(types.StoreTypePersistent)})
				if err != nil {
					return sdkerrors.Wrap(err, "error at set the store schema key values")
				}
			}
			// set the new snapshot store schema to root-store
			rs.schema = preg.StoreSchema

		case *storetypes.SnapshotItem_Store:
			storeName := item.Store.GetName()
			// checking the store schema exists or not
			if _, has := rs.schema[storeName]; !has {
				return sdkerrors.Wrapf(sdkerrors.ErrLogic, "received store name before store schema %s", storeName)
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
		return sdkerrors.Wrap(err, fmt.Sprintf("error while commit the store at height %d", height))
	}

	return nil
}

// Snapshot implements snapshottypes.Snapshotter.
func (rs *Store) Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error) {
	if height == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "cannot snapshot height 0")
	}
	if height > uint64(rs.LastCommitID().Version) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot snapshot future height %v", height)
	}
	versions, err := rs.stateDB.Versions()
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
		// Set up a stream pipeline to serialize snapshot nodes:
		// ExportNode -> delimited Protobuf -> zlib -> buffer -> chunkWriter -> chan io.ReadCloser
		chunkWriter := snapshots.NewChunkWriter(ch, snapshotChunkSize)
		defer chunkWriter.Close()
		bufWriter := bufio.NewWriterSize(chunkWriter, snapshotBufferSize)
		defer func() {
			if err := bufWriter.Flush(); err != nil {
				chunkWriter.CloseWithError(err)
			}
		}()
		zWriter, err := zlib.NewWriterLevel(bufWriter, 7)
		if err != nil {
			chunkWriter.CloseWithError(sdkerrors.Wrap(err, "zlib failure"))
			return
		}
		defer func() {
			if err := zWriter.Close(); err != nil {
				chunkWriter.CloseWithError(err)
			}
		}()
		protoWriter := protoio.NewDelimitedWriter(zWriter)
		defer func() {
			if err := protoWriter.Close(); err != nil {
				chunkWriter.CloseWithError(err)
			}
		}()

		var sKeys [][]byte
		// sending the snapshot store schema
		for sKey := range vs.schema {
			sKeys = append(sKeys, []byte(sKey))
		}

		err = protoWriter.WriteMsg(&storetypes.SnapshotItem{
			Item: &storetypes.SnapshotItem_Schema{
				Schema: &storetypes.SnapshotSchema{
					Keys: sKeys,
				},
			},
		})
		if err != nil {
			chunkWriter.CloseWithError(err)
			return
		}

		for sKey := range vs.schema {
			subStore, err := vs.getSubstore(sKey)
			if err := protoWriter.Close(); err != nil {
				chunkWriter.CloseWithError(err)
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
