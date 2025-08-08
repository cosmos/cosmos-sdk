package rootmulti

import (
	"errors"
	"fmt"
	"math"

	protoio "github.com/cosmos/gogoproto/io"
	"github.com/crypto-org-chain/cronos/memiavl"

	"cosmossdk.io/store/snapshots/types"
)

// Snapshot Implements interface Snapshotter
func (rs *Store) Snapshot(height uint64, protoWriter protoio.Writer) (returnErr error) {
	if height > math.MaxUint32 {
		return fmt.Errorf("height overflows uint32: %d", height)
	}
	version := uint32(height)

	exporter, err := memiavl.NewMultiTreeExporter(rs.dir, version, rs.supportExportNonSnapshotVersion)
	if err != nil {
		return err
	}

	defer func() {
		returnErr = errors.Join(returnErr, exporter.Close())
	}()

	for {
		item, err := exporter.Next()
		if err != nil {
			if errors.Is(err, memiavl.ErrorExportDone) {
				break
			}

			return err
		}

		switch item := item.(type) {
		case *memiavl.ExportNode:
			if err := protoWriter.WriteMsg(&types.SnapshotItem{
				Item: &types.SnapshotItem_IAVL{
					IAVL: &types.SnapshotIAVLItem{
						Key:     item.Key,
						Value:   item.Value,
						Height:  int32(item.Height),
						Version: item.Version,
					},
				},
			}); err != nil {
				return err
			}
		case string:
			if err := protoWriter.WriteMsg(&types.SnapshotItem{
				Item: &types.SnapshotItem_Store{
					Store: &types.SnapshotStoreItem{
						Name: item,
					},
				},
			}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown item type %T", item)
		}
	}

	return nil
}
