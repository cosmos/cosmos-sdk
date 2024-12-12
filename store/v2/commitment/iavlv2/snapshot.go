package iavlv2

import (
	"errors"

	"github.com/cosmos/iavl/v2"

	"cosmossdk.io/store/v2/commitment"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

// Exporter is a wrapper around iavl.Exporter.
type Exporter struct {
	exporter *iavl.Exporter
}

// Next returns the next item in the exporter.
func (e *Exporter) Next() (*snapshotstypes.SnapshotIAVLItem, error) {
	item, err := e.exporter.Next()
	if err != nil {
		if errors.Is(err, iavl.ErrorExportDone) {
			return nil, commitment.ErrorExportDone
		}
		return nil, err
	}

	return &snapshotstypes.SnapshotIAVLItem{
		Key:     item.Key(),
		Value:   item.Value(),
		Version: item.Version(),
		Height:  int32(item.Height()),
	}, nil
}

// Close closes the exporter.
func (e *Exporter) Close() error {
	return e.exporter.Close()
}

type Importer struct {
	importer *iavl.Importer
}

// Add adds the given item to the importer.
func (i *Importer) Add(item *snapshotstypes.SnapshotIAVLItem) error {
	return i.importer.Add(iavl.NewImportNode(item.Key, item.Value, item.Version, int8(item.Height)))
}

// Commit commits the importer.
func (i *Importer) Commit() error {
	return i.importer.Commit()
}

// Close closes the importer.
func (i *Importer) Close() error {
	i.importer.Close()

	return nil
}
