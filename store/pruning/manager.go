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
	db     dbm.DB
	logger log.Logger
	opts   types.PruningOptions
	// Snapshots are taken in a separate goroutine from the regular execution
	// and can be delivered asynchronously via HandleSnapshotHeight.
	// Therefore, we sync access to pruneSnapshotHeights, snapshotInterval and initFromDB with this mutex.
	pruneSnapshotHeightsMx sync.RWMutex
	// These are the heights that are multiples of snapshotInterval and kept for state sync snapshots.
	// The heights are added to be pruned when a snapshot is complete.
	pruneSnapshotHeights []int64
	snapshotInterval     uint64
	initFromDB           bool
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

// HandleSnapshotHeight persists the snapshot height to be pruned at the next appropriate
// height defined by the pruning strategy. It flushes the update to disk and panics if the flush fails.
// The input height must be greater than 0, and the pruning strategy must not be set to pruning nothing.
// If either of these conditions is not met, this function does nothing.
func (m *Manager) HandleSnapshotHeight(height int64) {
	if m.opts.GetPruningStrategy() == types.PruningNothing || height <= 0 {
		return
	}

	if m.snapshotInterval == 0 || uint64(height)%m.snapshotInterval != 0 {
		// ensure that we keep only heights that match the current snap interval
		return
	}

	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()

	m.logger.Debug("HandleSnapshotHeight", "height", height)
	m.pruneSnapshotHeights = append(m.pruneSnapshotHeights, height)
	sort.Slice(m.pruneSnapshotHeights, func(i, j int) bool { return m.pruneSnapshotHeights[i] < m.pruneSnapshotHeights[j] })

	if m.initFromDB && len(m.pruneSnapshotHeights) > 2 { // received 2 fresh snap heights as base level
		m.pruneSnapshotHeights = m.pruneSnapshotHeights[1:] // drop db state element
		m.initFromDB = false
	}

	k := 1
	for ; k < len(m.pruneSnapshotHeights); k++ {
		// any one missing in the sequence means they are still in flight
		if m.pruneSnapshotHeights[k] != m.pruneSnapshotHeights[k-1]+int64(m.snapshotInterval) {
			break
		}
	}
	// compact the heights list when in-flight snapshots have landed
	m.pruneSnapshotHeights = m.pruneSnapshotHeights[k-1:]

	if n := len(m.pruneSnapshotHeights); n > 100 {
		m.logger.Warn("Snapshot heights state in pruning manager grows unexpected ", "total", height)
	}

	// flush the max height to disk so that they are not lost if crash happens.
	if err := m.db.SetSync(pruneSnapshotHeightsKey, int64SliceToBytes(slices.Max(m.pruneSnapshotHeights))); err != nil {
		panic(err)
	}
}

// SetSnapshotInterval sets the interval at which the snapshots are taken.
func (m *Manager) SetSnapshotInterval(snapshotInterval uint64) {
	m.pruneSnapshotHeightsMx.Lock()
	defer m.pruneSnapshotHeightsMx.Unlock()

	if m.snapshotInterval != 0 && snapshotInterval != m.snapshotInterval {
		// interval is modified, we can not predict in flight snapshots anymore
		m.pruneSnapshotHeights = make([]int64, 0, 1)
	}
	m.snapshotInterval = snapshotInterval
}

// GetPruningHeight returns the height which can prune upto if it is able to prune at the given height.
func (m *Manager) GetPruningHeight(height int64) int64 {
	if m.opts.GetPruningStrategy() == types.PruningNothing {
		return 0
	}
	if m.opts.Interval <= 0 {
		return 0
	}

	if height%int64(m.opts.Interval) != 0 || height <= int64(m.opts.KeepRecent) {
		return 0
	}

	// Consider the snapshot height
	pruneHeight := height - 1 - int64(m.opts.KeepRecent) // we should keep the current height at least

	m.pruneSnapshotHeightsMx.RLock()
	defer m.pruneSnapshotHeightsMx.RUnlock()

	// snapshotInterval is zero, indicating that all heights can be pruned
	if m.snapshotInterval <= 0 {
		return pruneHeight
	}

	if len(m.pruneSnapshotHeights) == 0 { // do not prune before an initial snapshot
		return 0
	}

	// the snapshot `m.pruneSnapshotHeights[0]` is already operated,
	snHeight := m.pruneSnapshotHeights[0] - 1
	if m.pruneSnapshotHeights[0]%int64(m.snapshotInterval) == 0 { // ensure non legacy data
		// it is safe to prune up to `m.pruneSnapshotHeights[0] + int64(m.snapshotInterval) - 1`
		// before hitting any in-flight snapshot
		snHeight += int64(m.snapshotInterval)
	}
	return min(snHeight, pruneHeight)
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
	m.initFromDB = true
	return nil
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
