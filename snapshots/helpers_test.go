package snapshots_test

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	protoio "github.com/gogo/protobuf/io"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

// snapshotItems serialize a array of bytes as SnapshotItem_ExtensionPayload, and return the chunks.
func snapshotItems(items [][]byte) [][]byte {
	// copy the same parameters from the code
	snapshotChunkSize := uint64(10e6)
	snapshotBufferSize := int(snapshotChunkSize)

	ch := make(chan io.ReadCloser)
	go func() {
		chunkWriter := snapshots.NewChunkWriter(ch, snapshotChunkSize)
		bufWriter := bufio.NewWriterSize(chunkWriter, snapshotBufferSize)
		zWriter, _ := zlib.NewWriterLevel(bufWriter, 7)
		protoWriter := protoio.NewDelimitedWriter(zWriter)
		for _, item := range items {
			_ = types.WriteExtensionItem(protoWriter, item)
		}
		_ = protoWriter.Close()
		_ = zWriter.Close()
		_ = bufWriter.Flush()
		_ = chunkWriter.Close()
	}()

	var chunks [][]byte
	for chunkBody := range ch {
		chunk, err := io.ReadAll(chunkBody)
		if err != nil {
			panic(err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks
}

type mockSnapshotter struct {
	items [][]byte
}

func (m *mockSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (types.SnapshotItem, error) {
	if format == 0 {
		return types.SnapshotItem{}, types.ErrUnknownFormat
	}
	if m.items != nil {
		return types.SnapshotItem{}, errors.New("already has contents")
	}

	m.items = [][]byte{}
	for {
		item := &types.SnapshotItem{}
		err := protoReader.ReadMsg(item)
		if err == io.EOF {
			break
		} else if err != nil {
			return types.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}
		payload := item.GetExtensionPayload()
		if payload == nil {
			return types.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}
		m.items = append(m.items, payload.Payload)
	}

	return types.SnapshotItem{}, nil
}

func (m *mockSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	for _, item := range m.items {
		if err := types.WriteExtensionItem(protoWriter, item); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockSnapshotter) SnapshotFormat() uint32 {
	return types.CurrentFormat
}
func (m *mockSnapshotter) SupportedFormats() []uint32 {
	return []uint32{types.CurrentFormat}
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
	mgr := snapshots.NewManager(store, hung, nil)

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

func (m *hungSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	<-m.ch
	return nil
}

func (m *hungSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (types.SnapshotItem, error) {
	panic("not implemented")
}
