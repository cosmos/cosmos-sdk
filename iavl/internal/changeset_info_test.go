package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangesetInfo_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.dat")

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	require.NoError(t, err)
	defer file.Close()

	original := &ChangesetInfo{
		WALStartVersion:          100,
		WALEndVersion:            200,
		FirstCheckpoint:          100,
		LastCheckpoint:           200,
		LeafOrphans:              50,
		BranchOrphans:            25,
		LeafOrphanVersionTotal:   5000,
		BranchOrphanVersionTotal: 2500,
	}

	err = RewriteChangesetInfo(file, original)
	require.NoError(t, err)

	read, err := ReadChangesetInfo(file)
	require.NoError(t, err)

	require.Equal(t, original.WALStartVersion, read.WALStartVersion)
	require.Equal(t, original.WALEndVersion, read.WALEndVersion)
	require.Equal(t, original.FirstCheckpoint, read.FirstCheckpoint)
	require.Equal(t, original.LastCheckpoint, read.LastCheckpoint)
	require.Equal(t, original.LeafOrphans, read.LeafOrphans)
	require.Equal(t, original.BranchOrphans, read.BranchOrphans)
	require.Equal(t, original.LeafOrphanVersionTotal, read.LeafOrphanVersionTotal)
	require.Equal(t, original.BranchOrphanVersionTotal, read.BranchOrphanVersionTotal)
}

func TestChangesetInfo_RewriteOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.dat")

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	require.NoError(t, err)
	defer file.Close()

	// Write first version
	v1 := &ChangesetInfo{WALStartVersion: 1, WALEndVersion: 1, FirstCheckpoint: 1, LastCheckpoint: 1}
	err = RewriteChangesetInfo(file, v1)
	require.NoError(t, err)

	// Overwrite with second version
	v2 := &ChangesetInfo{WALStartVersion: 1, WALEndVersion: 2, FirstCheckpoint: 1, LastCheckpoint: 2}
	err = RewriteChangesetInfo(file, v2)
	require.NoError(t, err)

	// Read should return second version
	read, err := ReadChangesetInfo(file)
	require.NoError(t, err)
	require.Equal(t, uint32(1), read.WALStartVersion)
	require.Equal(t, uint32(2), read.WALEndVersion)
	require.Equal(t, uint32(1), read.FirstCheckpoint)
	require.Equal(t, uint32(2), read.LastCheckpoint)
}

func TestChangesetInfo_ReadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.dat")

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	require.NoError(t, err)
	defer file.Close()

	// Reading empty file should return default struct
	info, err := ReadChangesetInfo(file)
	require.NoError(t, err)
	require.Equal(t, &ChangesetInfo{}, info)
}

func TestChangesetInfo_ReadWrongSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.dat")

	// Write garbage data of wrong size
	err := os.WriteFile(path, []byte("short"), 0o600)
	require.NoError(t, err)

	file, err := os.OpenFile(path, os.O_RDONLY, 0o600)
	require.NoError(t, err)
	defer file.Close()

	_, err = ReadChangesetInfo(file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected size")
}
