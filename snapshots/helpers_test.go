package snapshots_test

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
)

func checksum(b []byte) []byte {
	hash := sha256.Sum256(b)
	return hash[:]
}

func checksums(slice [][]byte) [][]byte {
	checksums := [][]byte{}
	for _, chunk := range slice {
		checksums = append(checksums, checksum(chunk))
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
		ch <- ioutil.NopCloser(bytes.NewReader(chunk))
	}
	close(ch)
	return ch
}

func readChunks(chunks <-chan io.ReadCloser) [][]byte {
	bodies := [][]byte{}
	for chunk := range chunks {
		body, err := ioutil.ReadAll(chunk)
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

func (m *mockSnapshotter) Restore(height uint64, format uint32, chunks <-chan io.ReadCloser) error {
	if format == 0 {
		return types.ErrUnknownFormat
	}
	if m.chunks != nil {
		return errors.New("already has contents")
	}

	m.chunks = [][]byte{}
	for reader := range chunks {
		chunk, err := ioutil.ReadAll(reader)
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
		ch <- ioutil.NopCloser(bytes.NewReader(chunk))
	}
	close(ch)
	return ch, nil
}

// setupBusyManager creates a manager with an empty store that is busy taking a snapshot at height 1.
// The snapshot will complete when the returned closer is called.
func setupBusyManager(t *testing.T) (*snapshots.Manager, func()) {
	tempdir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	store, err := snapshots.NewStore(db.NewMemDB(), tempdir)
	require.NoError(t, err)
	hung := newHungSnapshotter()
	mgr := snapshots.NewManager(store, hung)

	go func() {
		_, err := mgr.Take(1)
		require.NoError(t, err)
	}()
	time.Sleep(10 * time.Millisecond)

	closer := func() {
		hung.Close()
		os.RemoveAll(tempdir)
	}
	return mgr, closer
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
	ch <- ioutil.NopCloser(bytes.NewReader([]byte{}))
	return ch, nil
}

func (m *hungSnapshotter) Restore(height uint64, format uint32, chunks <-chan io.ReadCloser) error {
	panic("not implemented")
}
