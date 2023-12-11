package ante

import (
	"context"
	"crypto/sha256"
	"sync"
	"time"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// DefaultMaxUnOrderedTTL defines the default maximum TTL an un-ordered transaction
	// can set.
	DefaultMaxUnOrderedTTL = 1024
)

// TxHash defines a transaction hash type alias, which is a fixed array of 32 bytes.
type TxHash [32]byte

// UnorderedTxManager contains the tx hash dictionary for duplicates checking,
// and expire them when block production progresses.
type UnorderedTxManager struct {
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

func NewUnorderedTxManager() *UnorderedTxManager {
	m := &UnorderedTxManager{
		blockCh:  make(chan uint64, 16),
		doneCh:   make(chan struct{}),
		txHashes: make(map[TxHash]uint64),
	}

	return m
}

func (m *UnorderedTxManager) Start() {
	go m.purgeLoop()
}

func (m *UnorderedTxManager) Close() error {
	close(m.blockCh)
	<-m.doneCh
	m.blockCh = nil

	return nil
}

func (m *UnorderedTxManager) Contains(hash TxHash) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.txHashes[hash]
	return ok
}

func (m *UnorderedTxManager) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.txHashes)
}

func (m *UnorderedTxManager) Add(txHash TxHash, ttl uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.txHashes[txHash] = ttl
}

// OnNewBlock send the latest block number to the background purge loop, which
// should be called in ABCI Commit event.
func (m *UnorderedTxManager) OnNewBlock(blockHeight uint64) {
	m.blockCh <- blockHeight
}

// expiredTxs returns expired tx hashes based on the provided block height.
func (m *UnorderedTxManager) expiredTxs(blockHeight uint64) []TxHash {
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

func (m *UnorderedTxManager) purge(txHashes []TxHash) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, txHash := range txHashes {
		delete(m.txHashes, txHash)
	}
}

// purgeLoop removes expired tx hashes in the background
func (m *UnorderedTxManager) purgeLoop() {
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

func (m *UnorderedTxManager) batchReceive() (uint64, bool) {
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

var _ sdk.AnteDecorator = (*UnorderedTxDecorator)(nil)

// UnorderedTxDecorator defines an AnteHandler decorator that is responsible for
// checking if a transaction is intended to be unordered and if so, evaluates
// the transaction accordingly. An unordered transaction will bypass having it's
// nonce incremented, which allows fire-and-forget along with possible parallel
// transaction processing, without having to deal with nonces.
//
// The transaction sender must ensure that unordered=true and a timeout_height
// is appropriately set. The AnteHandler will check that the transaction is not
// a duplicate and will evict it from memory when the timeout is reached.
//
// The UnorderedTxDecorator should be placed as early as possible in the AnteHandler
// chain to ensure that during DeliverTx, the transaction is added to the UnorderedTxManager.
type UnorderedTxDecorator struct {
	// maxUnOrderedTTL defines the maximum TTL a transaction can define.
	maxUnOrderedTTL uint64
	txManager       *UnorderedTxManager
}

func NewUnorderedTxDecorator(maxTTL uint64, m *UnorderedTxManager) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		maxUnOrderedTTL: maxTTL,
	}
}

func (d *UnorderedTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	unorderedTx, ok := tx.(sdk.TxWithUnordered)
	if !ok || !unorderedTx.GetUnordered() {
		// If the transaction does not implement unordered capabilities or has the
		// unordered value as false, we bypass.
		return next(ctx, tx, simulate)
	}

	if unorderedTx.GetTimeoutHeight() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "unordered transaction must have timeout_height set")
	}

	if unorderedTx.GetTimeoutHeight() > uint64(ctx.BlockHeight())+d.maxUnOrderedTTL {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unordered tx ttl exceeds %d", d.maxUnOrderedTTL)
	}

	txHash := sha256.Sum256(ctx.TxBytes())

	// check for duplicates
	if d.txManager.Contains(txHash) {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "tx %X is duplicated")
	}

	if ctx.ExecMode() == sdk.ExecModeFinalize {
		// a new tx included in the block, add the hash to the unordered tx manager
		d.txManager.Add(txHash, unorderedTx.GetTimeoutHeight())
	}

	return next(ctx, tx, simulate)
}
