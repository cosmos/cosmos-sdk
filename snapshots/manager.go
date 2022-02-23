package snapshots

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"sort"
	"sync"

	"github.com/cosmos/cosmos-sdk/snapshots/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	opNone     operation = ""
	opSnapshot operation = "snapshot"
	opPrune    operation = "prune"
	opRestore  operation = "restore"

	chunkBufferSize = 4

	snapshotMaxItemSize = int(64e6) // SDK has no key/value size limit, so we set an arbitrary limit
)

// operation represents a Manager operation. Only one operation can be in progress at a time.
type operation string

// restoreDone represents the result of a restore operation.
type restoreDone struct {
	complete bool  // if true, restore completed successfully (not prematurely)
	err      error // if non-nil, restore errored
}

// Manager manages snapshot and restore operations for an app, making sure only a single
// long-running operation is in progress at any given time, and provides convenience methods
// mirroring the ABCI interface.
//
// Although the ABCI interface (and this manager) passes chunks as byte slices, the internal
// snapshot/restore APIs use IO streams (i.e. chan io.ReadCloser), for two reasons:
//
// 1) In the future, ABCI should support streaming. Consider e.g. InitChain during chain
//    upgrades, which currently passes the entire chain state as an in-memory byte slice.
//    https://github.com/tendermint/tendermint/issues/5184
//
// 2) io.ReadCloser streams automatically propagate IO errors, and can pass arbitrary
//    errors via io.Pipe.CloseWithError().
type Manager struct {
	store      *Store
	multistore types.Snapshotter
	extensions map[string]types.ExtensionSnapshotter

	mtx                sync.Mutex
	operation          operation
	chRestore          chan<- io.ReadCloser
	chRestoreDone      <-chan restoreDone
	restoreChunkHashes [][]byte
	restoreChunkIndex  uint32
}

// NewManager creates a new manager.
func NewManager(store *Store, multistore types.Snapshotter, extensions map[string]types.ExtensionSnapshotter) *Manager {
	return &Manager{
		store:      store,
		multistore: multistore,
		extensions: extensions,
	}
}

// RegisterExtensions register extension snapshotters to manager
func (m *Manager) RegisterExtensions(extensions ...types.ExtensionSnapshotter) error {
	for _, extension := range extensions {
		name := extension.SnapshotName()
		if _, ok := m.extensions[name]; ok {
			return fmt.Errorf("duplicated snapshotter name: %s", name)
		}
		if !IsFormatSupported(extension, extension.SnapshotFormat()) {
			return fmt.Errorf("snapshotter don't support it's own snapshot format: %s %d", name, extension.SnapshotFormat())
		}
		m.extensions[name] = extension
	}
	return nil
}

// begin starts an operation, or errors if one is in progress. It manages the mutex itself.
func (m *Manager) begin(op operation) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.beginLocked(op)
}

// beginLocked begins an operation while already holding the mutex.
func (m *Manager) beginLocked(op operation) error {
	if op == opNone {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "can't begin a none operation")
	}
	if m.operation != opNone {
		return sdkerrors.Wrapf(sdkerrors.ErrConflict, "a %v operation is in progress", m.operation)
	}
	m.operation = op
	return nil
}

// end ends the current operation.
func (m *Manager) end() {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.endLocked()
}

// endLocked ends the current operation while already holding the mutex.
func (m *Manager) endLocked() {
	m.operation = opNone
	if m.chRestore != nil {
		close(m.chRestore)
		m.chRestore = nil
	}
	m.chRestoreDone = nil
	m.restoreChunkHashes = nil
	m.restoreChunkIndex = 0
}

