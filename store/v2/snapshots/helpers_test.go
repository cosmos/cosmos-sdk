package snapshots_test

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	protoio "github.com/cosmos/gogoproto/io"
	"github.com/stretchr/testify/require"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
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
func snapshotItems(items [][]byte, ext snapshots.ExtensionSnapshotter) [][]byte {
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
			_ = snapshotstypes.WriteExtensionPayload(protoWriter, item)
		}
		// write extension metadata
		_ = protoWriter.WriteMsg(&snapshotstypes.SnapshotItem{
			Item: &snapshotstypes.SnapshotItem_Extension{
				Extension: &snapshotstypes.SnapshotExtensionMeta{
					Name:   ext.SnapshotName(),
					Format: ext.SnapshotFormat(),
				},
			},
		})
		_ = ext.SnapshotExtension(0, func(payload []byte) error {
			return snapshotstypes.WriteExtensionPayload(protoWriter, payload)
		})
		_ = protoWriter.Close()
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

type mockCommitSnapshotter struct {
	items [][]byte
}

func (m *mockCommitSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshotstypes.SnapshotItem, error) {
	if format == 0 {
		return snapshotstypes.SnapshotItem{}, snapshotstypes.ErrUnknownFormat
	}
	if m.items != nil {
		return snapshotstypes.SnapshotItem{}, errors.New("already has contents")
	}

	var item snapshotstypes.SnapshotItem
	m.items = [][]byte{}
	for {
		item.Reset()
		err := protoReader.ReadMsg(&item)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return snapshotstypes.SnapshotItem{}, fmt.Errorf("invalid protobuf message: %w", err)
		}
		payload := item.GetExtensionPayload()
		if payload == nil {
			break
		}
		m.items = append(m.items, payload.Payload)
	}

	return item, nil
}

func (m *mockCommitSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	for _, item := range m.items {
		if err := snapshotstypes.WriteExtensionPayload(protoWriter, item); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockCommitSnapshotter) SnapshotFormat() uint32 {
	return snapshotstypes.CurrentFormat
}

func (m *mockCommitSnapshotter) SupportedFormats() []uint32 {
	return []uint32{snapshotstypes.CurrentFormat}
}

type mockErrorCommitSnapshotter struct{}

var _ snapshots.CommitSnapshotter = (*mockErrorCommitSnapshotter)(nil)

func (m *mockErrorCommitSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	return errors.New("mock snapshot error")
}

func (m *mockErrorCommitSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshotstypes.SnapshotItem, error) {
	return snapshotstypes.SnapshotItem{}, errors.New("mock restore error")
}

func (m *mockErrorCommitSnapshotter) SnapshotFormat() uint32 {
	return snapshotstypes.CurrentFormat
}

func (m *mockErrorCommitSnapshotter) SupportedFormats() []uint32 {
	return []uint32{snapshotstypes.CurrentFormat}
}

// setupBusyManager creates a manager with an empty store that is busy creating a snapshot at height 1.
// The snapshot will complete when the returned closer is called.
func setupBusyManager(t *testing.T) *snapshots.Manager {
	t.Helper()
	store, err := snapshots.NewStore(t.TempDir())
	require.NoError(t, err)
	hung := newHungCommitSnapshotter()
	mgr := snapshots.NewManager(store, opts, hung, nil, coretesting.NewNopLogger())

	// Channel to ensure the test doesn't finish until the goroutine is done.
	// Without this, there are intermittent test failures about
	// the t.TempDir() cleanup failing due to the directory not being empty.
	done := make(chan struct{})

	go func() {
		defer close(done)
		_, err := mgr.Create(1)
		require.NoError(t, err)
	}()
	time.Sleep(10 * time.Millisecond)

	t.Cleanup(func() {
		<-done
	})

	t.Cleanup(hung.Close)

	return mgr
}

// hungCommitSnapshotter can be used to test operations in progress. Call close to end the snapshot.
type hungCommitSnapshotter struct {
	ch chan struct{}
}

func newHungCommitSnapshotter() *hungCommitSnapshotter {
	return &hungCommitSnapshotter{
		ch: make(chan struct{}),
	}
}

func (m *hungCommitSnapshotter) Close() {
	close(m.ch)
}

func (m *hungCommitSnapshotter) Snapshot(height uint64, protoWriter protoio.Writer) error {
	<-m.ch
	return nil
}

func (m *hungCommitSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshotstypes.SnapshotItem, error) {
	panic("not implemented")
}

type extSnapshotter struct {
	state []uint64
}

func newExtSnapshotter(count int) *extSnapshotter {
	state := make([]uint64, 0, count)
	for i := 0; i < count; i++ {
		state = append(state, uint64(i))
	}
	return &extSnapshotter{
		state,
	}
}

func (s *extSnapshotter) SnapshotName() string {
	return "mock"
}

func (s *extSnapshotter) SnapshotFormat() uint32 {
	return 1
}

func (s *extSnapshotter) SupportedFormats() []uint32 {
	return []uint32{1}
}

func (s *extSnapshotter) SnapshotExtension(height uint64, payloadWriter snapshots.ExtensionPayloadWriter) error {
	for _, i := range s.state {
		if err := payloadWriter(snapshotstypes.Uint64ToBigEndian(i)); err != nil {
			return err
		}
	}
	return nil
}

func (s *extSnapshotter) RestoreExtension(height uint64, format uint32, payloadReader snapshots.ExtensionPayloadReader) error {
	for {
		payload, err := payloadReader()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		s.state = append(s.state, snapshotstypes.BigEndianToUint64(payload))
	}
	// finalize restoration
	return nil
}
