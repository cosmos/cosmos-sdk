package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/btree"
)

// load reconstructs the tree's in-memory state from disk on startup.
//
// This is where crash recovery happens. The on-disk state consists of:
//
//  1. Changeset directories — each contains a WAL (write-ahead log), checkpoint files
//     (branches.dat, leaves.dat, kv.dat, checkpoints.dat), and metadata. Changesets are
//     organized by version range and may be compacted together over time.
//
//  2. Checkpoints — periodic snapshots of the full tree structure written to the changeset's
//     data files. Loading from a checkpoint is fast (just mmap the files), but checkpoints may
//     lag behind the latest committed version because they're written in the background.
//
//  3. WAL files — append-only logs of every set/delete operation. The WAL is the source of
//     truth for durability. After loading the latest checkpoint, we replay WAL entries forward
//     to reach the committed version.
//
// Recovery works as follows:
//   - First, clean up incomplete state: delete any -tmp directories (from interrupted compactions)
//     and run VerifyAndFix on each changeset (which rolls back incomplete checkpoints).
//   - Load the latest checkpoint to get a tree root at some version.
//   - Replay WAL entries from that checkpoint version forward to reach the expected version.
//   - If the WAL contains entries BEYOND the expected version (from a crash mid-commit before
//     the commit info file was renamed), auto-repair truncates those entries away.
//   - The expected version comes from the commit info file at the CommitMultiTree level —
//     if that file doesn't exist for a version, that version was never durably committed,
//     and any WAL entries for it are safely discarded.
func (ts *TreeStore) load() error {
	ctx, span := tracer.Start(context.Background(), "TreeStore.load")
	defer span.End()

	// Scan the tree's directory for changeset subdirectories.
	dirs, err := os.ReadDir(ts.dir)
	if err != nil {
		return fmt.Errorf("failed to read tree store dir: %w", err)
	}

	// Build a map of: startVersion -> compactedAt -> dirName
	// This lets us find the best (most compacted) directory for each version range.
	var dirMap btree.Map[uint32, *btree.Map[uint32, string]]
	for _, de := range dirs {
		if !de.IsDir() {
			continue
		}

		dirName := de.Name()

		// -tmp directories are leftovers from interrupted compaction operations.
		// Compaction writes to a -tmp dir first, then renames atomically when done.
		// If we find one here, the compaction never completed — safe to delete.
		if strings.HasSuffix(dirName, "-tmp") {
			dir := filepath.Join(ts.dir, dirName)
			ts.logger.WarnContext(ctx, "found incomplete changeset dir, deleting", "dir", dir)
			err := os.RemoveAll(dir)
			if err != nil {
				ts.logger.ErrorContext(ctx, "failed to remove incomplete changeset dir", "dir", dir, "error", err)
			}
			continue
		}

		startVersion, _, compactedAt, valid := ParseChangesetDirName(dirName)
		if !valid {
			continue
		}

		dir := filepath.Join(ts.dir, dirName)
		if _, found := dirMap.Get(startVersion); !found {
			dirMap.Set(startVersion, &btree.Map[uint32, string]{})
		}
		caMap, _ := dirMap.Get(startVersion)
		caMap.Set(compactedAt, dir)
	}

	// Load changesets in version order.
	// For each startVersion, we pick the most-compacted directory (highest compactedAt)
	// and discard older versions.
	for {
		startVersion, compactionMap, ok := dirMap.PopMin()
		if !ok {
			break
		}
		// PopMax gives us the most recent compaction for this version range.
		_, dirName, ok := compactionMap.PopMax()
		if !ok {
			return fmt.Errorf("internal error: no changeset entries for start version %d", startVersion)
		}

		ts.logger.DebugContext(ctx, "loading changeset", "startVersion", startVersion, "dir", dirName)

		// OpenChangeset opens the changeset files and runs VerifyAndFix if autoRepair is enabled.
		// VerifyAndFix checks that checkpoint data is consistent and rolls back the last checkpoint
		// if it's incomplete (e.g. from a crash during background checkpoint writing).
		// This is safe because checkpoints are an optimization — rolling one back just means we'll
		// replay a bit more WAL on startup, but no data is lost.
		cs, err := OpenChangeset(ts, dirName, !ts.opts.DisableAutoRepair)
		if err != nil {
			return fmt.Errorf("failed to open changeset in %s: %w", dirName, err)
		}

		// Index the changeset by its start version for later WAL replay.
		ts.changesetsByVersion.Set(startVersion, cs)

		// Also index by checkpoint number so the checkpointer can find roots by checkpoint ID.
		rdr, pin := cs.TryPinReader()
		if rdr == nil {
			return fmt.Errorf("changeset reader is not available for changeset starting at version %d", startVersion)
		}
		firstCheckpoint := rdr.FirstCheckpoint()
		if firstCheckpoint != 0 {
			if firstCheckpoint <= ts.checkpointer.savedCheckpoint.Load() {
				return fmt.Errorf("found duplicate or out-of-order checkpoint %d in changeset starting at version %d", firstCheckpoint, startVersion)
			}
			ts.checkpointer.changesetsByCheckpoint.Set(firstCheckpoint, cs)
			lastCheckpoint := rdr.LastCheckpoint()
			ts.checkpointer.savedCheckpoint.Store(lastCheckpoint)
			ts.lastCheckpoint.Store(lastCheckpoint)
			info, err := rdr.GetCheckpointInfo(lastCheckpoint)
			if err != nil {
				return fmt.Errorf("failed to get checkpoint info for checkpoint %d in changeset starting at version %d: %w", lastCheckpoint, startVersion, err)
			}
			ts.lastCheckpointVersion = info.Version
		}
		pin.Unpin()

		// Clean up superseded compaction directories for the same start version.
		// These are older compaction outputs that have been replaced by a newer compaction.
		compactionMap.Ascend(0, func(_ uint32, dir string) bool {
			ts.logger.WarnContext(ctx, "deleting superseded changeset dir", "dir", dir)
			if err := os.RemoveAll(dir); err != nil {
				ts.logger.ErrorContext(ctx, "failed to delete superseded changeset dir", "dir", dir, "error", err)
			}
			return true
		})

		// If this changeset was produced by compaction (has an endVersion), it subsumes
		// all older changesets in that range — delete them.
		endVersion := cs.Files().EndVersion()
		if endVersion > 0 {
			for {
				nextStart, compactionMap, ok := dirMap.Min()
				if !ok || nextStart > endVersion {
					break
				}
				compactionMap.Ascend(0, func(_ uint32, dir string) bool {
					ts.logger.WarnContext(ctx, "deleting compacted changeset dir", "dir", dir)
					if err := os.RemoveAll(dir); err != nil {
						ts.logger.ErrorContext(ctx, "failed to delete compacted changeset dir", "dir", dir, "error", err)
					}
					return true
				})
				dirMap.Delete(nextStart)
			}
		}
	}

	// --- Reconstruct the tree from checkpoint + WAL replay ---

	// Start from the latest checkpoint. This gives us a tree root at some version
	// that may be behind the latest committed version.
	cpInfo, err := ts.checkpointer.LatestCheckpointRoot()
	if err != nil {
		return fmt.Errorf("failed to load root after loading changesets: %w", err)
	}

	root := cpInfo.Root
	version := cpInfo.Version
	// Sanity check: the checkpoint should never be ahead of the expected version.
	// If it is, something is seriously wrong (e.g. data corruption or mismatched data directories).
	if ts.opts.ExpectedVersion != 0 && version > ts.opts.ExpectedVersion {
		return fmt.Errorf("latest checkpoint version %d is greater than expected version %d, this indicates data corruption", version, ts.opts.ExpectedVersion)
	}

	// Find the changeset whose WAL contains the entries we need to replay
	// (from checkpoint version forward to the expected version).
	replayFrom := ts.changesetForVersion(version + 1)
	var replayFromVersion uint32 // default to 0
	if replayFrom != nil {
		replayFromVersion = replayFrom.Files().StartVersion()
	}

	// Replay WAL entries forward from the checkpoint version.
	// ReplayWALForStartup applies each set/delete from the WAL to the in-memory tree.
	//
	// If it encounters entries exactly ONE version beyond the expected version (N+1),
	// it truncates them away (auto-repair). This handles the normal crash scenario:
	// a commit was in progress (WAL was being written for version N+1) but the process
	// crashed before the commit info file was renamed, so version N+1 was never durably committed.
	//
	// If it encounters entries MORE than one version beyond expected (N+2 or higher),
	// that's treated as unrecoverable corruption and returns an error. This should never
	// happen in normal operation because commits are sequential — you can only be writing
	// WAL entries for the next version (N+1), never for N+2, when a crash occurs.
	// Seeing N+2 means something went seriously wrong (data directory mix-up, disk corruption, etc.).
	ts.changesetsByVersion.Ascend(replayFromVersion, func(_ uint32, cs *Changeset) bool {
		var repaired bool
		root, version, repaired, err = ReplayWALForStartup(ctx, root, cs.files.WALFile(), version, ts.opts.ExpectedVersion, ts.logger, !ts.opts.DisableAutoRepair)
		if err != nil {
			return false
		}
		if repaired {
			// If we truncated WAL entries, reload the changeset reader so it reflects
			// the repaired file state (the reader mmaps the file, so stale mappings would be wrong).
			err = cs.openNewReader()
			if err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return fmt.Errorf("failed to replay WAL after loading changesets: %w", err)
	}

	// Set the in-memory root to the fully-replayed tree.
	ts.root.Store(&versionedRoot{
		root:    root,
		version: version,
	})

	return nil
}
