package pruning

import (
	"encoding/binary"
	"fmt"
	"slices"
	"sort"
	"sync"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/store/pruning/types"
)

// Manager is an abstraction to handle the logic needed for
// determining when to prune old heights of the store
// based on the strategy described by the pruning options.
type Manager struct {
	db               dbm.DB
	logger           log.Logger
	opts             types.PruningOptions
	snapshotInterval uint64
	// Snapshots are taken in a separate goroutine from the regular execution
	// and can be delivered asynchronously via HandleSnapshotHeight.
	// Therefore, we sync access to pruneSnapshotHeights, inflightSnapshotHeights and initFromStore with this mutex.
	pruneSnapshotHeightsMx sync.RWMutex
	// These are the heights that are multiples of snapshotInterval and kept for state sync snapshots.
	// The heights are added to be pruned when a snapshot is complete.
	pruneSnapshotHeights    []int64
	inflightSnapshotHeights []int64
	initFromStore           bool
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
// The retuned manager uses a pruning strategy of "nothing" which
// keeps all heights. Users of the Manager may change the strategy
// by calling SetOptions.
func NewManager(db dbm.DB, logger log.Logger) *Manager {
	return &Manager{
		db:                   db,
		logger:               logger,
		opts:                 types.NewPruningOptions(types.PruningNothing),
		pruneSnapshotHeights: []int64{0}, // init with 0 block height
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

// AnnounceSnapshotHeight announces a new snapshot height for tracking and pruning.
func (m *Manager) AnnounceSnapshotHeight(height int64) {
	if m.opts.GetPruningStrategy() == types.PruningNothing || height <= 0 {
		return
	}
	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()
	// called in ascending order so no sorting required
	m.inflightSnapshotHeights = append(m.inflightSnapshotHeights, height)
}

// HandleSnapshotHeight persists the snapshot height to be pruned at the next appropriate
// height defined by the pruning strategy. It flushes the update to disk and panics if the flush fails.
// The input height must be greater than 0, and the pruning strategy must not be set to pruning nothing.
// If either of these conditions is not met, this function does nothing.
func (m *Manager) HandleSnapshotHeight(height int64) {
	if m.opts.GetPruningStrategy() == types.PruningNothing || height <= 0 {
		return
	}

	m.logger.Debug("HandleSnapshotHeight", "height", height)

	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()

	// remove from the in-flight list
	if position := slices.Index(m.inflightSnapshotHeights, height); position != -1 {
		m.inflightSnapshotHeights = append(m.inflightSnapshotHeights[:position], m.inflightSnapshotHeights[position+1:]...)
	}

	if m.initFromStore {
		// drop the legacy state as it may belong to a different interval or an outdated snapshot
		// that is not in sequence with the current one
		m.pruneSnapshotHeights = m.pruneSnapshotHeights[1:]
		m.initFromStore = false
	}

	m.pruneSnapshotHeights = append(m.pruneSnapshotHeights, height)
	sort.Slice(m.pruneSnapshotHeights, func(i, j int) bool { return m.pruneSnapshotHeights[i] < m.pruneSnapshotHeights[j] })

	// in-flight snapshots may land out of order due to the concurrent nature of the snapshotter.
	// we need to detect them to prevent pruning their heights while the snapshots are still in progress.
	k := 1
	for ; k < len(m.pruneSnapshotHeights); k++ {
		if m.pruneSnapshotHeights[k] != m.pruneSnapshotHeights[k-1]+int64(m.snapshotInterval) {
			// gap detected, snapshot is in-flight
			break
		}
	}
	// compact the height list for the snapshots in sequence
	// the last snapshot height is used to allow pruning up to the next interval height
	m.pruneSnapshotHeights = m.pruneSnapshotHeights[k-1:]

	// flush the max height to store so that they are not lost if a crash happens.
	// only the max height matters as there are no in-flight snapshots after a restart
	if err := storePruningSnapshotHeight(m.db, slices.Max(m.pruneSnapshotHeights)); err != nil {
		panic(err)
	}
}

// SetSnapshotInterval sets the interval at which the snapshots are taken.
// This value should be set on startup and not exceed max int64 (2^63-1). Concurrent modifications are not supported.
func (m *Manager) SetSnapshotInterval(snapshotInterval uint64) {
	m.snapshotInterval = snapshotInterval
}

// GetPruningHeight returns the height which can prune upto if it is able to prune at the given height.
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

	if len(m.pruneSnapshotHeights) == 0 { // do not prune before an initial snapshot
		return 0
	}

	// highest version based on completed snapshots
	snHeight := m.pruneSnapshotHeights[0] - 1
	if !m.initFromStore { // ensure non-legacy data
		// with no inflight snapshots, we may prune up to the next snap interval -1
		snHeight += int64(m.snapshotInterval)
	}
	if len(m.inflightSnapshotHeights) == 0 {
		return min(snHeight, pruneHeight)
	}
	// highest version based on started snapshots
	inFlightHeight := m.inflightSnapshotHeights[0] - 1
	return min(snHeight, pruneHeight, inFlightHeight)
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
	m.pruneSnapshotHeights = []int64{slices.Max(loadedPruneSnapshotHeights)}
	m.initFromStore = true
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
	bz := make([]byte, 0, len(slice)*8)
	for _, ph := range slice {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(ph))
		bz = append(bz, buf...)
	}
	return bz
}
