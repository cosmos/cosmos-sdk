package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ChangesetFiles encapsulates management of changeset files.
// This type is shared between the Changeset and ChangesetWriter types.
type ChangesetFiles struct {
	dir          string
	treeDir      string
	startVersion uint32
	endVersion   uint32 // 0 if original changeset
	compactedAt  uint32

	walFile         *os.File
	kvDataFile      *os.File
	branchesFile    *os.File
	leavesFile      *os.File
	checkpointsFile *os.File
	orphansFile     *os.File

	closed bool
}

// CreateChangesetFiles creates a new changeset directory and files that are ready to be written to.
// If compactedAt is 0, the changeset is considered original and uncompacted.
// If compactedAt is greater than 0, the changeset is considered compacted and will be suffixed with -tmp
// until MarkReady is called.
func CreateChangesetFiles(treeDir string, startVersion, compactedAt uint32) (*ChangesetFiles, error) {
	// ensure absolute path
	var err error
	treeDir, err = filepath.Abs(treeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", treeDir, err)
	}

	dirName := fmt.Sprintf("%d", startVersion)
	if compactedAt > 0 {
		dirName = fmt.Sprintf("%d.%d-tmp", startVersion, compactedAt)
	}
	dir := filepath.Join(treeDir, dirName)

	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		return nil, fmt.Errorf("failed to create changeset dir: %w", err)
	}

	cr := &ChangesetFiles{
		dir:          dir,
		treeDir:      treeDir,
		startVersion: startVersion,
		compactedAt:  compactedAt,
	}

	err = cr.open(writeModeFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	return cr, nil
}

const writeModeFlags = os.O_RDWR | os.O_CREATE | os.O_APPEND

// OpenChangesetFiles opens an existing changeset directory and files.
// All files are opened in readonly mode, except for orphans.dat and info.dat which are opened in read-write mode
// to track orphan data and statistics.
func OpenChangesetFiles(dirName string) (*ChangesetFiles, error) {
	startVersion, endVersion, compactedAt, valid := ParseChangesetDirName(filepath.Base(dirName))
	if !valid {
		return nil, fmt.Errorf("invalid changeset dir name: %s", dirName)
	}

	dir, err := filepath.Abs(dirName)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", dirName, err)
	}

	treeDir := filepath.Dir(dir)

	cr := &ChangesetFiles{
		dir:          dir,
		treeDir:      treeDir,
		startVersion: startVersion,
		endVersion:   endVersion,
		compactedAt:  compactedAt,
	}

	err = cr.open(os.O_RDONLY)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	return cr, nil
}

func (cr *ChangesetFiles) open(mode int) error {
	var err error

	walPath := filepath.Join(cr.dir, "wal.log")
	cr.walFile, err = os.OpenFile(walPath, mode, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open WAL data file: %w", err)
	}

	kvPath := filepath.Join(cr.dir, "kv.dat")
	cr.kvDataFile, err = os.OpenFile(kvPath, mode, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open KV data file: %w", err)
	}

	leavesPath := filepath.Join(cr.dir, "leaves.dat")
	cr.leavesFile, err = os.OpenFile(leavesPath, mode, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open leaves data file: %w", err)
	}

	branchesPath := filepath.Join(cr.dir, "branches.dat")
	cr.branchesFile, err = os.OpenFile(branchesPath, mode, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open branches data file: %w", err)
	}

	checkpointsPath := filepath.Join(cr.dir, "checkpoints.dat")
	cr.checkpointsFile, err = os.OpenFile(checkpointsPath, mode, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open checkpoints data file: %w", err)
	}

	orphansPath := filepath.Join(cr.dir, "orphans.dat")
	cr.orphansFile, err = os.OpenFile(orphansPath, writeModeFlags, 0o600) // the orphans file is always opened for writing
	if err != nil {
		return fmt.Errorf("failed to open orphans data file: %w", err)
	}

	return nil
}

