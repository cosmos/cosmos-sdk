package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cosmossdk.io/log/v2"
	cp "github.com/otiai10/copy"
)

// RollbackMultiTree performs an OFFLINE rollback of an entire multi-tree to a target version.
// This is designed to be run while the node is stopped, via the `iavl rollback` CLI tool.
//
// The rollback:
//  1. Backs up and removes commit info files for versions beyond the target.
//  2. For each IAVL tree store, calls RollbackTree to truncate WALs and checkpoint data.
//
// After rollback, the node can be restarted normally — the standard load() path will reconstruct
// each tree from the remaining checkpoints + WAL replay.
//
// All original files are moved to backupDir (not deleted) so the rollback can be undone manually
// if needed by moving them back.
func RollbackMultiTree(multiTreeDir string, targetVersion uint64, logger log.Logger, backupDir string) error {
	if backupDir == "" {
		backupDir = filepath.Join(multiTreeDir, fmt.Sprintf("bak-%s", time.Now().Format("20060102150405")))
	}

	logger.Info("Rolling back multi-tree", "dir", multiTreeDir, "targetVersion", targetVersion)
	err := rollbackCommitInfos(multiTreeDir, targetVersion, logger, backupDir)
	if err != nil {
		return fmt.Errorf("failed to rollback commit infos: %w", err)
	}

	storesDir := filepath.Join(multiTreeDir, "stores")
	dirs, err := os.ReadDir(storesDir)
	if err != nil {
		return fmt.Errorf("failed to read stores dir: %w", err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if !strings.HasSuffix(dir.Name(), ".iavl") {
			continue
		}
		treeDir := filepath.Join(storesDir, dir.Name())
		backupDir := filepath.Join(backupDir, "stores", dir.Name())
		err := os.MkdirAll(backupDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create backup dir for tree %s: %w", dir.Name(), err)
		}
		err = RollbackTree(treeDir, targetVersion, logger, backupDir)
		if err != nil {
			return fmt.Errorf("failed to rollback tree in dir %s: %w", treeDir, err)
		}
	}

	return nil
}

// RollbackTree performs an offline rollback of a single IAVL tree to a target version.
// It operates directly on the filesystem — no tree stores or in-memory structures are involved:
//  1. Moves changeset directories with startVersion > targetVersion to backupDir.
//  2. Backs up and then truncates the latest changeset's WAL (removes entries beyond targetVersion).
//  3. Truncates checkpoint data files to remove checkpoints beyond targetVersion.
//
// The tree will be reconstructed from the truncated files on next startup via load().
func RollbackTree(treeDir string, targetVersion uint64, logger log.Logger, backupDir string) error {
	logger.Info("Rolling back tree", "dir", treeDir, "targetVersion", targetVersion)
	dirs, err := os.ReadDir(treeDir)
	if err != nil {
		return fmt.Errorf("failed to read tree dir: %w", err)
	}

	var latestStartVersion uint32
	var latestChangesetDir string
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		startVersion, _, _, valid := ParseChangesetDirName(dir.Name())
		if !valid {
			continue
		}
		if uint64(startVersion) > targetVersion {
			backupPath := filepath.Join(backupDir, dir.Name())
			logger.Info("Deleting changeset", "treeDir", treeDir, "dir", dir.Name(), "startVersion", startVersion, "backupPath", backupPath)
			err := os.Rename(filepath.Join(treeDir, dir.Name()), backupPath)
			if err != nil {
				return fmt.Errorf("failed to move old changeset dir %s to backup: %w", dir.Name(), err)
			}
		} else if startVersion > latestStartVersion {
			latestStartVersion = startVersion
			latestChangesetDir = dir.Name()
		}
	}
	if latestChangesetDir == "" {
		return fmt.Errorf("no valid changeset dir found in tree dir %s", treeDir)
	}

	backupPath := filepath.Join(backupDir, latestChangesetDir)
	dir := filepath.Join(treeDir, latestChangesetDir)
	logger.Info("Rolling back latest changeset dir", "dir", latestChangesetDir, "startVersion", latestStartVersion)
	// copy everything to backup first before rolling back any files, so that if something goes wrong during the rollback we still have the original files in backup
	err = os.MkdirAll(backupPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create backup dir for latest changeset: %w", err)
	}
	err = cp.Copy(dir, backupPath)
	if err != nil {
		return fmt.Errorf("failed to copy latest changeset dir %s to backup: %w", latestChangesetDir, err)
	}

	files, err := OpenChangesetFiles(dir)
	defer files.Close()
	if err != nil {
		return fmt.Errorf("failed to open changeset files in dir %s: %w", latestChangesetDir, err)
	}

	logger.Info("Rolling back changeset files in latest changeset", "dir", dir, "targetVersion", targetVersion)
	logger.Info("Rolling back WAL file", "file", files.WALFile().Name(), "targetVersion", targetVersion)
	err = RollbackWAL(files.WALFile(), targetVersion)
	if err != nil {
		return fmt.Errorf("failed to rollback WAL file: %w", err)
	}

	// rollback checkpoints
	err = rollbackCheckpoints(files, uint32(targetVersion), logger)
	if err != nil {
		return fmt.Errorf("failed to rollback checkpoint files: %w", err)
	}

	return nil
}

