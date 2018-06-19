package lcd

import (
	"encoding/hex"
	"sort"
	"sync"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

type memStoreProvider struct {
	mtx sync.RWMutex
	// byHeight is always sorted by Height... need to support range search (nil, h]
	// btree would be more efficient for larger sets
	byHeight fullCommits
	byHash   map[string]FullCommit

	sorted bool
}

// fullCommits just exists to allow easy sorting
type fullCommits []FullCommit

func (s fullCommits) Len() int      { return len(s) }
func (s fullCommits) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s fullCommits) Less(i, j int) bool {
	return s[i].Height() < s[j].Height()
}

// NewMemStoreProvider returns a new in-memory provider.
func NewMemStoreProvider() Provider {
	return &memStoreProvider{
		byHeight: fullCommits{},
		byHash:   map[string]FullCommit{},
	}
}

func (m *memStoreProvider) encodeHash(hash []byte) string {
	return hex.EncodeToString(hash)
}

// StoreCommit stores a FullCommit after verifying it.
func (m *memStoreProvider) StoreCommit(fc FullCommit) error {
	// make sure the fc is self-consistent before saving
	err := fc.ValidateBasic(fc.Commit.Header.ChainID)
	if err != nil {
		return err
	}

	// store the valid fc
	key := m.encodeHash(fc.ValidatorsHash())

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.byHash[key] = fc
	m.byHeight = append(m.byHeight, fc)
	m.sorted = false
	return nil
}

// GetByHeight returns the FullCommit for height h or an error if the commit is not found.
func (m *memStoreProvider) GetByHeight(h int64) (FullCommit, error) {
	// By heuristics, GetByHeight with linearsearch is fast enough
	// for about 50 keys but after that, it needs binary search.
	// See https://github.com/tendermint/tendermint/pull/1043#issue-285188242
	m.mtx.RLock()
	n := len(m.byHeight)
	m.mtx.RUnlock()

	if n <= 50 {
		return m.getByHeightLinearSearch(h)
	}
	return m.getByHeightBinarySearch(h)
}

func (m *memStoreProvider) sortByHeightIfNecessaryLocked() {
	if !m.sorted {
		sort.Sort(m.byHeight)
		m.sorted = true
	}
}

func (m *memStoreProvider) getByHeightLinearSearch(h int64) (FullCommit, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.sortByHeightIfNecessaryLocked()
	// search from highest to lowest
	for i := len(m.byHeight) - 1; i >= 0; i-- {
		if fc := m.byHeight[i]; fc.Height() <= h {
			return fc, nil
		}
	}
	return FullCommit{}, lcdErr.ErrCommitNotFound()
}

func (m *memStoreProvider) getByHeightBinarySearch(h int64) (FullCommit, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.sortByHeightIfNecessaryLocked()
	low, high := 0, len(m.byHeight)-1
	var mid int
	var hmid int64
	var midFC FullCommit
	// Our goal is to either find:
	//   * item ByHeight with the query
	//   * greatest height with a height <= query
	for low <= high {
		mid = int(uint(low+high) >> 1) // Avoid an overflow
		midFC = m.byHeight[mid]
		hmid = midFC.Height()
		switch {
		case hmid == h:
			return midFC, nil
		case hmid < h:
			low = mid + 1
		case hmid > h:
			high = mid - 1
		}
	}

	if high >= 0 {
		if highFC := m.byHeight[high]; highFC.Height() < h {
			return highFC, nil
		}
	}
	return FullCommit{}, lcdErr.ErrCommitNotFound()
}

// GetByHash returns the FullCommit for the hash or an error if the commit is not found.
func (m *memStoreProvider) GetByHash(hash []byte) (FullCommit, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	fc, ok := m.byHash[m.encodeHash(hash)]
	if !ok {
		return fc, lcdErr.ErrCommitNotFound()
	}
	return fc, nil
}

// LatestCommit returns the latest FullCommit or an error if no commits exist.
func (m *memStoreProvider) LatestCommit() (FullCommit, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	l := len(m.byHeight)
	if l == 0 {
		return FullCommit{}, lcdErr.ErrCommitNotFound()
	}
	m.sortByHeightIfNecessaryLocked()
	return m.byHeight[l-1], nil
}
