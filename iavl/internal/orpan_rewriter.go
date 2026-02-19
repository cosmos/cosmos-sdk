package internal

import (
	"fmt"
)

type OrphanRewriter struct {
	existingWriter *StructWriter[OrphanEntry]
	lastCount      int
}

func NewOrphanRewriter(existingWriter *StructWriter[OrphanEntry]) (*OrphanRewriter, error) {
	return &OrphanRewriter{
		existingWriter: existingWriter,
	}, nil
}

// Preprocess reads the existing orphan log and writes a compacted version to the new orphan log file.
// It returns a map of NodeIDs to their orphaned versions for nodes that should be deleted
// according to the provided retainCriteria function.
func (or *OrphanRewriter) Preprocess(retainCriteria RetainCriteria, compactedOrphanWriter *StructWriter[OrphanEntry]) (toDelete map[NodeID]uint32, err error) {
	rdr, err := NewStructMmap[OrphanEntry](or.existingWriter.file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rdr.Close(); err != nil {
			fmt.Printf("failed to close orphan rewriter reader: %v\n", err)
		}
	}()

	toDelete = make(map[NodeID]uint32)
	err = or.existingWriter.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush existing orphan writer: %w", err)
	}
	n := rdr.Count()
	for i := 0; i < n; i++ {
		entry := rdr.UnsafeItem(i)
		if retainCriteria(entry.NodeID.Checkpoint(), entry.OrphanedVersion) {
			// this node should be retained, so write it to the new orphan log
			if err := compactedOrphanWriter.Append(entry); err != nil {
				return nil, err
			}
		} else {
			toDelete[entry.NodeID] = entry.OrphanedVersion
		}
	}

	or.lastCount = n

	return toDelete, nil
}

func (or *OrphanRewriter) FinishRewrite(compactedOrphanWriter *StructWriter[OrphanEntry]) error {
	err := or.existingWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush existing orphan writer: %w", err)
	}

	// open new reader with new mmap
	rdr, err := NewStructMmap[OrphanEntry](or.existingWriter.file)
	if err != nil {
		return fmt.Errorf("failed to open new mmap for orphan rewriter: %w", err)
	}
	defer func() {
		if err := rdr.Close(); err != nil {
			fmt.Printf("failed to close orphan rewriter reader: %v\n", err)
		}
	}()

	newCount := rdr.Count()

	for i := or.lastCount; i < newCount; i++ {
		entry := rdr.UnsafeItem(i)
		err = compactedOrphanWriter.Append(entry)
		if err != nil {
			return err
		}
	}

	return nil
}
