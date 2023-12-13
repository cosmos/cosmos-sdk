package unorderedtx

import (
	"context"
	"sync"
	"time"
)

const (
	// DefaultMaxUnOrderedTTL defines the default maximum TTL an un-ordered transaction
	// can set.
	DefaultMaxUnOrderedTTL = 1024
)

// TxHash defines a transaction hash type alias, which is a fixed array of 32 bytes.
type TxHash [32]byte

// Manager contains the tx hash dictionary for duplicates checking, and expire
// them when block production progresses.
type Manager struct {
	// blockCh defines a channel to receive newly committed block heights
	blockCh chan uint64
	// doneCh allows us to ensure the purgeLoop has gracefully terminated prior to closing
	doneCh chan struct{}

	mu sync.RWMutex
	// txHashes defines a map from tx hash -> TTL value, which is used for duplicate
	// checking and replay protection, as well as purging the map when the TTL is
	// expired.
	txHashes map[TxHash]uint64
}

func NewManager() *Manager {
	m := &Manager{
		blockCh:  make(chan uint64, 16),
		doneCh:   make(chan struct{}),
		txHashes: make(map[TxHash]uint64),
	}

	return m
}

func (m *Manager) Start() {
	go m.purgeLoop()
}

func (m *Manager) Close() error {
	close(m.blockCh)
	<-m.doneCh
	m.blockCh = nil

	return nil
}

func (m *Manager) Contains(hash TxHash) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.txHashes[hash]
	return ok
}

func (m *Manager) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.txHashes)
}

func (m *Manager) Add(txHash TxHash, ttl uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.txHashes[txHash] = ttl
}

// OnNewBlock send the latest block number to the background purge loop, which
// should be called in ABCI Commit event.
func (m *Manager) OnNewBlock(blockHeight uint64) {
	m.blockCh <- blockHeight
}

// expiredTxs returns expired tx hashes based on the provided block height.
func (m *Manager) expiredTxs(blockHeight uint64) []TxHash {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []TxHash
	for txHash, ttl := range m.txHashes {
		if blockHeight > ttl {
			result = append(result, txHash)
		}
	}

	return result
}

func (m *Manager) purge(txHashes []TxHash) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, txHash := range txHashes {
		delete(m.txHashes, txHash)
	}
}

// purgeLoop removes expired tx hashes in the background
func (m *Manager) purgeLoop() {
	for {
		latestHeight, ok := m.batchReceive()
		if !ok {
			// channel closed
			m.doneCh <- struct{}{}
			return
		}

		hashes := m.expiredTxs(latestHeight)
		if len(hashes) > 0 {
			m.purge(hashes)
		}
	}
}

func (m *Manager) batchReceive() (uint64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var latestHeight uint64
	for {
		select {
		case <-ctx.Done():
			return latestHeight, true

		case blockHeight, ok := <-m.blockCh:
			if !ok {
				// channel is closed
				return 0, false
			}
			if blockHeight > latestHeight {
				latestHeight = blockHeight
			}
		}
	}
}
