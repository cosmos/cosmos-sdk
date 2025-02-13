package iavl

import (
	"github.com/cosmos/iavl"

	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

// Importer is a wrapper around iavl.Importer.
type Importer struct {
	importer *iavl.Importer
}

// Add adds the given item to the importer.
func (i *Importer) Add(item *snapshotstypes.SnapshotIAVLItem) error {
	return i.importer.Add(&iavl.ExportNode{
		Key:     item.Key,
		Value:   item.Value,
		Version: item.Version,
		Height:  int8(item.Height),
	})
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
