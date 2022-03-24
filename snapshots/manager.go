package snapshots

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
	"sync"

	"github.com/cosmos/cosmos-sdk/snapshots/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/tendermint/libs/log"
)

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
	// store is the snapshot store where all completed snapshots are persisted.
	store  *Store
	opts   *types.SnapshotOptions
	// target is the store from which snapshots are taken.
	target types.Snapshotter
	logger log.Logger

	mtx                sync.Mutex
	operation          operation
	chRestore          chan<- io.ReadCloser
	chRestoreDone      <-chan restoreDone
	restoreChunkHashes [][]byte
	restoreChunkIndex  uint32
}

// operation represents a Manager operation. Only one operation can be in progress at a time.
type operation string

// restoreDone represents the result of a restore operation.
type restoreDone struct {
	complete bool  // if true, restore completed successfully (not prematurely)
	err      error // if non-nil, restore errored
}

const (
	opNone     operation = ""
	opSnapshot operation = "snapshot"
	opPrune    operation = "prune"
	opRestore  operation = "restore"

	chunkBufferSize = 4
)

var (
	ErrOptsZeroSnapshotInterval = errors.New("snaphot-interval must not be 0")
)

// NewManager creates a new manager.
func NewManager(store *Store, opts *types.SnapshotOptions, target types.Snapshotter, logger log.Logger) *Manager {
	target.SetSnapshotInterval(opts.Interval)
	return &Manager{
		store:  store,
		opts:   opts,
		target: target,
		logger: logger,
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

// GetInterval returns snapshot interval.
func (m *Manager) GetInterval() uint64 {
	return m.opts.Interval
}

// GetKeepRecent returns snapshot keep-recent.
func (m *Manager) GetKeepRecent() uint32 {
	return m.opts.KeepRecent
}

// GetSnapshotBlockRetentionHeights returns the number of heights needed
// for block retention. Blocks since the oldest available snapshot must be
// available for state sync nodes to catch up (oldest because a node may be
// restoring an old snapshot while a new snapshot was taken).
func (m *Manager) GetSnapshotBlockRetentionHeights() int64 {
	return int64(m.opts.Interval * uint64(m.opts.KeepRecent))
}

// Create creates a snapshot and returns its metadata.
func (m *Manager) Create(height uint64) (*types.Snapshot, error) {
	if m == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "no snapshot store configured")
	}

	defer m.target.PruneSnapshotHeight(int64(height))

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
		return sdkerrors.Wrap(types.ErrInvalidMetadata, "no chunks")
	}
	if uint32(len(snapshot.Metadata.ChunkHashes)) != snapshot.Chunks {
		return sdkerrors.Wrapf(types.ErrInvalidMetadata, "snapshot has %v chunk hashes, but %v chunks",
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
	chReady := make(chan struct{}, 1)
	chDone := make(chan restoreDone, 1)
	go func() {
		err := m.target.Restore(snapshot.Height, snapshot.Format, chChunks, chReady)
		chDone <- restoreDone{
			complete: err == nil,
			err:      err,
		}
		close(chDone)
	}()

	// Check for any initial errors from the restore, before any chunks are fed.
	select {
	case done := <-chDone:
		m.endLocked()
		if done.err != nil {
			return done.err
		}
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "restore ended unexpectedly")
	case <-chReady:
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
	m.chRestore <- ioutil.NopCloser(bytes.NewReader(chunk))
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

// SnapshotIfApplicable takes a snapshot of the current state if we are on a snapshot height. 
// It also prunes any old snapshots. The snapshotting and pruning happen in separate goroutines.
func (m *Manager) SnapshotIfApplicable(height int64) {
	if m == nil {
		return
	}
	if !m.shouldTakeSnapshot(height) {
		m.logger.Debug("snapshot is skipped", "height", height)
		return
	}
	go m.snapshot(height)
}

// shouldTakeSnapshot returns true is snapshot should be taken at height.
func (m *Manager) shouldTakeSnapshot(height int64) bool {
	return m.opts.Interval > 0 && uint64(height)%m.opts.Interval == 0
}

func (m *Manager) snapshot(height int64) {
	m.logger.Info("creating state snapshot", "height", height)

	snapshot, err := m.Create(uint64(height))
	if err != nil {
		m.logger.Error("failed to create state snapshot", "height", height, "err", err)
		return
	}

	m.logger.Info("completed state snapshot", "height", height, "format", snapshot.Format)

	if m.opts.KeepRecent > 0 {
		m.logger.Debug("pruning state snapshots")

		pruned, err := m.Prune(m.opts.KeepRecent)
		if err != nil {
			m.logger.Error("Failed to prune state snapshots", "err", err)
			return
		}

		m.logger.Debug("pruned state snapshots", "pruned", pruned)
	}
}