// ParseChangesetDirName parses a changeset directory name and returns the start version,
// and end version plus compacted at version (if these are available).
// If the directory name is invalid, valid will be false.
// If a changeset is original and uncompacted, endVersion and compactedAt will be 0.
func ParseChangesetDirName(dirName string) (startVersion, endVersion, compactedAt uint32, valid bool) {
	var err error
	var v uint64
	// if no dot, it's an original changeset
	if !strings.Contains(dirName, ".") {
		v, err = strconv.ParseUint(dirName, 10, 32)
		if err != nil {
			return 0, 0, 0, false
		}
		return uint32(v), 0, 0, true
	}

	parts := strings.Split(dirName, ".")
	if len(parts) != 2 {
		return 0, 0, 0, false
	}

	spanParts := strings.Split(parts[0], "-")
	if len(spanParts) != 2 {
		return 0, 0, 0, false
	}

	v, err = strconv.ParseUint(spanParts[0], 10, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	startVersion = uint32(v)

	v, err = strconv.ParseUint(spanParts[1], 10, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	endVersion = uint32(v)

	v, err = strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	compactedAt = uint32(v)

	return startVersion, endVersion, compactedAt, true
}

// Dir returns the changeset directory path.
func (cr *ChangesetFiles) Dir() string {
	return cr.dir
}

// TreeDir returns the parent tree directory path.
func (cr *ChangesetFiles) TreeDir() string {
	return cr.treeDir
}

// WALFile returns the wal.log file handle.
func (cr *ChangesetFiles) WALFile() *os.File {
	return cr.walFile
}

// KVDataFile returns the kv.dat file handle.
func (cr *ChangesetFiles) KVDataFile() *os.File {
	return cr.kvDataFile
}

// BranchesFile returns the branches.dat file handle.
func (cr *ChangesetFiles) BranchesFile() *os.File {
	return cr.branchesFile
}

// LeavesFile returns the leaves.dat file handle.
func (cr *ChangesetFiles) LeavesFile() *os.File {
	return cr.leavesFile
}

// CheckpointsFile returns the checkpoints.dat file handle.
func (cr *ChangesetFiles) CheckpointsFile() *os.File {
	return cr.checkpointsFile
}

// OrphansFile returns the orphans.dat file handle.
func (cr *ChangesetFiles) OrphansFile() *os.File {
	return cr.orphansFile
}

// StartVersion returns the start version of the changeset (directory name).
// This could be WAL start or checkpoint start depending on compaction state.
func (cr *ChangesetFiles) StartVersion() uint32 {
	return cr.startVersion
}

// CompactedAtVersion returns the compacted at version of the changeset.
// If the changeset is original and uncompacted, this will be 0.
func (cr *ChangesetFiles) CompactedAtVersion() uint32 {
	return cr.compactedAt
}

// EndVersion returns the end version of the changeset if it is known, or 0.
// For original changesets, this will always return 0.
// For sealed, compacted changesets, this will return the accurate end version of the changeset.
func (cr *ChangesetFiles) EndVersion() uint32 {
	return cr.endVersion
}

func (cr *ChangesetFiles) MarkReadyAndClose(endVersion uint32) (finalDir string, err error) {
	tmpDir := cr.dir
	if !strings.HasSuffix(tmpDir, "-tmp") {
		return "", fmt.Errorf("cannot mark changeset as ready: directory name does not have -tmp suffix: %s", tmpDir)
	}
	finalDir = filepath.Join(cr.treeDir, fmt.Sprintf("%d-%d.%d", cr.startVersion, endVersion, cr.compactedAt))
	err = os.Rename(tmpDir, finalDir)
	if err != nil {
		return "", fmt.Errorf("failed to rename changeset directory from %s to %s: %w", tmpDir, finalDir, err)
	}
	err = cr.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close changeset files: %w", err)
	}
	cr.dir = finalDir
	cr.endVersion = endVersion
	return finalDir, nil
}

// Close closes all changeset files.
func (cr *ChangesetFiles) Close() error {
	if cr.closed {
		return nil
	}

	cr.closed = true
	err := errors.Join(
		cr.walFile.Close(),
		cr.kvDataFile.Close(),
		cr.branchesFile.Close(),
		cr.leavesFile.Close(),
		cr.checkpointsFile.Close(),
		cr.orphansFile.Close(),
	)
	return err
}

// DeleteFiles deletes all changeset files and the changeset directory.
// If the files were not already closed, they will be closed first.
func (cr *ChangesetFiles) DeleteFiles() error {
	return errors.Join(
		cr.Close(),           // first close all files
		os.RemoveAll(cr.dir), // then delete the directory and all files within it
	)
}
