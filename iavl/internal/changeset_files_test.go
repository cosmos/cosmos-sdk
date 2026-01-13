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
		wantCompactedAt uint32
		wantValid       bool
	}{
		{
			name:            "uncompacted",
			dirName:         "100",
			wantStart:       100,
			wantCompactedAt: 0,
			wantValid:       true,
		},
		{
			name:            "compacted",
			dirName:         "100.200",
			wantStart:       100,
			wantCompactedAt: 200,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, compactedAt, valid := ParseChangesetDirName(tt.dirName)
			require.Equal(t, tt.wantValid, valid)
			if valid {
				require.Equal(t, tt.wantStart, start)
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
	require.Equal(t, uint32(100), cf.StartLayer())
	require.Equal(t, uint32(0), cf.CompactedAtVersion())

	// Uncompacted should be ready immediately (no pending marker)
	ready, err := IsChangesetReady(cf.Dir())
	require.NoError(t, err)
	require.True(t, ready)

	// All files should exist
	require.NotNil(t, cf.KVDataFile())
	require.NotNil(t, cf.BranchesFile())
	require.NotNil(t, cf.LeavesFile())
	require.NotNil(t, cf.LayersFile())
	require.NotNil(t, cf.OrphansFile())
	require.NotNil(t, cf.Info())
}

func TestCreateChangesetFiles_Compacted(t *testing.T) {
	treeDir := t.TempDir()

	cf, err := CreateChangesetFiles(treeDir, 100, 200)
	require.NoError(t, err)
	defer cf.Close()

	// Check directory name includes compactedAt
	require.Equal(t, filepath.Join(treeDir, "100.200"), cf.Dir())
	require.Equal(t, uint32(100), cf.StartLayer())
	require.Equal(t, uint32(200), cf.CompactedAtVersion())

	// Compacted should NOT be ready until MarkReady called
	ready, err := IsChangesetReady(cf.Dir())
	require.NoError(t, err)
	require.False(t, ready)

	// Mark ready
	err = cf.MarkReady()
	require.NoError(t, err)

	// Now should be ready
	ready, err = IsChangesetReady(cf.Dir())
	require.NoError(t, err)
	require.True(t, ready)
}

func TestChangesetFiles_OpenExisting(t *testing.T) {
	treeDir := t.TempDir()

	// Create and close
	cf, err := CreateChangesetFiles(treeDir, 100, 0)
	require.NoError(t, err)

	// Modify info
	cf.Info().StartVersion = 100
	cf.Info().EndVersion = 150
	cf.Info().LeafOrphans = 42

	// Persist info
	require.NoError(t, cf.RewriteInfo())
	require.NoError(t, cf.Close())

	// Reopen
	cf2, err := OpenChangesetFiles(filepath.Join(treeDir, "100"))
	require.NoError(t, err)
	defer cf2.Close()

	require.Equal(t, uint32(100), cf2.StartLayer())
	require.Equal(t, uint32(0), cf2.CompactedAtVersion())

	// Info should be persisted
	require.Equal(t, uint32(100), cf2.Info().StartVersion)
	require.Equal(t, uint32(150), cf2.Info().EndVersion)
	require.Equal(t, uint32(42), cf2.Info().LeafOrphans)
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
	invalidDir := filepath.Join(treeDir, "not-a-layer")
	require.NoError(t, os.MkdirAll(invalidDir, 0o755))

	_, err := OpenChangesetFiles(invalidDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid changeset dir name")
}
