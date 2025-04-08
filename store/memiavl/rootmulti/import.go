package rootmulti

import (
	"fmt"
	"io"
	"math"

	"cosmossdk.io/errors"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots/types"
	protoio "github.com/cosmos/gogoproto/io"

	"github.com/crypto-org-chain/cronos/memiavl"
)

// Implements interface Snapshotter
func (rs *Store) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (types.SnapshotItem, error) {
	if rs.db != nil {
		if err := rs.db.Close(); err != nil {
			return types.SnapshotItem{}, fmt.Errorf("failed to close db: %w", err)
		}
		rs.db = nil
	}

	item, err := rs.restore(height, format, protoReader)
	if err != nil {
		return types.SnapshotItem{}, err
	}

	return item, rs.LoadLatestVersion()
}

func (rs *Store) restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (types.SnapshotItem, error) {
	importer, err := memiavl.NewMultiTreeImporter(rs.dir, height)
	if err != nil {
		return types.SnapshotItem{}, err
	}
	defer importer.Close()

	var snapshotItem types.SnapshotItem
loop:
	for {
		snapshotItem = types.SnapshotItem{}
		err := protoReader.ReadMsg(&snapshotItem)
		if err == io.EOF {
			break
		} else if err != nil {
			return types.SnapshotItem{}, errors.Wrap(err, "invalid protobuf message")
		}

		switch item := snapshotItem.Item.(type) {
		case *types.SnapshotItem_Store:
			if err := importer.AddTree(item.Store.Name); err != nil {
				return types.SnapshotItem{}, err
			}
		case *types.SnapshotItem_IAVL:
			if item.IAVL.Height > math.MaxInt8 {
				return types.SnapshotItem{}, errors.Wrapf(store.ErrLogic, "node height %v cannot exceed %v",
					item.IAVL.Height, math.MaxInt8)
			}
			node := &memiavl.ExportNode{
				Key:     item.IAVL.Key,
				Value:   item.IAVL.Value,
				Height:  int8(item.IAVL.Height),
				Version: item.IAVL.Version,
			}
			// Protobuf does not differentiate between []byte{} as nil, but fortunately IAVL does
			// not allow nil keys nor nil values for leaf nodes, so we can always set them to empty.
			if node.Key == nil {
				node.Key = []byte{}
			}
			if node.Height == 0 && node.Value == nil {
				node.Value = []byte{}
			}
			importer.AddNode(node)
		default:
			// unknown element, could be an extension
			break loop
		}
	}

	if err := importer.Finalize(); err != nil {
		return types.SnapshotItem{}, err
	}

	return snapshotItem, nil
}
