package iavlx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ChangesetFiles struct {
	dir          string
	treeDir      string
	startVersion uint32
	compactedAt  uint32

	kvlogFile    *os.File
	kvlogPath    string
	branchesFile *os.File
	leavesFile   *os.File
	versionsFile *os.File
	orphansFile  *os.File
	infoFile     *os.File
	info         *ChangesetInfo

	closed bool
}

func CreateChangesetFiles(treeDir string, startVersion, compactedAt uint32, kvlogPath string) (*ChangesetFiles, error) {
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
		err := os.WriteFile(filepath.Join(dir, "pending"), []byte{}, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to create pending marker file for compacted changeset: %w", err)
		}
	}

	localKVLogPath := filepath.Join(dir, "kv.log")
	if kvlogPath == "" {
		// For original (non-compacted) changesets, normalize the path by evaluating
		// symlinks in the directory path to ensure consistent comparisons later.
		// This handles platform differences like /var vs /private/var on macOS.
		normalizedDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to eval directory path: %w", err)
		}
		kvlogPath = filepath.Join(normalizedDir, "kv.log")
	} else {
		// create symlink to kvlog so that it can be reopened later
		err := os.Symlink(kvlogPath, localKVLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create kvlog symlink: %w", err)
		}
		kvlogPath, err = filepath.EvalSymlinks(localKVLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to eval kvlog symlink: %w", err)
		}
	}

	cr := &ChangesetFiles{
		dir:          dir,
		treeDir:      treeDir,
		startVersion: startVersion,
		compactedAt:  compactedAt,
		kvlogPath:    kvlogPath,
	}

	err = cr.open(os.O_RDWR | os.O_CREATE | os.O_APPEND)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}
	return cr, nil
}

func OpenChangesetFiles(dirName string) (*ChangesetFiles, error) {
	startVersion, compactedAt, valid := ParseChangesetDirName(filepath.Base(dirName))
	if !valid {
		return nil, fmt.Errorf("invalid changeset dir name: %s", dirName)
	}

	dir, err := filepath.Abs(dirName)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", dirName, err)
	}

	treeDir := filepath.Dir(dir)

	localKVLogPath := filepath.Join(dir, "kv.log")
	kvlogPath, err := filepath.EvalSymlinks(localKVLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to eval kvlog symlink: %w", err)
	}

	cr := &ChangesetFiles{
		dir:          dir,
		treeDir:      treeDir,
		startVersion: uint32(startVersion),
		compactedAt:  uint32(compactedAt),
		kvlogPath:    kvlogPath,
	}

	err = cr.open(os.O_RDWR)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	return cr, nil
}

func (cr *ChangesetFiles) open(mode int) error {
	var err error
	leavesPath := filepath.Join(cr.dir, "leaves.dat")

	cr.kvlogFile, err = os.OpenFile(cr.kvlogPath, mode, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create KV log file: %w", err)
	}

	cr.leavesFile, err = os.OpenFile(leavesPath, mode, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create leaves data file: %w", err)
	}

	branchesPath := filepath.Join(cr.dir, "branches.dat")
	cr.branchesFile, err = os.OpenFile(branchesPath, mode, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create branches data file: %w", err)
	}

	versionsPath := filepath.Join(cr.dir, "versions.dat")
	cr.versionsFile, err = os.OpenFile(versionsPath, mode, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create versions data file: %w", err)
	}

	orphansPath := filepath.Join(cr.dir, "orphans.dat")
	cr.orphansFile, err = os.OpenFile(orphansPath, mode, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create orphans data file: %w", err)
	}

	infoPath := filepath.Join(cr.dir, "info.dat")
	cr.infoFile, err = os.OpenFile(infoPath, mode, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create changeset info file: %w", err)
	}

	cr.info, err = ReadChangesetInfo(cr.infoFile)
	if err != nil {
		return fmt.Errorf("failed to read changeset info: %w", err)
	}

	return nil
}

func ParseChangesetDirName(dirName string) (startVersion, compactedAt uint64, valid bool) {
	var err error
	// if no dot, it's an original changeset
	if !strings.Contains(dirName, ".") {
		startVersion, err = strconv.ParseUint(dirName, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		return startVersion, 0, true
	} else {
		parts := strings.Split(dirName, ".")
		if len(parts) != 2 {
			return 0, 0, false
		}
		startVersion, err = strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return 0, 0, false
		}
		compactedAt, err = strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return 0, 0, false
		}
		return startVersion, compactedAt, true
	}
}

func (cr *ChangesetFiles) TreeDir() string {
	return cr.treeDir
}

func (cr *ChangesetFiles) KVLogPath() string {
	return cr.kvlogFile.Name()
}

func (cr *ChangesetFiles) StartVersion() uint32 {
	return cr.startVersion
}

func (cr *ChangesetFiles) CompactedAtVersion() uint32 {
	return cr.compactedAt
}

func (cr *ChangesetFiles) Info() *ChangesetInfo {
	return cr.info
}

func (cr *ChangesetFiles) RewriteInfo() error {
	return RewriteChangesetInfo(cr.infoFile, cr.info)
}

func IsChangesetReady(dir string) (bool, error) {
	pendingPath := filepath.Join(dir, "pending")
	_, err := os.Stat(pendingPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, fmt.Errorf("failed to stat pending marker file: %w", err)
	}
	return false, nil
}

func (cr *ChangesetFiles) MarkReady() error {
	pendingPath := filepath.Join(cr.dir, "pending")
	err := os.Remove(pendingPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove pending marker file: %w", err)
	}
	return nil
}

type ChangesetDeleteArgs struct {
	SaveKVLogPath string
}

func (cr *ChangesetFiles) Close() error {
	if cr.closed {
		return nil
	}

	cr.closed = true
	err := errors.Join(
		cr.RewriteInfo(),
		cr.kvlogFile.Close(),
		cr.branchesFile.Close(),
		cr.leavesFile.Close(),
		cr.versionsFile.Close(),
		cr.orphansFile.Close(),
		cr.infoFile.Close(),
	)
	cr.info = nil
	return err
}

func (cr *ChangesetFiles) DeleteFiles(args ChangesetDeleteArgs) error {
	errs := []error{
		os.Remove(cr.infoFile.Name()),
		os.Remove(cr.leavesFile.Name()),
		os.Remove(cr.branchesFile.Name()),
		os.Remove(cr.versionsFile.Name()),
		os.Remove(cr.orphansFile.Name()),
	}

	localKVLogPath := filepath.Join(cr.dir, "kv.log")
	if cr.kvlogPath != args.SaveKVLogPath {
		// delete the local kv.log (which might be a symlink)
		errs = append(errs, os.Remove(localKVLogPath))
	}
	err := errors.Join(errs...)
	if err != nil {
		return fmt.Errorf("failed to delete changeset files: %w", err)
	}
	// delete dir if empty
	_ = os.Remove(cr.dir)
	return nil
}
