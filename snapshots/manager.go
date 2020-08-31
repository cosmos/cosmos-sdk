package snapshots

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/snapshots/types"
)

const (
	opNone     operation = ""
	opSnapshot operation = "snapshot"
	opPrune    operation = "prune"
	opRestore  operation = "restore"

	chunkBufferSize = 4
)

// operation represents a Manager operation. Only one operation can be in progress at a time.
type operation string

// Manager manages snapshot and restore operations for an app, making sure only a single
// long-running operation is in progress at any given time, and provides convenience methods
// mirroring the ABCI interface.
type Manager struct {
	store  *Store
	target types.Snapshotter

	mtx                sync.Mutex
	operation          operation
	chRestore          chan<- io.ReadCloser
	chRestoreDone      <-chan error
	restoreChunkHashes [][]byte
	restoreChunkIndex  uint32
}

// NewManager creates a new manager.
func NewManager(store *Store, target types.Snapshotter) *Manager {
	return &Manager{
		store:  store,
		target: target,
	}
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
		return errors.New("can't begin a none operation")
	}
	if m.operation != opNone {
		return fmt.Errorf("a %v operation is in progress", m.operation)
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

// Create creates a snapshot and returns its metadata.
func (m *Manager) Create(height uint64) (*types.Snapshot, error) {
	if m == nil {
		return nil, errors.New("no snapshot store configured")
	}
	err := m.begin(opSnapshot)
	if err != nil {
		return nil, err
	}
	defer m.end()

	latest, err := m.store.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to examine latest snapshot: %w", err)
	}
	if latest != nil && latest.Height >= height {
		return nil, fmt.Errorf("a more recent snapshot already exists at height %v", latest.Height)
	}

	chunks, err := m.target.Snapshot(height, types.CurrentFormat)
	if err != nil {
		return nil, err
	}
	return m.store.Save(height, types.CurrentFormat, chunks)
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

	return ioutil.ReadAll(reader)
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
		return fmt.Errorf("%w: no chunks", types.ErrInvalidMetadata)
	}
	if uint32(len(snapshot.Metadata.ChunkHashes)) != snapshot.Chunks {
		return fmt.Errorf("%w: snapshot has %v chunk hashes, but %v chunks",
			types.ErrInvalidMetadata,
			uint32(len(snapshot.Metadata.ChunkHashes)),
			snapshot.Chunks)
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()
	err := m.beginLocked(opRestore)
	if err != nil {
		return err
	}

	// Start an asynchronous snapshot restoration, passing chunks and completion status via channels.
	chChunks := make(chan io.ReadCloser, chunkBufferSize)
	chDone := make(chan error, 1)
	go func() {
		chDone <- m.target.Restore(snapshot.Height, snapshot.Format, chChunks)
		close(chDone)
	}()

	// Check for any initial errors from the restore, before any chunks are fed.
	select {
	case err := <-chDone:
		if err == nil {
			err = errors.New("restore ended unexpectedly")
		}
		m.endLocked()
		return err
	case <-time.After(20 * time.Millisecond):
	}

	m.chRestore = chChunks
	m.chRestoreDone = chDone
	m.restoreChunkHashes = snapshot.Metadata.ChunkHashes
	m.restoreChunkIndex = 0
	return nil
}

// RestoreChunk adds a chunk to an active snapshot restoration, mirroring ABCI ApplySnapshotChunk.
// Chunks must be given until the restore is complete, returning true, or a chunk errors.
func (m *Manager) RestoreChunk(chunk []byte) (bool, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if m.operation != opRestore {
		return false, fmt.Errorf("no restore operation in progress")
	}

	if int(m.restoreChunkIndex) >= len(m.restoreChunkHashes) {
		return false, errors.New("received unexpected chunk")
	}

	// Check if any errors have occurred yet.
	select {
	case err := <-m.chRestoreDone:
		if err == nil {
			err = errors.New("restore ended unexpectedly")
		}
		m.endLocked()
		return false, err
	default:
	}

	// Verify the chunk hash.
	hash := sha256.Sum256(chunk)
	expected := m.restoreChunkHashes[m.restoreChunkIndex]
	if !bytes.Equal(hash[:], expected) {
		return false, fmt.Errorf("%w (expected %x, got %x)",
			types.ErrChunkHashMismatch, hash, expected)
	}

	// Pass the chunk to the restore, and wait for completion if it was the final one.
	m.chRestore <- ioutil.NopCloser(bytes.NewReader(chunk))
	m.restoreChunkIndex++

	if int(m.restoreChunkIndex) >= len(m.restoreChunkHashes) {
		close(m.chRestore)
		m.chRestore = nil
		err := <-m.chRestoreDone
		m.endLocked()
		return true, err
	}
	return false, nil
}
