package unorderedtx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

const (
	// DefaultMaxUnOrderedTTL defines the default maximum TTL an un-ordered transaction
	// can set.
	DefaultMaxUnOrderedTTL = 1024

	dirName  = "unordered_txs"
	fileName = "data"
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

	// dataDir defines the directory to store unexpired unordered transactions
	//
	// XXX: Note, ideally we avoid the need to store unexpired unordered transactions
	// directly to file. However, store v1 does not allow such a primitive. But,
	// once store v2 is fully integrated, we can remove manual file handling and
	// store the unexpired unordered transactions directly to SS.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/18467
	dataDir string

	mu sync.RWMutex
	// txHashes defines a map from tx hash -> TTL value, which is used for duplicate
	// checking and replay protection, as well as purging the map when the TTL is
	// expired.
	txHashes map[TxHash]uint64
}

func NewManager(dataDir string) *Manager {
	path := filepath.Join(dataDir, dirName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		_ = os.Mkdir(path, os.ModePerm)
	}

	m := &Manager{
		dataDir:  dataDir,
		blockCh:  make(chan uint64, 16),
		doneCh:   make(chan struct{}),
		txHashes: make(map[TxHash]uint64),
	}

	return m
}

func (m *Manager) Start() {
	go m.purgeLoop()
}

// Close must be called when a node gracefully shuts down. Typically, this should
// be called in an application's Close() function, which is called by the server.
// Note, Start() must be called in order for Close() to not hang.
//
// It will free all necessary resources as well as writing all unexpired unordered
// transactions along with their TTL values to file.
func (m *Manager) Close() error {
	close(m.blockCh)
	<-m.doneCh
	m.blockCh = nil

	return m.flushToFile()
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

// OnInit must be called when a node starts up. Typically, this should be called
// in an application's constructor, which is called by the server.
func (m *Manager) OnInit() error {
	f, err := os.Open(filepath.Join(m.dataDir, dirName, fileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File does not exist, which we can assume that there are no unexpired
			// unordered transactions.
			return nil
		}

		return fmt.Errorf("failed to open unconfirmed txs file: %w", err)
	}
	defer f.Close()

	var (
		r   = bufio.NewReader(f)
		buf = make([]byte, chunkSize)
	)
	for {
		n, err := io.ReadFull(r, buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return fmt.Errorf("failed to read unconfirmed txs file: %w", err)
			}
		}
		if n != 32+8 {
			return fmt.Errorf("read unexpected number of bytes from unconfirmed txs file: %d", n)
		}

		var txHash TxHash
		copy(txHash[:], buf[:txHashSize])

		m.Add(txHash, binary.BigEndian.Uint64(buf[txHashSize:]))
	}

	return nil
}

// OnNewBlock sends the latest block number to the background purge loop, which
// should be called in ABCI Commit event.
func (m *Manager) OnNewBlock(blockHeight uint64) {
	m.blockCh <- blockHeight
}

func (m *Manager) exportSnapshot(height uint64, snapshotWriter func([]byte) error) error {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	keys := maps.Keys(m.txHashes)
	sort.Slice(keys, func(i, j int) bool { return bytes.Compare(keys[i][:], keys[j][:]) < 0 })

	for _, txHash := range keys {
		ttl := m.txHashes[txHash]
		if height > ttl {
			// skip expired txs that have yet to be purged
			continue
		}

		chunk := unorderedTxToBytes(txHash, ttl)

		if _, err := w.Write(chunk); err != nil {
			return fmt.Errorf("failed to write unordered tx to buffer: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush unordered txs buffer: %w", err)
	}

	return snapshotWriter(buf.Bytes())
}

// flushToFile writes all unexpired unordered transactions along with their TTL
// to file, overwriting the existing file if it exists.
func (m *Manager) flushToFile() error {
	f, err := os.Create(filepath.Join(m.dataDir, dirName, fileName))
	if err != nil {
		return fmt.Errorf("failed to create unordered txs file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for txHash, ttl := range m.txHashes {
		chunk := unorderedTxToBytes(txHash, ttl)

		if _, err = w.Write(chunk); err != nil {
			return fmt.Errorf("failed to write unordered tx to buffer: %w", err)
		}
	}

	if err = w.Flush(); err != nil {
		return fmt.Errorf("failed to flush unordered txs buffer: %w", err)
	}

	return nil
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

func unorderedTxToBytes(txHash TxHash, ttl uint64) []byte {
	chunk := make([]byte, chunkSize)
	copy(chunk[:txHashSize], txHash[:])

	ttlBz := make([]byte, ttlSize)
	binary.BigEndian.PutUint64(ttlBz, ttl)
	copy(chunk[txHashSize:], ttlBz)

	return chunk
}