// sortedExtensionNames sort extension names for deterministic iteration.
func (m *Manager) sortedExtensionNames() []string {
	names := make([]string, 0, len(m.extensions))
	for name := range m.extensions {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Create creates a snapshot and returns its metadata.
func (m *Manager) Create(height uint64) (*types.Snapshot, error) {
	if m == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "no snapshot store configured")
	}
	err := m.begin(opSnapshot)
	if err != nil {
		return nil, err
	}
	defer m.end()

	latest, err := m.store.GetLatest()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to examine latest snapshot")
	}
	if latest != nil && latest.Height >= height {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrConflict,
			"a more recent snapshot already exists at height %v", latest.Height)
	}

	// Spawn goroutine to generate snapshot chunks and pass their io.ReadClosers through a channel
	ch := make(chan io.ReadCloser)
	go m.createSnapshot(height, ch)

	return m.store.Save(height, types.CurrentFormat, ch)
}

// createSnapshot do the heavy work of snapshotting after the validations of request are done
// the produced chunks are written to the channel.
func (m *Manager) createSnapshot(height uint64, ch chan<- io.ReadCloser) {
	streamWriter := NewStreamWriter(ch)
	if streamWriter == nil {
		return
	}
	defer streamWriter.Close()
	if err := m.multistore.Snapshot(height, streamWriter); err != nil {
		streamWriter.CloseWithError(err)
		return
	}
	for _, name := range m.sortedExtensionNames() {
		extension := m.extensions[name]
		// write extension metadata
		err := streamWriter.WriteMsg(&types.SnapshotItem{
			Item: &types.SnapshotItem_Extension{
				Extension: &types.SnapshotExtensionMeta{
					Name:   name,
					Format: extension.SnapshotFormat(),
				},
			},
		})
		if err != nil {
			streamWriter.CloseWithError(err)
			return
		}
		if err := extension.Snapshot(height, streamWriter); err != nil {
			streamWriter.CloseWithError(err)
			return
		}
	}
}

// List lists snapshots, mirroring ABCI ListSnapshots. It can be concurrent with other operations.
func (m *Manager) List() ([]*types.Snapshot, error) {
	return m.store.List()
}

// LoadChunk loads a chunk into a byte slice, mirroring ABCI LoadChunk. It can be called
// concurrently with other operations. If the chunk does not exist, nil is returned.
func (m *Manager) LoadChunk(height uint64, format uint32, chunk uint32) ([]byte, error) {
	reader, err := m.store.LoadChunk(height, format, chunk)
	if err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, nil
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Prune prunes snapshots, if no other operations are in progress.
func (m *Manager) Prune(retain uint32) (uint64, error) {
	err := m.begin(opPrune)
	if err != nil {
		return 0, err
	}
	defer m.end()
	return m.store.Prune(retain)
}

// Restore begins an async snapshot restoration, mirroring ABCI OfferSnapshot. Chunks must be fed
// via RestoreChunk() until the restore is complete or a chunk fails.
func (m *Manager) Restore(snapshot types.Snapshot) error {
	if snapshot.Chunks == 0 {
		return sdkerrors.Wrap(types.ErrInvalidMetadata, "no chunks")
	}
	if uint32(len(snapshot.Metadata.ChunkHashes)) != snapshot.Chunks {
		return sdkerrors.Wrapf(types.ErrInvalidMetadata, "snapshot has %v chunk hashes, but %v chunks",
			uint32(len(snapshot.Metadata.ChunkHashes)),
			snapshot.Chunks)
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// check multistore supported format preemptive
	if snapshot.Format != types.CurrentFormat {
		return sdkerrors.Wrapf(types.ErrUnknownFormat, "snapshot format %v", snapshot.Format)
	}
	if snapshot.Height == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "cannot restore snapshot at height 0")
	}
	if snapshot.Height > uint64(math.MaxInt64) {
		return sdkerrors.Wrapf(types.ErrInvalidMetadata,
			"snapshot height %v cannot exceed %v", snapshot.Height, int64(math.MaxInt64))
	}

	err := m.beginLocked(opRestore)
	if err != nil {
		return err
	}

	// Start an asynchronous snapshot restoration, passing chunks and completion status via channels.
	chChunks := make(chan io.ReadCloser, chunkBufferSize)
	chDone := make(chan restoreDone, 1)

	go func() {
		err := m.restoreSnapshot(snapshot, chChunks)
		chDone <- restoreDone{
			complete: err == nil,
			err:      err,
		}
		close(chDone)
	}()

	m.chRestore = chChunks
	m.chRestoreDone = chDone
	m.restoreChunkHashes = snapshot.Metadata.ChunkHashes
	m.restoreChunkIndex = 0
	return nil
}

