package pruning

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/cosmos/cosmos-sdk/pruning/types"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type Manager struct {
	logger               log.Logger
	opts                 *types.PruningOptions
	snapshotInterval     uint64
	pruneHeights         []int64
	pruneSnapshotHeights *list.List
	mx                   sync.Mutex
}

const errNegativeHeightsFmt = "failed to get pruned heights: %d"

var (
	pruneHeightsKey         = []byte("s/pruneheights")
	pruneSnapshotHeightsKey = []byte("s/pruneSnheights")
)

func NewManager(logger log.Logger) *Manager {
	return &Manager{
		logger:       logger,
		opts:         types.NewPruningOptions(types.PruningNothing),
		pruneHeights: []int64{},
		// These are the heights that are multiples of snapshotInterval and kept for state sync snapshots.
		// The heights are added to this list to be pruned when a snapshot is complete.
		pruneSnapshotHeights: list.New(),
		mx:                   sync.Mutex{},
	}
}

// SetOptions sets the pruning strategy on the manager.
func (m *Manager) SetOptions(opts *types.PruningOptions) {
	m.opts = opts
}

// GetOptions fetches the pruning strategy from the manager.
func (m *Manager) GetOptions() *types.PruningOptions {
	return m.opts
}

// GetPruningHeights returns all heights to be pruned during the next call to Prune().
func (m *Manager) GetPruningHeights() []int64 {
	return m.pruneHeights
}

// ResetPruningHeights resets the heights to be pruned.
func (m *Manager) ResetPruningHeights() {
	// reuse previously allocated memory.
	m.pruneHeights = m.pruneHeights[:0]
}

// HandleHeight determines if pruneHeight height needs to be kept for pruning at the right interval prescribed by
// the pruning strategy. Returns true if the given height was kept to be pruned at the next call to Prune(), false otherwise
func (m *Manager) HandleHeight(previousHeight int64) int64 {
	if m.opts.GetPruningStrategy() == types.PruningNothing {
		return 0
	}

	defer func() {
		// handle persisted snapshot heights
		m.mx.Lock()
		defer m.mx.Unlock()
		var next *list.Element
		for e := m.pruneSnapshotHeights.Front(); e != nil; e = next {
			snHeight := e.Value.(int64)
			if snHeight < previousHeight-int64(m.opts.KeepRecent) {
				m.pruneHeights = append(m.pruneHeights, snHeight)

				// We must get next before removing to be able to continue iterating.
				next = e.Next()
				m.pruneSnapshotHeights.Remove(e)
			} else {
				next = e.Next()
			}
		}
	}()

	if int64(m.opts.KeepRecent) < previousHeight {
		pruneHeight := previousHeight - int64(m.opts.KeepRecent)
		// We consider this height to be pruned iff:
		//
		// - snapshotInterval is zero as that means that all heights should be pruned.
		// - snapshotInterval % (height - KeepRecent) != 0 as that means the height is not
		// a 'snapshot' height.
		if m.snapshotInterval == 0 || pruneHeight%int64(m.snapshotInterval) != 0 {
			m.pruneHeights = append(m.pruneHeights, pruneHeight)
			return pruneHeight
		}
	}
	return 0
}

func (m *Manager) HandleHeightSnapshot(height int64) {
	if m.opts.GetPruningStrategy() == types.PruningNothing {
		return
	}
	m.mx.Lock()
	defer m.mx.Unlock()
	m.logger.Debug("HandleHeightSnapshot", "height", height)
	m.pruneSnapshotHeights.PushBack(height)
}

// SetSnapshotInterval sets the interval at which the snapshots are taken.
func (m *Manager) SetSnapshotInterval(snapshotInterval uint64) {
	m.snapshotInterval = snapshotInterval
}

// ShouldPruneAtHeight return true if the given height should be pruned, false otherwise
func (m *Manager) ShouldPruneAtHeight(height int64) bool {
	return m.opts.Interval > 0 && m.opts.GetPruningStrategy() != types.PruningNothing && height%int64(m.opts.Interval) == 0
}

// FlushPruningHeights flushes the pruning heights to the database for crash recovery.
func (m *Manager) FlushPruningHeights(batch dbm.Batch) {
	if m.opts.GetPruningStrategy() == types.PruningNothing {
		return
	}
	m.flushPruningHeights(batch)
	m.flushPruningSnapshotHeights(batch)
}

// LoadPruningHeights loads the pruning heights from the database as a crash recovery.
func (m *Manager) LoadPruningHeights(db dbm.DB) error {
	if m.opts.GetPruningStrategy() == types.PruningNothing {
		return nil
	}
	if err := m.loadPruningHeights(db); err != nil {
		return err
	}
	return m.loadPruningSnapshotHeights(db)
}

func (m *Manager) loadPruningHeights(db dbm.DB) error {
	bz, err := db.Get(pruneHeightsKey)
	if err != nil {
		return fmt.Errorf("failed to get pruned heights: %w", err)
	}
	if len(bz) == 0 {
		return nil
	}

	prunedHeights := make([]int64, len(bz)/8)
	i, offset := 0, 0
	for offset < len(bz) {
		h := int64(binary.BigEndian.Uint64(bz[offset : offset+8]))
		if h < 0 {
			return fmt.Errorf(errNegativeHeightsFmt, h)
		}

		prunedHeights[i] = h
		i++
		offset += 8
	}

	if len(prunedHeights) > 0 {
		m.pruneHeights = prunedHeights
	}

	return nil
}

func (m *Manager) loadPruningSnapshotHeights(db dbm.DB) error {
	bz, err := db.Get(pruneSnapshotHeightsKey)
	if err != nil {
		return fmt.Errorf("failed to get post-snapshot pruned heights: %w", err)
	}
	if len(bz) == 0 {
		return nil
	}

	pruneSnapshotHeights := list.New()
	i, offset := 0, 0
	for offset < len(bz) {
		h := int64(binary.BigEndian.Uint64(bz[offset : offset+8]))
		if h < 0 {
			return fmt.Errorf(errNegativeHeightsFmt, h)
		}

		pruneSnapshotHeights.PushBack(h)
		i++
		offset += 8
	}

	m.mx.Lock()
	defer m.mx.Unlock()
	m.pruneSnapshotHeights = pruneSnapshotHeights

	return nil
}

func (m *Manager) flushPruningHeights(batch dbm.Batch) {
	batch.Set(pruneHeightsKey, int64SliceToBytes(m.pruneHeights))
}

func (m *Manager) flushPruningSnapshotHeights(batch dbm.Batch) {
	m.mx.Lock()
	defer m.mx.Unlock()
	batch.Set(pruneSnapshotHeightsKey, listToBytes(m.pruneSnapshotHeights))
}

// TODO: convert to a generic version with Go 1.18.
func int64SliceToBytes(slice []int64) []byte {
	bz := make([]byte, 0, len(slice)*8)
	for _, ph := range slice {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(ph))
		bz = append(bz, buf...)
	}
	return bz
}

func listToBytes(list *list.List) []byte {
	bz := make([]byte, 0, list.Len()*8)
	for e := list.Front(); e != nil; e = e.Next() {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(e.Value.(int64)))
		bz = append(bz, buf...)
	}
	return bz
}
