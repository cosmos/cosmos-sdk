package internal

import (
	"fmt"
)

// OrphanRewriter handles orphan entries during compaction.
//
// Compaction needs to do two things with orphans:
//  1. Decide which orphaned nodes to prune (delete from data files) vs retain (keep for historical queries).
//  2. Copy retained orphan entries to the new compacted changeset's orphan file.
//
// The tricky part is that new orphans can be written to the source changeset WHILE compaction
// is running (the live OrphanProcessor is still active). OrphanRewriter handles this with a
// two-phase approach:
//   - Preprocess: reads all orphan entries that exist at the start of compaction. Retained entries
//     are copied to the new file; prunable entries go into the deleteMap. Records lastCount.
//   - FinishRewrite: reads any NEW orphan entries that were appended since Preprocess (from
//     lastCount to current count). These are always retained — they were just written and
//     are too new to prune. This runs during the switchover while the orphan proc lock is held.
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
	rdr, err := or.openReader()
	if err != nil {
		return nil, err
	}
	defer rdr.Close()

	toDelete = make(map[NodeID]uint32)
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
	rdr, err := or.openReader()
	if err != nil {
		return err
	}
	defer rdr.Close()

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

func (or *OrphanRewriter) openReader() (*StructMmap[OrphanEntry], error) {
	err := or.existingWriter.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush existing orphan writer: %w", err)
	}
	return NewStructMmap[OrphanEntry](or.existingWriter.file)
}
