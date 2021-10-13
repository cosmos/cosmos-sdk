package snapshots_test

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
)

func checksums(slice [][]byte) [][]byte {
	hasher := sha256.New()
	checksums := make([][]byte, len(slice))
	for i, chunk := range slice {
		hasher.Write(chunk)
		checksums[i] = hasher.Sum(nil)
		hasher.Reset()
	}
	return checksums
}

func hash(chunks [][]byte) []byte {
	hasher := sha256.New()
	for _, chunk := range chunks {
		hasher.Write(chunk)
	}
	return hasher.Sum(nil)
}

func makeChunks(chunks [][]byte) <-chan io.ReadCloser {
	ch := make(chan io.ReadCloser, len(chunks))
	for _, chunk := range chunks {
		ch <- io.NopCloser(bytes.NewReader(chunk))
	}
	close(ch)
	return ch
}

func readChunks(chunks <-chan io.ReadCloser) [][]byte {
	bodies := [][]byte{}
	for chunk := range chunks {
		body, err := io.ReadAll(chunk)
		if err != nil {
			panic(err)
		}
		bodies = append(bodies, body)
	}
	return bodies
}

type mockSnapshotter struct {
	chunks [][]byte
}

func (m *mockSnapshotter) Restore(
	height uint64, format uint32, chunks <-chan io.ReadCloser, ready chan<- struct{},
) error {
	if format == 0 {
		return types.ErrUnknownFormat
	}
	if m.chunks != nil {
		return errors.New("already has contents")
	}
	if ready != nil {
		close(ready)
	}

	m.chunks = [][]byte{}
	for reader := range chunks {
		chunk, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		m.chunks = append(m.chunks, chunk)
	}

	return nil
}

func (m *mockSnapshotter) Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error) {
	if format == 0 {
		return nil, types.ErrUnknownFormat
	}
	ch := make(chan io.ReadCloser, len(m.chunks))
	for _, chunk := range m.chunks {
		ch <- io.NopCloser(bytes.NewReader(chunk))
	}
	close(ch)
	return ch, nil
}

// setupBusyManager creates a manager with an empty store that is busy creating a snapshot at height 1.
// The snapshot will complete when the returned closer is called.
func setupBusyManager(t *testing.T) *snapshots.Manager {
	// os.MkdirTemp() is used instead of testing.T.TempDir()
	// see https://github.com/cosmos/cosmos-sdk/pull/8475 for
	// this change's rationale.
	tempdir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tempdir) })

	store, err := snapshots.NewStore(db.NewMemDB(), tempdir)
	require.NoError(t, err)
	hung := newHungSnapshotter()
	mgr := snapshots.NewManager(store, hung)

	go func() {
		_, err := mgr.Create(1)
		require.NoError(t, err)
	}()
	time.Sleep(10 * time.Millisecond)
	t.Cleanup(hung.Close)

	return mgr
}

// hungSnapshotter can be used to test operations in progress. Call close to end the snapshot.
type hungSnapshotter struct {
	ch chan struct{}
}

func newHungSnapshotter() *hungSnapshotter {
	return &hungSnapshotter{
		ch: make(chan struct{}),
	}
}

func (m *hungSnapshotter) Close() {
	close(m.ch)
}

func (m *hungSnapshotter) Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error) {
	<-m.ch
	ch := make(chan io.ReadCloser, 1)
	ch <- io.NopCloser(bytes.NewReader([]byte{}))
	return ch, nil
}

func (m *hungSnapshotter) Restore(
	height uint64, format uint32, chunks <-chan io.ReadCloser, ready chan<- struct{},
) error {
	panic("not implemented")
}
