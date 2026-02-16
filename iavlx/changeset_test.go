package iavlx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangeset_TryDisposeClearsOrphanWriter(t *testing.T) {
	dir := t.TempDir()

	writer, err := NewChangesetWriter(dir, 1, nil)
	require.NoError(t, err)

	reader, err := writer.CreatedSharedReader()
	require.NoError(t, err)
	require.NotNil(t, reader.orphanWriter)

	reader.Evict()
	require.True(t, reader.TryDispose())
	require.Nil(t, reader.orphanWriter)
}
