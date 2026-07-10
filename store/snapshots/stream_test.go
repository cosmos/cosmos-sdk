package snapshots

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	snapshottypes "github.com/cosmos/cosmos-sdk/store/v2/snapshots/types"
)

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

	// budget exhausted, no boundary crossed yet
	_, err = bounded.Read(buf)
	require.ErrorIs(t, err, snapshottypes.ErrDecompressedChunkTooLarge)
	_, err = bounded.Read(buf)
	require.ErrorIs(t, err, snapshottypes.ErrDecompressedChunkTooLarge)

	// cross into the second physical chunk
	one := make([]byte, 1)
	_, err = chunkReader.Read(one)
	require.NoError(t, err)
	require.Equal(t, 1, chunkReader.chunksOpened)
	_, err = chunkReader.Read(one)
	require.NoError(t, err)
	require.Equal(t, 2, chunkReader.chunksOpened)

	// new chunk grants a fresh budget
	n, err = bounded.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
}
