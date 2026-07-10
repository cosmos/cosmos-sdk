package snapshots

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	snapshottypes "github.com/cosmos/cosmos-sdk/store/v2/snapshots/types"
)

// TestChunkBoundedReader exercises chunkBoundedReader directly (white-box) rather than through
// NewStreamReader, since the production constants (10MB chunks, 1GB budget) are too large to
// drive through a fast unit test. chunkReader is only consulted for its opened-chunk count, so
// it can be advanced independently of the bytes chunkBoundedReader actually reads.
func TestChunkBoundedReader(t *testing.T) {
	ch := make(chan io.ReadCloser, 2)
	ch <- io.NopCloser(bytes.NewReader([]byte{1}))
	ch <- io.NopCloser(bytes.NewReader([]byte{2}))
	close(ch)
	chunkReader := NewChunkReader(ch)

	src := bytes.NewReader(bytes.Repeat([]byte{0xAA}, 100))
	bounded := newChunkBoundedReader(src, chunkReader, 4)

	buf := make([]byte, 4)
	n, err := bounded.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)

	// Budget exhausted and no physical chunk boundary crossed yet: further reads must fail.
	_, err = bounded.Read(buf)
	require.ErrorIs(t, err, snapshottypes.ErrDecompressedChunkTooLarge)
	_, err = bounded.Read(buf)
	require.ErrorIs(t, err, snapshottypes.ErrDecompressedChunkTooLarge)

	// Drive chunkReader across the boundary into the second physical chunk.
	one := make([]byte, 1)
	_, err = chunkReader.Read(one)
	require.NoError(t, err)
	require.Equal(t, 1, chunkReader.chunksOpened)
	_, err = chunkReader.Read(one)
	require.NoError(t, err)
	require.Equal(t, 2, chunkReader.chunksOpened)

	// The new chunk grants a fresh budget.
	n, err = bounded.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
}
