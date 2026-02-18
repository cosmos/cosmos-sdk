package internal

import (
	"errors"
	"fmt"
	"io"
)

type OrphanRewriter struct {
	existingWriter *OrphanWriter
	rdr            *OrphanLogReader
}

func NewOrphanRewriter(existingWriter *OrphanWriter) (*OrphanRewriter, error) {
	rdr, err := ReadOrphanLog(existingWriter.file)
	if err != nil {
		return nil, err
	}

	return &OrphanRewriter{
		existingWriter: existingWriter,
		rdr:            rdr,
	}, nil
}

// Preprocess reads the existing orphan log and writes a compacted version to the new orphan log file.
// It returns a map of NodeIDs to their orphaned versions for nodes that should be deleted
// according to the provided retainCriteria function.
func (or *OrphanRewriter) Preprocess(retainCriteria RetainCriteria, compactedOrphanWriter *OrphanWriter) (toDelete map[NodeID]uint32, err error) {
	toDelete = make(map[NodeID]uint32)
	err = or.existingWriter.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush existing orphan writer: %w", err)
	}
	for {
		entry, err := or.rdr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return toDelete, nil
			}
			return nil, err
		}
		if retainCriteria(entry.NodeID.Checkpoint(), entry.OrphanedVersion) {
			// this node should be retained, so write it to the new orphan log
			if err := compactedOrphanWriter.WriteOrphan(entry.OrphanedVersion, entry.NodeID); err != nil {
				return nil, err
			}
		} else {
			toDelete[entry.NodeID] = entry.OrphanedVersion
		}
	}
}

func (or *OrphanRewriter) FinishRewrite(compactedOrphanWriter *OrphanWriter) error {
	err := or.existingWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush existing orphan writer: %w", err)
	}

	or.rdr.resume()

	for {
		entry, err := or.rdr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return or.rdr.Close()
			}
			return err
		}

		err = compactedOrphanWriter.WriteOrphan(entry.OrphanedVersion, entry.NodeID)
		if err != nil {
			return err
		}
	}
}
