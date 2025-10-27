package iavlx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ChangesetFiles struct {
	dir          string
	treeDir      string
	startVersion uint32
	compactedAt  uint32

	kvlogFile    *os.File
	branchesFile *os.File
	leavesFile   *os.File
	versionsFile *os.File
	infoFile     *os.File

	info     *ChangesetInfo
	infoMmap *StructMmap[ChangesetInfo]

	closed bool
}

func OpenChangesetFiles(treeDir string, startVersion, compactedAt uint32, kvlogPath string) (*ChangesetFiles, error) {
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

	if kvlogPath == "" {
		kvlogPath = filepath.Join(dir, "kv.log")
	}
	kvlogFile, err := os.OpenFile(kvlogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create KV log file: %w", err)
	}

	leavesPath := filepath.Join(dir, "leaves.dat")
	leavesFile, err := os.OpenFile(leavesPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create leaves data file: %w", err)
	}

	branchesPath := filepath.Join(dir, "branches.dat")
	branchesFile, err := os.OpenFile(branchesPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create branches data file: %w", err)
	}

	versionsPath := filepath.Join(dir, "versions.dat")
	versionsFile, err := os.OpenFile(versionsPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create versions data file: %w", err)
	}

	infoPath := filepath.Join(dir, "info.dat")
	infoFile, err := os.OpenFile(infoPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create changeset info file: %w", err)
	}

	// check file size to see if we need to initialize
	stat, err := infoFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat info file: %w", err)
	}

	if stat.Size() == 0 {
		// file is empty, initialize it
		infoWriter := NewStructWriter[ChangesetInfo](infoFile)
		if err := infoWriter.Append(&ChangesetInfo{}); err != nil {
			return nil, fmt.Errorf("failed to write initial changeset info: %w", err)
		}
		if err := infoWriter.Flush(); err != nil {
			return nil, fmt.Errorf("failed to flush initial changeset info: %w", err)
		}
	}

	// now create the mmap reader
	infoMmap, err := NewStructReader[ChangesetInfo](infoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset info: %w", err)
	}

	if infoMmap.Count() != 1 {
		return nil, fmt.Errorf("changeset info file has unexpected item count: %d", infoMmap.Count())
	}

	return &ChangesetFiles{
		dir:          dir,
		treeDir:      treeDir,
		startVersion: startVersion,
		compactedAt:  compactedAt,
		kvlogFile:    kvlogFile,
		branchesFile: branchesFile,
		leavesFile:   leavesFile,
		versionsFile: versionsFile,
		infoFile:     infoFile,
		info:         infoMmap.UnsafeItem(0),
		infoMmap:     infoMmap,
	}, nil
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

type ChangesetDeleteArgs struct {
	SaveKVLogPath string
}

func (cr *ChangesetFiles) Close() error {
	if cr.closed {
		return nil
	}

	cr.closed = true
	cr.info = nil
	return errors.Join(
		cr.kvlogFile.Close(),
		cr.branchesFile.Close(),
		cr.leavesFile.Close(),
		cr.versionsFile.Close(),
		cr.infoFile.Close(),
		cr.infoMmap.Close(),
	)
}

func (cr *ChangesetFiles) DeleteFiles(args ChangesetDeleteArgs) error {
	errs := []error{
		os.Remove(cr.infoFile.Name()),
		os.Remove(cr.leavesFile.Name()),
		os.Remove(cr.branchesFile.Name()),
		os.Remove(cr.versionsFile.Name()),
	}
	if cr.kvlogFile.Name() != args.SaveKVLogPath {
		errs = append(errs, os.Remove(cr.kvlogFile.Name()))
	}
	err := errors.Join(errs...)
	if err != nil {
		return fmt.Errorf("failed to delete changeset files: %w", err)
	}
	// delete dir if empty
	_ = os.Remove(cr.dir)
	return nil
}
