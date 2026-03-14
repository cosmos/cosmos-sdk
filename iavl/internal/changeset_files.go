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
//
// A changeset contains the following files:
// - wal.log: the WAL entries for all versions in the changeset
// - kv.dat: key-value data which isn't in the WAL
// - checkpoints.dat: checkpoint metadata
// - branches.dat: the actual BranchLayout structs for checkpoints
// - leaves.dat: the actual LeafLayout structs for checkpoints
// - orphans.dat: a log of orphan node IDs and versions they were orphaned at for pruning
//
// Changeset are identified by their directory in two formats:
// - original, uncompacted changesets: {startVersion}/
// - compacted changesets: {startVersion}-{endVersion}.{compactedAt}/
type ChangesetFiles struct {
	dir          string
	treeDir      string
	startVersion uint32
	endVersion   uint32 // 0 if original changeset
	compactedAt  uint32 // 0 if original changeset

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
//
// If the directory already exists, it will be used if and only if all the changeset files within it are empty,
// otherwise an error will be returned to avoid data corruption.
// Sometimes during normal operation, a new changeset is created but not written to, this is okay,
// and we want to continue using this directory, but only if it is truly unused (i.e. all files are empty).
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

	var existingDir bool
	if err := os.Mkdir(dir, 0o755); err != nil {
		if errors.Is(err, os.ErrExist) {
			existingDir = true
		} else {
			return nil, fmt.Errorf("failed to create changeset dir: %w", err)
		}
	}

	cr := &ChangesetFiles{
		dir:          dir,
		treeDir:      treeDir,
		startVersion: startVersion,
		compactedAt:  compactedAt,
	}

	err = cr.open(writeModeFlags)
	if err != nil {
		cr.Close() // close any files that were opened before returning error
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	if existingDir {
		// an existing directory is okay if and only if all the files are empty,
		// otherwise we risk data loss by writing to an existing changeset
		for _, f := range cr.allFiles() {
			info, err := f.Stat()
			if err == nil && info.Size() > 0 {
				cr.Close() // close files before returning error
				return nil, fmt.Errorf("changeset dir already exists and is not empty: %s", dir)
			}
			if err != nil {
				cr.Close() // close files before returning error
				return nil, fmt.Errorf("failed to stat file %s: %w", f.Name(), err)
			}
		}
	}

	return cr, nil
}

const writeModeFlags = os.O_RDWR | os.O_CREATE | os.O_APPEND

// OpenChangesetFiles opens an existing changeset directory and files.
// All files are opened in readonly mode, except for orphans.dat which is opened in read-write mode
// to track orphan data.
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
		cr.Close() // close any files that were opened before returning error
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	return cr, nil
}

// open opens all changeset files with the specified mode, except for orphans.dat which is always opened with write
// permissions to allow tracking of orphan data.
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
//
// Valid directory name formats are:
// - original, uncompacted changesets: {startVersion}/
// - compacted changesets: {startVersion}-{endVersion}.{compactedAt}/
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
// For original changesets, this will always return 0 and the end version must be determined by
// reading to the end of the WAL.
// For sealed, compacted changesets, this will always return the accurate end version of the changeset.
func (cr *ChangesetFiles) EndVersion() uint32 {
	return cr.endVersion
}

// MarkReadyAndClose marks a compacted changeset as ready by renaming the directory to remove the -tmp suffix,
// and then closes all files.
// It is an error to call this method on an original, uncompacted changeset (i.e. one without the -tmp suffix)
// since these changesets are already considered ready.
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
	cr.dir = finalDir
	cr.endVersion = endVersion
	err = cr.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close changeset files: %w", err)
	}
	return finalDir, nil
}

func (cr *ChangesetFiles) allFiles() []*os.File {
	return []*os.File{
		cr.walFile,
		cr.kvDataFile,
		cr.leavesFile,
		cr.branchesFile,
		cr.checkpointsFile,
		cr.orphansFile,
	}
}

// Close closes all changeset files.
// Calls to Close are idempotent and will not return an error if the files are already closed.
func (cr *ChangesetFiles) Close() error {
	if cr.closed {
		return nil
	}

	cr.closed = true
	var errs []error
	for _, f := range cr.allFiles() {
		if f != nil {
			err := f.Close()
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to close file %s: %w", f.Name(), err))
			}
		}
	}
	return errors.Join(errs...)
}

// DeleteFiles deletes all changeset files and the changeset directory.
// If the files were not already closed, they will be closed first.
func (cr *ChangesetFiles) DeleteFiles() error {
	return errors.Join(
		cr.Close(),           // first close all files
		os.RemoveAll(cr.dir), // then delete the directory and all files within it
	)
}
