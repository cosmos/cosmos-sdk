# ADR 070: Un-Ordered Transaction Inclusion

## Changelog

* Dec 4, 2023: Initial Draft

## Status

Proposed

## Abstract

We propose a way to do replay-attack protection without enforcing the order of transactions, it don't use nonce at all. In this way we can support un-ordered transaction inclusion.

## Context

As of today, the nonce value (account sequence number) prevents replay-attack and ensures the transactions from the same sender are included into blocks and executed in sequential order. However it makes it tricky to send many transactions from the same sender concurrently in a reliable way. IBC relayer and crypto exchanges are typical examples of such use cases.

## Decision

We propose to add a boolean field `unordered` to transaction body to mark "un-ordered" transactions.

Un-ordered transactions will bypass the nonce rules and follow the rules described below instead, in contrary, the default transactions are not impacted by this proposal, they'll follow the nonce rules the same as before.

When an un-ordered transaction are included into block, the transaction hash is recorded in a dictionary, new transactions are checked against this dictionary for duplicates, and to prevent the dictionary grow indefinitly, the transaction must specify `timeout_height` for expiration, so it's safe to removed it from the dictionary after it's expired.

The dictionary can be simply implemented as an in-memory golang map, a preliminary analysis shows that the memory consumption won't be too big, for example `32M = 32 * 1024 * 1024` can support 1024 blocks where each block contains 1024 unordered transactions. For safty, we should limit the range of `timeout_height` to prevent very long expiration, and limit the size of the dictionary.

### Transaction Format

```protobuf
message TxBody {
  ...

  boolean unordered = 4; 
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
  // MaxNumberOfTxHash * 32 = 128M max memory usage
  MaxNumberOfTxHash = 1024 * 1024 * 4

  // MaxUnOrderedTTL defines the maximum ttl an un-order tx can set
  MaxUnOrderedTTL = 1024
)

type DedupTxDecorator struct {
  hashes map[TxHash]struct{}
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

  if !ctx.IsCheckTx() {
    // a new tx included in the block, add the hash to the dictionary
    if len(dtd.hashes) >= MaxNumberOfTxHash {
      return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "dedup map is full")
    }
    dtd.hashes[tx.Hash()] = struct{}
  } else {
    // check for duplicates
    if _, ok := dtd.hashes[tx.Hash()]; ok {
      return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "tx is dupliated")
    }
  }

  return next(ctx, tx, simulate)
}
```

### Start Up

On start up, the node needs to re-fill the tx hash dictionary by scanning `MaxUnOrderedTTL` number of historical blocks for un-ordered transactions.

An alternative design is to store the tx hash dictionary in kv store, then no need to warm up on start up.

## Consequences

### Positive

* Support un-ordered and concurrent transaction inclusion.

### Negative

- Start up overhead to scan historical blocks.

## References

* https://github.com/cosmos/cosmos-sdk/issues/13009
