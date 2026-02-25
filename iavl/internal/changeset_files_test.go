package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseChangesetDirName(t *testing.T) {
	tests := []struct {
		name            string
		dirName         string
		wantStart       uint32
		wantEnd         uint32
		wantCompactedAt uint32
		wantValid       bool
	}{
		{
			name:            "uncompacted",
			dirName:         "100",
			wantStart:       100,
			wantEnd:         0,
			wantCompactedAt: 0,
			wantValid:       true,
		},
		{
			name:            "compacted",
			dirName:         "100-200.300",
			wantStart:       100,
			wantEnd:         200,
			wantCompactedAt: 300,
			wantValid:       true,
		},
		{
			name:      "invalid - not a number",
			dirName:   "abc",
			wantValid: false,
		},
		{
			name:      "invalid - too many dots",
			dirName:   "1.2.3",
			wantValid: false,
		},
		{
			name:      "invalid - empty",
			dirName:   "",
			wantValid: false,
		},
		{
			name:      "invalid - overflow",
			dirName:   "5000000000",
			wantValid: false,
		},
		{
			name:      "zero version",
			dirName:   "0",
			wantStart: 0,
			wantValid: true,
		},
		{
			name:      "max uint32",
			dirName:   "4294967295",
			wantStart: 4294967295,
			wantValid: true,
		},
		{
			name:      "one-past max uint32",
			dirName:   "4294967296",
			wantValid: false,
		},
		{
			name:      "dash but no dot",
			dirName:   "100-200",
			wantValid: false,
		},
		{
			name:      "invalid startVersion in compacted",
			dirName:   "abc-200.300",
			wantValid: false,
		},
		{
			name:      "invalid endVersion in compacted",
			dirName:   "100-abc.300",
			wantValid: false,
		},
		{
			name:      "invalid compactedAt",
			dirName:   "100-200.abc",
			wantValid: false,
		},
		{
			name:      "empty startVersion",
			dirName:   "-200.300",
			wantValid: false,
		},
		{
			name:      "empty endVersion",
			dirName:   "100-.300",
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, compactedAt, valid := ParseChangesetDirName(tt.dirName)
			require.Equal(t, tt.wantValid, valid)
			if valid {
				require.Equal(t, tt.wantStart, start)
				require.Equal(t, tt.wantEnd, end)
				require.Equal(t, tt.wantCompactedAt, compactedAt)
			}
		})
	}
}

func TestCreateChangesetFiles_Uncompacted(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)
	defer cf.Close()

	// Check directory created
	require.Equal(t, filepath.Join(treeDir, "100"), cf.Dir())
	require.Equal(t, uint32(100), cf.StartVersion())
	require.Equal(t, uint32(0), cf.CompactedAtVersion())
	require.Equal(t, uint32(0), cf.EndVersion())

	// All files should exist
	require.NotNil(t, cf.KVDataFile())
	require.NotNil(t, cf.WALFile())
	require.NotNil(t, cf.BranchesFile())
	require.NotNil(t, cf.LeavesFile())
	require.NotNil(t, cf.CheckpointsFile())
	require.NotNil(t, cf.OrphansFile())

	require.Equal(t, treeDir, cf.TreeDir())
}

func TestCreateChangesetFiles_Compacted(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 200)
	require.NoError(t, err)
	defer cf.Close()

	require.Contains(t, cf.Dir(), "-tmp")

	finalDir, err := cf.MarkReadyAndClose(150)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(treeDir, "100-150.200"), finalDir)
}

func TestChangesetFiles_OpenExisting(t *testing.T) {
	treeDir := t.TempDir()

	// Create and close
	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)
	require.NoError(t, cf.Close())

	// Reopen
	cf2, err := OpenChangesetFiles(filepath.Join(treeDir, "100"))
	require.NoError(t, err)
	defer cf2.Close()

	require.Equal(t, uint32(100), cf2.StartVersion())
	require.Equal(t, uint32(0), cf2.CompactedAtVersion())
}

func TestChangesetFiles_DeleteFiles(t *testing.T) {
	tests := []struct {
		name         string
		startVersion uint32
		compactedAt  uint32
	}{
		{
			name:         "uncompacted",
			startVersion: 100,
		},
		{
			name:         "compacted",
			startVersion: 5,
			compactedAt:  10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			treeDir := t.TempDir()

			cf, err := CreateChangesetFiles(treeDir, tt.startVersion, tt.compactedAt)
			require.NoError(t, err)

			dir := cf.Dir()

			// Directory should exist
			_, err = os.Stat(dir)
			require.NoError(t, err)

			// Delete
			err = cf.DeleteFiles()
			require.NoError(t, err)

			// Directory should be gone
			_, err = os.Stat(dir)
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestChangesetFiles_CloseIdempotent(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)

	// Close multiple times should not error
	require.NoError(t, cf.Close())
	require.NoError(t, cf.Close())
}

func TestOpenChangesetFiles_InvalidDir(t *testing.T) {
	treeDir := t.TempDir()

	// Create a directory with invalid name
	invalidDir := filepath.Join(treeDir, "not-a-changeset")
	require.NoError(t, os.MkdirAll(invalidDir, 0o755))

	_, err := OpenChangesetFiles(invalidDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid changeset dir name")
}

func TestCreateChangesetFiles_ExistingNonEmptyDir(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)

	// Write a byte to the WAL file to make the directory non-empty
	_, err = cf.WALFile().Write([]byte{0x01})
	require.NoError(t, err)
	require.NoError(t, cf.Close())

	// Second creation attempt must fail
	_, err = CreateChangesetFiles(treeDir, 100, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists and is not empty")
}

func TestCreateChangesetFiles_ExistingEmptyDir(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)
	require.NoError(t, cf.Close())

	// All files are empty — re-opening should succeed
	cf2, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)
	defer cf2.Close()
}

func TestOpenChangesetFiles_CompactedDir(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 200)
	require.NoError(t, err)

	finalDir, err := cf.MarkReadyAndClose(150)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(treeDir, "100-150.200"), finalDir)

	cf2, err := OpenChangesetFiles(finalDir)
	require.NoError(t, err)
	defer cf2.Close()

	require.Equal(t, uint32(100), cf2.StartVersion())
	require.Equal(t, uint32(150), cf2.EndVersion())
	require.Equal(t, uint32(200), cf2.CompactedAtVersion())
	require.NotNil(t, cf2.WALFile())
}

func TestOpenChangesetFiles_NonExistent(t *testing.T) {
	treeDir := t.TempDir()

	// Valid name, but the directory was never created
	_, err := OpenChangesetFiles(filepath.Join(treeDir, "999"))
	require.Error(t, err)
}

func TestMarkReadyAndClose_OnUncompacted(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)
	defer cf.Close()

	// Uncompacted dirs have no -tmp suffix; MarkReadyAndClose must reject them
	_, err = cf.MarkReadyAndClose(150)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not have -tmp suffix")
}

func TestDeleteFiles_AfterClose(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)

	dir := cf.Dir()

	require.NoError(t, cf.Close())
	require.NoError(t, cf.DeleteFiles())

	_, err = os.Stat(dir)
	require.True(t, os.IsNotExist(err))
}
