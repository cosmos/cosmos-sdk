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
	compactedAt  uint32

	walFile      *os.File
	kvDataFile   *os.File
	branchesFile *os.File
	leavesFile   *os.File
	layerFiles   *os.File
	orphansFile  *os.File
	infoFile     *os.File
	info         *ChangesetInfo

	closed bool
}

// CreateChangesetFiles creates a new changeset directory and files that are ready to be written to.
// If compactedAt is 0, the changeset is considered original and uncompacted.
// If compactedAt is greater than 0, the changeset is considered compacted and a pending marker file
// will be created to indicate that the changeset is not yet ready for use.
// This pending marker file must be removed once the compaction is fully complete by calling MarkReady,
// otherwise the changeset will be considered incomplete and deleted at the next startup.
func CreateChangesetFiles(treeDir string, startVersion, compactedAt uint32, haveWal bool) (*ChangesetFiles, error) {
	// ensure absolute path
	var err error
	treeDir, err = filepath.Abs(treeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", treeDir, err)
	}

	dirName := fmt.Sprintf("%d", startVersion)
	if compactedAt > 0 {
		dirName = fmt.Sprintf("%d.%d", startVersion, compactedAt)
	}
	dir := filepath.Join(treeDir, dirName)

	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		return nil, fmt.Errorf("failed to create changeset dir: %w", err)
	}

	// create pending marker file for compacted changesets
	if compactedAt > 0 {
		err := os.WriteFile(pendingFilename(dir), []byte{}, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to create pending marker file for compacted changeset: %w", err)
		}
	}

	cr := &ChangesetFiles{
		dir:          dir,
		treeDir:      treeDir,
		startVersion: startVersion,
		compactedAt:  compactedAt,
	}

	err = cr.open(writeModeFlags, haveWal)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	// set start version in info file
	cr.info.StartVersion = startVersion

	return cr, nil
}

const writeModeFlags = os.O_RDWR | os.O_CREATE | os.O_APPEND

// OpenChangesetFiles opens an existing changeset directory and files.
// All files are opened in readonly mode, except for orphans.dat and info.dat which are opened in read-write mode
// to track orphan data and statistics.
func OpenChangesetFiles(dirName string) (*ChangesetFiles, error) {
	startLayer, compactedAt, valid := ParseChangesetDirName(filepath.Base(dirName))
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
		startVersion: startLayer,
		compactedAt:  compactedAt,
	}

	haveWal := false
	walPath := filepath.Join(cr.dir, "wal.log")
	_, err = os.Stat(walPath)
	if err == nil {
		haveWal = true
	}

	err = cr.open(os.O_RDONLY, haveWal)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	return cr, nil
}

func (cr *ChangesetFiles) open(mode int, haveWal bool) error {
	var err error

	if haveWal {
		walPath := filepath.Join(cr.dir, "wal.log")
		cr.walFile, err = os.OpenFile(walPath, mode, 0o600)
		if err != nil {
			return fmt.Errorf("failed to open WAL data file: %w", err)
		}
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

	layersPath := filepath.Join(cr.dir, "layers.dat")
	cr.layerFiles, err = os.OpenFile(layersPath, mode, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open versions data file: %w", err)
	}

	orphansPath := filepath.Join(cr.dir, "orphans.dat")
	cr.orphansFile, err = os.OpenFile(orphansPath, writeModeFlags, 0o600) // the orphans file is always opened for writing
	if err != nil {
		return fmt.Errorf("failed to open orphans data file: %w", err)
	}

	infoPath := filepath.Join(cr.dir, "info.dat")
	cr.infoFile, err = os.OpenFile(infoPath, os.O_RDWR|os.O_CREATE, 0o600) // info file uses random access, not append
	if err != nil {
		return fmt.Errorf("failed to open changeset info file: %w", err)
	}

	cr.info, err = ReadChangesetInfo(cr.infoFile)
	if err != nil {
		return fmt.Errorf("failed to read changeset info: %w", err)
	}

	return nil
}

// ParseChangesetDirName parses a changeset directory name and returns the start version and compacted at version.
// If the directory name is invalid, valid will be false.
// If a changeset is original and uncompacted, compactedAt will be 0.
func ParseChangesetDirName(dirName string) (startLayer, compactedAt uint32, valid bool) {
	var err error
	var v uint64
	// if no dot, it's an original changeset
	if !strings.Contains(dirName, ".") {
		v, err = strconv.ParseUint(dirName, 10, 32)
		if err != nil {
			return 0, 0, false
		}
		return uint32(v), 0, true
	}

	parts := strings.Split(dirName, ".")
	if len(parts) != 2 {
		return 0, 0, false
	}

	v, err = strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, 0, false
	}
	startLayer = uint32(v)

	v, err = strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, 0, false
	}
	compactedAt = uint32(v)

	return startLayer, compactedAt, true
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

// LayersFile returns the layers.dat file handle.
func (cr *ChangesetFiles) LayersFile() *os.File {
	return cr.layerFiles
}

// OrphansFile returns the orphans.dat file handle.
func (cr *ChangesetFiles) OrphansFile() *os.File {
	return cr.orphansFile
}

// Info returns the changeset info struct.
// This struct is writeable and changes will be persisted to disk when RewriteInfo is called.
func (cr *ChangesetFiles) Info() *ChangesetInfo {
	return cr.info
}

// RewriteInfo rewrites the changeset info file with the current info struct.
func (cr *ChangesetFiles) RewriteInfo() error {
	return RewriteChangesetInfo(cr.infoFile, cr.info)
}

// StartVersion returns the start version of the changeset.
func (cr *ChangesetFiles) StartVersion() uint32 {
	return cr.startVersion
}

// CompactedAtVersion returns the compacted at version of the changeset.
// If the changeset is original and uncompacted, this will be 0.
func (cr *ChangesetFiles) CompactedAtVersion() uint32 {
	return cr.compactedAt
}

func pendingFilename(dir string) string {
	return filepath.Join(dir, "pending")
}

// IsChangesetReady checks if the changeset is ready to be used.
// A changeset is considered ready if the pending marker file does not exist.
// This is used by startup code only to detect whether a changeset compaction was interrupted before it could complete.
func IsChangesetReady(dir string) (bool, error) {
	pendingPath := pendingFilename(dir)
	_, err := os.Stat(pendingPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, fmt.Errorf("failed to stat pending marker file: %w", err)
	}
	return false, nil
}

// MarkReady marks the changeset as ready by removing the pending marker file.
// This is only necessary for compacted changesets.
func (cr *ChangesetFiles) MarkReady() error {
	pendingPath := pendingFilename(cr.dir)
	err := os.Remove(pendingPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove pending marker file: %w", err)
	}
	return nil
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
		cr.layerFiles.Close(),
		cr.orphansFile.Close(),
		cr.infoFile.Close(),
	)
	cr.info = nil
	return err
}

// DeleteFiles deletes all changeset files and the changeset directory.
// If the files were not already closed, they will be closed first.
func (cr *ChangesetFiles) DeleteFiles() error {
	return errors.Join(
		cr.Close(), // first close all files
		os.Remove(cr.walFile.Name()),
		os.Remove(cr.infoFile.Name()),
		os.Remove(cr.leavesFile.Name()),
		os.Remove(cr.branchesFile.Name()),
		os.Remove(cr.layerFiles.Name()),
		os.Remove(cr.orphansFile.Name()),
		os.Remove(cr.kvDataFile.Name()),
		cr.MarkReady(), // remove pending marker file if it exists
		os.Remove(cr.dir),
	)
}
