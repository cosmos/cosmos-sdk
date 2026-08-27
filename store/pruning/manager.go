package pruning

import (
	"encoding/binary"
	"fmt"
	"slices"
	"sync"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
)

// Manager is an abstraction to handle the logic needed for
// determining when to prune old heights of the store
// based on the strategy described by the pruning options.
type Manager struct {
	db               dbm.DB
	logger           log.Logger
	opts             types.PruningOptions
	snapshotInterval uint64
	// Snapshots are taken in a separate goroutine from regular execution.
	pruneSnapshotHeightsMx  sync.RWMutex
	completedSnapshotHeight int64
	inflightSnapshotHeights map[int64]struct{}
	loadedFromDisk          bool
}

// NegativeHeightsError is returned when a negative height is provided to the manager.
type NegativeHeightsError struct {
	Height int64
}

var _ error = &NegativeHeightsError{}

func (e *NegativeHeightsError) Error() string {
	return fmt.Sprintf("failed to get pruned heights: %d", e.Height)
}

var pruneSnapshotHeightsKey = []byte("s/prunesnapshotheights")

// NewManager returns a new Manager with the given db and logger.
// The returned manager uses a pruning strategy of "nothing" which
// keeps all heights. Users of the Manager may change the strategy
// by calling SetOptions.
func NewManager(db dbm.DB, logger log.Logger) *Manager {
	return &Manager{
		db:                      db,
		logger:                  logger,
		opts:                    types.NewPruningOptions(types.PruningNothing),
		inflightSnapshotHeights: make(map[int64]struct{}),
	}
}

// SetOptions sets the pruning strategy on the manager.
func (m *Manager) SetOptions(opts types.PruningOptions) {
	m.opts = opts
}

// GetOptions fetches the pruning strategy from the manager.
func (m *Manager) GetOptions() types.PruningOptions {
	return m.opts
}

// StartSnapshot tracks a snapshot while it is being created.
func (m *Manager) StartSnapshot(height int64) {
	if m.opts.GetPruningStrategy() == types.PruningNothing || height <= 0 {
		return
	}
	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()
	m.inflightSnapshotHeights[height] = struct{}{}
}

// AnnounceSnapshotHeight is kept for compatibility with legacy callers.
func (m *Manager) AnnounceSnapshotHeight(height int64) {
	m.StartSnapshot(height)
}

// FailSnapshot removes a failed snapshot height from in-flight tracking.
func (m *Manager) FailSnapshot(height int64) {
	if m.opts.GetPruningStrategy() == types.PruningNothing || height <= 0 {
		return
	}
	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()
	delete(m.inflightSnapshotHeights, height)
}

// HandleSnapshotHeight persists the snapshot height to be pruned at the next appropriate
// height defined by the pruning strategy. It flushes the update to disk and panics if the flush fails.
// The input height must be greater than 0, and the pruning strategy must not be set to pruning nothing.
// If either of these conditions is not met, this function does nothing.
func (m *Manager) CompleteSnapshot(height int64) {
	if m.opts.GetPruningStrategy() == types.PruningNothing || height <= 0 {
		return
	}

	m.logger.Debug("CompleteSnapshot", "height", height)

	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()

	delete(m.inflightSnapshotHeights, height)

	if height > m.completedSnapshotHeight {
		m.completedSnapshotHeight = height
		m.loadedFromDisk = false
	}

	// flush the max height to store so that they are not lost if a crash happens.
	// only the max height matters as there are no in-flight snapshots after a restart
	if err := storePruningSnapshotHeight(m.db, m.completedSnapshotHeight); err != nil {
		panic(err)
	}
}

// HandleSnapshotHeight is kept for compatibility with legacy callers.
func (m *Manager) HandleSnapshotHeight(height int64) {
	m.CompleteSnapshot(height)
}

// SetSnapshotInterval sets the interval at which the snapshots are taken.
// This value should be set on startup and not exceed max int64 (2^63-1). Concurrent modifications are not supported.
func (m *Manager) SetSnapshotInterval(snapshotInterval uint64) {
	m.snapshotInterval = snapshotInterval
}

// GetPruningHeight returns the height which can prune up to if it is able to prune at the given height.
func (m *Manager) GetPruningHeight(height int64) int64 {
	if m.opts.GetPruningStrategy() == types.PruningNothing ||
		m.opts.Interval <= 0 ||
		height <= int64(m.opts.KeepRecent) ||
		height%int64(m.opts.Interval) != 0 {
		return 0
	}

	// Consider the snapshot height
	pruneHeight := height - 1 - int64(m.opts.KeepRecent) // we should keep the current height at least

	// snapshotInterval is zero, indicating that all heights can be pruned
	if m.snapshotInterval <= 0 {
		return pruneHeight
	}

	m.pruneSnapshotHeightsMx.RLock()
	defer m.pruneSnapshotHeightsMx.RUnlock()

	completedLimit := m.completedSnapshotHeight - 1
	if !m.loadedFromDisk {
		completedLimit += int64(m.snapshotInterval)
	}
	if len(m.inflightSnapshotHeights) == 0 {
		return min(completedLimit, pruneHeight)
	}
	inFlightHeight := int64(^uint64(0) >> 1)
	for snapshotHeight := range m.inflightSnapshotHeights {
		inFlightHeight = min(inFlightHeight, snapshotHeight-1)
	}
	return min(completedLimit, pruneHeight, inFlightHeight)
}

// LoadSnapshotHeights loads the snapshot heights from the database as a crash recovery.
func (m *Manager) LoadSnapshotHeights(db dbm.DB) error {
	if m.opts.GetPruningStrategy() == types.PruningNothing {
		return nil
	}

	// loading list for backwards compatibility
	loadedPruneSnapshotHeights, err := loadPruningSnapshotHeights(db)
	if err != nil {
		return err
	}

	if len(loadedPruneSnapshotHeights) == 0 {
		return nil
	}
	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()
	// restore max only as there are no in-flight snapshots after a restart
	m.completedSnapshotHeight = slices.Max(loadedPruneSnapshotHeights)
	m.loadedFromDisk = true
	return nil
}

func storePruningSnapshotHeight(db dbm.DB, val int64) error {
	return db.SetSync(pruneSnapshotHeightsKey, int64SliceToBytes(val))
}

func loadPruningSnapshotHeights(db dbm.DB) ([]int64, error) {
	bz, err := db.Get(pruneSnapshotHeightsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get post-snapshot pruned heights: %w", err)
	}
	if len(bz) == 0 {
		return []int64{}, nil
	}

	pruneSnapshotHeights := make([]int64, len(bz)/8)
	i, offset := 0, 0
	for offset < len(bz) {
		h := int64(binary.BigEndian.Uint64(bz[offset : offset+8]))
		if h < 0 {
			return nil, &NegativeHeightsError{Height: h}
		}
		pruneSnapshotHeights[i] = h
		i++
		offset += 8
	}

	return pruneSnapshotHeights, nil
}

func int64SliceToBytes(slice ...int64) []byte {
	bz := make([]byte, len(slice)*8)
	for i, ph := range slice {
		binary.BigEndian.PutUint64(bz[i<<3:], uint64(ph))
	}
	return bz
}
