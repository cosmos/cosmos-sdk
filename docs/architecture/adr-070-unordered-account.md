# ADR 070: Un-Ordered Transaction Inclusion

## Changelog

* Dec 4, 2023: Initial Draft

## Status

ACCEPTED

## Abstract

We propose a way to do replay-attack protection without enforcing the order of
transactions, without requiring the use of nonces. In this way, we can support
un-ordered transaction inclusion.

## Context

As of today, the nonce value (account sequence number) prevents replay-attack and
ensures the transactions from the same sender are included into blocks and executed
in sequential order. However it makes it tricky to send many transactions from the
same sender concurrently in a reliable way. IBC relayer and crypto exchanges are
typical examples of such use cases.

## Decision

We propose to add a boolean field `unordered` to transaction body to mark "un-ordered"
transactions.

Un-ordered transactions will bypass the nonce rules and follow the rules described
below instead, in contrary, the default ordered transactions are not impacted by
this proposal, they'll follow the nonce rules the same as before.

When an un-ordered transaction is included into a block, the transaction hash is
recorded in a dictionary. New transactions are checked against this dictionary for
duplicates, and to prevent the dictionary grow indefinitely, the transaction must
specify `timeout_height` for expiration, so it's safe to removed it from the
dictionary after it's expired.

The dictionary can be simply implemented as an in-memory golang map, a preliminary
analysis shows that the memory consumption won't be too big, for example `32M = 32 * 1024 * 1024`
can support 1024 blocks where each block contains 1024 unordered transactions. For
safety, we should limit the range of `timeout_height` to prevent very long expiration,
and limit the size of the dictionary.

### Transaction Format

```protobuf
message TxBody {
  ...

  bool unordered = 4; 
}
```

### `DedupTxHashManager`

```golang
const PurgeLoopSleepMS = 500

// DedupTxHashManager contains the tx hash dictionary for duplicates checking,
// and expire them when block number progresses.
type DedupTxHashManager struct {
  mutex sync.RWMutex
  // tx hash -> expire block number
  // for duplicates checking and expiration
  hashes map[TxHash]uint64
  // channel to receive latest block numbers
  blockCh chan uint64
}

func NewDedupTxHashManager() *DedupTxHashManager {
  m := &DedupTxHashManager{
    hashes: make(map[TxHash]uint64),
    blockCh: make(ch *uint64, 16),
  }
  go m.purgeLoop()
  return m
}

func (dtm *DedupTxHashManager) Close() error {
  close(dtm.blockCh)
  dtm.blockCh = nil
  return nil
}

func (dtm *DedupTxHashManager) Contains(hash TxHash) (ok bool) {
  dtm.mutex.RLock()
  defer dtm.mutex.RUnlock()

  _, ok = dtm.hashes[hash]
  return
}

func (dtm *DedupTxHashManager) Size() int {
  dtm.mutex.RLock()
  defer dtm.mutex.RUnlock()

  return len(dtm.hashes)
}

func (dtm *DedupTxHashManager) Add(hash TxHash, expire uint64) (ok bool) {
  dtm.mutex.Lock()
  defer dtm.mutex.Unlock()

  dtm.hashes[hash] = expire
  return
}

// OnNewBlock send the latest block number to the background purge loop,
// it should be called in abci commit event.
func (dtm *DedupTxHashManager) OnNewBlock(blockNumber uint64) {
  dtm.blockCh <- &blockNumber
}

// purgeLoop removes expired tx hashes at background
func (dtm *DedupTxHashManager) purgeLoop() error {
  for {
    blocks := channelBatchRecv(dtm.blockCh)
    if len(blocks) == 0 {
      // channel closed
      break
    }
    
    latest := *blocks[len(blocks)-1]
    hashes := dtm.expired(latest)
    if len(hashes) > 0 {
      dtm.purge(hashes)
    }
    
    // avoid burning cpu in catching up phase
    time.Sleep(PurgeLoopSleepMS * time.Millisecond)
  }
}

// expired find out expired tx hashes based on latest block number
func (dtm *DedupTxHashManager) expired(block uint64) []TxHash {
  dtm.mutex.RLock()
  defer dtm.mutex.RUnlock()

  var result []TxHash
  for h, expire := range dtm.hashes {
    if block > expire {
      result = append(result, h)
    }
  }
  return result
}

func (dtm *DedupTxHashManager) purge(hashes []TxHash) {
  dtm.mutex.Lock()
  defer dtm.mutex.Unlock()

  for _, hash := range hashes {
    delete(dtm.hashes, hash)
  }
}

// channelBatchRecv try to exhaust the channel buffer when it's not empty,
// and block when it's empty.
func channelBatchRecv[T any](ch <-chan *T) []*T {
	item := <-ch  // block if channel is empty
	if item == nil {
		// channel is closed
		return nil
	}

	remaining := len(ch)
	result := make([]*T, 0, remaining+1)
	result = append(result, item)
	for i := 0; i < remaining; i++ {
		result = append(result, <-ch)
	}

	return result
}
```

### Ante Handlers

Bypass the nonce decorator for un-ordered transactions.

```golang
func (isd IncrementSequenceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
  if tx.UnOrdered() {
    return next(ctx, tx, simulate)
  }

  // the previous logic
}
```

A decorator for the new logic.

```golang
type TxHash [32]byte

const (
  // MaxUnOrderedTTL defines the maximum ttl an un-order tx can set
  MaxUnOrderedTTL = 1024
)

type DedupTxDecorator struct {
  m *DedupTxHashManager
}

func (dtd *DedupTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
  // only apply to un-ordered transactions
  if !tx.UnOrdered() {
    return next(ctx, tx, simulate)
  }

  if tx.TimeoutHeight() == 0 {
    return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "unordered tx must set timeout-height")
  }

  if tx.TimeoutHeight() > ctx.BlockHeight() + MaxUnOrderedTTL {
    return nil, errorsmod.Wrapf(sdkerrors.ErrLogic, "unordered tx ttl exceeds %d", MaxUnOrderedTTL)
  }

  // check for duplicates
  if dtd.m.Contains(tx.Hash()) {
    return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "tx is duplicated")
  }

  if !ctx.IsCheckTx() {
    // a new tx included in the block, add the hash to the dictionary
    dtd.m.Add(tx.Hash(), tx.TimeoutHeight())
  }

  return next(ctx, tx, simulate)
}
```

### `OnNewBlock`

Wire the `OnNewBlock` method of `DedupTxHashManager` into the BaseApp's ABCI Commit event.

### Start Up

On start up, the node needs to re-fill the tx hash dictionary of `DedupTxHashManager`
by scanning `MaxUnOrderedTTL` number of historical blocks for existing un-expired
un-ordered transactions.

An alternative design is to store the tx hash dictionary in kv store, then no need
to warm up on start up.

## Consequences

### Positive

* Support un-ordered and concurrent transaction inclusion.

### Negative

* Start up overhead to scan historical blocks.

## References

* https://github.com/cosmos/cosmos-sdk/issues/13009