// restoreSnapshot do the heavy work of snapshot restoration after preliminary checks on request have passed.
func (m *Manager) restoreSnapshot(snapshot types.Snapshot, chChunks <-chan io.ReadCloser) error {
	streamReader, err := NewStreamReader(chChunks)
	if err != nil {
		return err
	}
	defer streamReader.Close()

	next, err := m.multistore.Restore(snapshot.Height, snapshot.Format, streamReader)
	if err != nil {
		return sdkerrors.Wrap(err, "multistore restore")
	}
	for {
		if next.Item == nil {
			// end of stream
			break
		}
		metadata := next.GetExtension()
		if metadata == nil {
			return sdkerrors.Wrapf(sdkerrors.ErrLogic, "unknown snapshot item %T", next.Item)
		}
		extension, ok := m.extensions[metadata.Name]
		if !ok {
			return sdkerrors.Wrapf(sdkerrors.ErrLogic, "unknown extension snapshotter %s", metadata.Name)
		}
		if !IsFormatSupported(extension, metadata.Format) {
			return sdkerrors.Wrapf(types.ErrUnknownFormat, "format %v for extension %s", metadata.Format, metadata.Name)
		}
		next, err = extension.Restore(snapshot.Height, metadata.Format, streamReader)
		if err != nil {
			return sdkerrors.Wrapf(err, "extension %s restore", metadata.Name)
		}
	}
	return nil
}

// RestoreChunk adds a chunk to an active snapshot restoration, mirroring ABCI ApplySnapshotChunk.
// Chunks must be given until the restore is complete, returning true, or a chunk errors.
func (m *Manager) RestoreChunk(chunk []byte) (bool, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if m.operation != opRestore {
		return false, sdkerrors.Wrap(sdkerrors.ErrLogic, "no restore operation in progress")
	}

	if int(m.restoreChunkIndex) >= len(m.restoreChunkHashes) {
		return false, sdkerrors.Wrap(sdkerrors.ErrLogic, "received unexpected chunk")
	}

	// Check if any errors have occurred yet.
	select {
	case done := <-m.chRestoreDone:
		m.endLocked()
		if done.err != nil {
			return false, done.err
		}
		return false, sdkerrors.Wrap(sdkerrors.ErrLogic, "restore ended unexpectedly")
	default:
	}

	// Verify the chunk hash.
	hash := sha256.Sum256(chunk)
	expected := m.restoreChunkHashes[m.restoreChunkIndex]
	if !bytes.Equal(hash[:], expected) {
		return false, sdkerrors.Wrapf(types.ErrChunkHashMismatch,
			"expected %x, got %x", hash, expected)
	}

	// Pass the chunk to the restore, and wait for completion if it was the final one.
	m.chRestore <- io.NopCloser(bytes.NewReader(chunk))
	m.restoreChunkIndex++

	if int(m.restoreChunkIndex) >= len(m.restoreChunkHashes) {
		close(m.chRestore)
		m.chRestore = nil
		done := <-m.chRestoreDone
		m.endLocked()
		if done.err != nil {
			return false, done.err
		}
		if !done.complete {
			return false, sdkerrors.Wrap(sdkerrors.ErrLogic, "restore ended prematurely")
		}
		return true, nil
	}
	return false, nil
}

// IsFormatSupported returns if the snapshotter supports restoration from given format.
func IsFormatSupported(snapshotter types.ExtensionSnapshotter, format uint32) bool {
	for _, i := range snapshotter.SupportedFormats() {
		if i == format {
			return true
		}
	}
	return false
}
