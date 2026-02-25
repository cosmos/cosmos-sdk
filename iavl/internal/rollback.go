package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/log/v2"
	cp "github.com/otiai10/copy"
)

func RollbackMultiTree(multiTreeDir string, targetVersion uint64, logger log.Logger, backupDir string) error {
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
		}

		if startVersion > latestStartVersion {
			latestStartVersion = startVersion
			latestChangesetDir = dir.Name()
		}
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
		errors.Join(
			RollbackFileToOffset(files.CheckpointsFile(), 0),
			RollbackFileToOffset(files.BranchesFile(), 0),
			RollbackFileToOffset(files.LeavesFile(), 0),
			RollbackFileToOffset(files.OrphansFile(), 0),
			RollbackFileToOffset(files.KVDataFile(), 0),
		)
	} else {
		branchesOffset := (lastGoodCp.Branches.StartOffset + lastGoodCp.Branches.Count) * SizeBranch
		leavesOffset := (lastGoodCp.Leaves.StartOffset + lastGoodCp.Leaves.Count) * SizeLeaf
		kvDataOffset := lastGoodCp.KVEndOffset
		cpInfoOffset := int64((lastGoodIdx + 1) * CheckpointInfoSize)
		// TODO clean up orphans file
		errors.Join(
			RollbackFileToOffset(files.CheckpointsFile(), cpInfoOffset),
			RollbackFileToOffset(files.BranchesFile(), int64(branchesOffset)),
			RollbackFileToOffset(files.LeavesFile(), int64(leavesOffset)),
			RollbackFileToOffset(files.KVDataFile(), int64(kvDataOffset)),
		)
	}

	return nil
}
