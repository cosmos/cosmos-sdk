package snapshots_test

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
	"testing"
	"time"

	protoio "github.com/gogo/protobuf/io"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	snaphotsTestUtil "github.com/cosmos/cosmos-sdk/testutil/snapshots"
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
			types.WriteExtensionItem(protoWriter, item)
		}
		protoWriter.Close()
		zWriter.Close()
		bufWriter.Flush()
		chunkWriter.Close()
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
	prunedHeights    map[int64]struct{}
	snapshotInterval uint64
	items            [][]byte
}

func (m *mockSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshottypes.SnapshotItem, error) {
	if format == 0 {
		return snapshottypes.SnapshotItem{}, types.ErrUnknownFormat
	}
	if m.items != nil {
		return snapshottypes.SnapshotItem{}, errors.New("already has contents")
	}

	m.items = [][]byte{}
	for {
		item := &snapshottypes.SnapshotItem{}
		err := protoReader.ReadMsg(item)
		if err == io.EOF {
			break
		} else if err != nil {
			return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}
		payload := item.GetExtensionPayload()
		if payload == nil {
			return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}
		m.items = append(m.items, payload.Payload)
	}

	return snapshottypes.SnapshotItem{}, nil
}

func (m *mockSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	for _, item := range m.items {
		if err := types.WriteExtensionItem(protoWriter, item); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockSnapshotter) PruneSnapshotHeight(height int64) {
	m.prunedHeights[height] = struct{}{}
}

func (m *mockSnapshotter) GetSnapshotInterval() uint64 {
	return m.snapshotInterval
}

func (m *mockSnapshotter) SetSnapshotInterval(snapshotInterval uint64) {
	m.snapshotInterval = snapshotInterval
}

func (m *mockSnapshotter) SnapshotFormat() uint32 {
	return 2
}

func (m *mockSnapshotter) SupportedFormats() []uint32 {
	return []uint32{2}
}

// setupBusyManager creates a manager with an empty store that is busy creating a snapshot at height 1.
// The snapshot will complete when the returned closer is called.
func setupBusyManager(t *testing.T) *snapshots.Manager {
	tempdir := snaphotsTestUtil.GetTempDir(t)

	store, err := snapshots.NewStore(db.NewMemDB(), tempdir)
	require.NoError(t, err)
	hung := newHungSnapshotter()
	hung.SetSnapshotInterval(opts.Interval)
	mgr := snapshots.NewManager(store, opts, hung, log.NewNopLogger())
	require.Equal(t, opts.Interval, hung.snapshotInterval)

	go func() {
		_, err := mgr.Create(1)
		require.NoError(t, err)
		_, didPruneHeight := hung.prunedHeights[1]
		require.True(t, didPruneHeight)
	}()
	time.Sleep(10 * time.Millisecond)
	t.Cleanup(hung.Close)

	return mgr
}

// hungSnapshotter can be used to test operations in progress. Call close to end the snapshot.
type hungSnapshotter struct {
	ch               chan struct{}
	prunedHeights    map[int64]struct{}
	snapshotInterval uint64
}

func newHungSnapshotter() *hungSnapshotter {
	return &hungSnapshotter{
		ch:            make(chan struct{}),
		prunedHeights: make(map[int64]struct{}),
	}
}

func (m *hungSnapshotter) Close() {
	close(m.ch)
}

func (m *hungSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	<-m.ch
	return nil
}

func (m *hungSnapshotter) PruneSnapshotHeight(height int64) {
	m.prunedHeights[height] = struct{}{}
}

func (m *hungSnapshotter) SetSnapshotInterval(snapshotInterval uint64) {
	m.snapshotInterval = snapshotInterval
}

func (m *hungSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshottypes.SnapshotItem, error) {
	panic("not implemented")
}