func rollbackCommitInfos(multiTreeDir string, targetVersion uint64, logger log.Logger, backupDir string) error {
	logger.Info("Deleting commit infos after target version", "dir", multiTreeDir, "targetVersion", targetVersion)
	commitInfoDir := filepath.Join(multiTreeDir, commitInfoSubPath)
	backupCommitInfoDir := filepath.Join(backupDir, commitInfoSubPath)
	err := os.MkdirAll(backupCommitInfoDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create backup commit info dir: %w", err)
	}

	entries, err := os.ReadDir(commitInfoDir)
	if err != nil {
		return fmt.Errorf("failed to read commit info dir: %w", err)
	}

	for _, entry := range entries {
		var version uint64
		_, err := fmt.Sscanf(entry.Name(), "%d", &version)
		if err != nil {
			continue
		}
		if version > targetVersion {
			logger.Info("Deleting commit info file",
				"file", entry.Name(),
				"version", version,
				"targetVersion", targetVersion,
				"backupPath", filepath.Join(backupCommitInfoDir, entry.Name()),
			)
			err := os.Rename(filepath.Join(commitInfoDir, entry.Name()), filepath.Join(backupCommitInfoDir, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to move old commit info file %s to backup: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

// rollbackCheckpoints truncates checkpoint data files to remove all checkpoints beyond targetVersion.
// It reads checkpoints.dat to find the last checkpoint at or before targetVersion, then truncates
// branches.dat, leaves.dat, kv.dat, and checkpoints.dat to the offsets recorded in that checkpoint.
// If no valid checkpoint exists at or before targetVersion, all files are truncated to zero.
func rollbackCheckpoints(files *ChangesetFiles, targetVersion uint32, logger log.Logger) error {
	cpInfos, err := NewStructMmap[CheckpointInfo](files.CheckpointsFile())
	if err != nil {
		return fmt.Errorf("failed to create mmap for checkpoint infos: %w", err)
	}
	defer cpInfos.Close()

	n := cpInfos.Count()
	var lastGoodIdx int
	var lastGoodCp *CheckpointInfo
	for i := 0; i < n; i++ {
		cpInfo := cpInfos.UnsafeItem(i)
		if cpInfo.Version <= targetVersion {
			lastGoodIdx = i
			lastGoodCp = cpInfo
		} else {
			// we want to delete all checkpoint infos after the last good index, so we can just break here and then truncate the file
			break
		}
	}

	if lastGoodCp == nil {
		// just truncate everything to zero
		err = errors.Join(
			RollbackFileToOffset(files.CheckpointsFile(), 0),
			RollbackFileToOffset(files.BranchesFile(), 0),
			RollbackFileToOffset(files.LeavesFile(), 0),
			RollbackFileToOffset(files.KVDataFile(), 0),
			RollbackFileToOffset(files.OrphansFile(), 0),
		)
	} else {
		branchesOffset := (lastGoodCp.Branches.StartOffset + lastGoodCp.Branches.Count) * SizeBranch
		leavesOffset := (lastGoodCp.Leaves.StartOffset + lastGoodCp.Leaves.Count) * SizeLeaf
		kvDataOffset := lastGoodCp.KVEndOffset
		cpInfoOffset := int64((lastGoodIdx + 1) * CheckpointInfoSize)
		err = errors.Join(
			RollbackFileToOffset(files.CheckpointsFile(), cpInfoOffset),
			RollbackFileToOffset(files.BranchesFile(), int64(branchesOffset)),
			RollbackFileToOffset(files.LeavesFile(), int64(leavesOffset)),
			RollbackFileToOffset(files.KVDataFile(), int64(kvDataOffset)),
			// TODO clean up orphans file preserving valid orphans, for now just nuke it and leave some garbage
			RollbackFileToOffset(files.OrphansFile(), 0),
		)
	}
	if err != nil {
		return fmt.Errorf("failed to rollback checkpoint files: %w", err)
	}

	return nil
}
