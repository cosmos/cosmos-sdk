package iavl

import (
	"errors"

	"github.com/cosmos/iavl"

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
		Key:     item.Key,
		Value:   item.Value,
		Version: item.Version,
		Height:  int32(item.Height),
	}, nil
}

// Close closes the exporter.
func (e *Exporter) Close() error {
	e.exporter.Close()

	return nil
}
