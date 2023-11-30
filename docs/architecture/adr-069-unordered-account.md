# ADR 069: Un-Ordered Nonce Lane

## Changelog

* Nov 24, 2023: Initial Draft

## Status

Proposed

## Abstract

We propose to add an extra nonce "lane" to support un-ordered (concurrent) transaction inclusion.

## Context

As of today, the nonce value (account sequence number) prevents replay-attack and ensures the transactions from the same sender are included into blocks and executed in sequential order. However it makes it tricky to send many transactions concurrently in a reliable way. IBC relayer and crypto exchanges would be typical examples of such use cases.

## Decision

Add an extra un-ordered nonce lane to support un-ordered transaction inclusion, it works at the same time as the default ordered lane, the transaction builder can choose either lane to use.

The un-ordered nonce lane accepts nonce values that is greater than the current sequence number plus 1, which effectively creates "gaps" of nonce values, those gap values are tracked in the account state, and can be filled by future transactions, thus allows transactions to be executed in a un-ordered way.

It also tracks the last block time when the latest nonce is updated, and expire the gaps certain timeout reached, to mitigate the risk that a middleman intercept an old transaction and re-execute it in a longer future, which might cause unexpected result.

The design introducs almost zero overhead for users who don't use the new feature.

### Transaction Format

Add a boolean field `unordered` to the `BaseAccount` message, when it's set to `true`, the transaction will use the un-ordered lane, otherwise the default ordered lane.

```protobuf
message BaseAccount {
  ...
  uint64 account_number = 3;
  uint64 sequence       = 4;
  boolean unordered     = 5;
}
```

### Account State

Add some optional fields to the account state:

```golang
type UnorderedNonceManager struct {
  Sequence uint64
  Timestamp uint64  // the block time when the Sequence is updated
  Gaps IntSet
}

type Account struct {
  // default to `nil`, only initialized when the new feature is first used.
  unorderedNonceManager *UnorderedNonceManager
}
```

The un-ordered nonce state includes a normal sequence value plus the set of unused(gap) values in recent history, these recorded gap values can be reused by future transactions, after used they are removed from the set and can't be used again, the gap set has a maximum capacity to limit the resource usage, when the capacity is reached, the oldest gap value is removed, which also makes the pending transaction using that value as nonce will not be accepted anymore.

### Nonce Validation Logic

The prototype implementation use a roaring bitmap to record these gap values, where the set bits represents the the gaps.

```golang
// CheckNonce switches to un-ordered lane if the MSB of the nonce is set.
func (acct *Account) CheckNonce(nonce uint64, unordered bool, blockTime uint64) error {
  if unordered {
    if acct.unorderedNonceManager == nil {
      acct.unorderedNonceManager = NewUnorderedNonceManager()
    }
    return acct.unorderedNonceManager.CheckNonce(nonce, blockTime)
  }

  // current ordered nonce logic
}
```

### UnorderedNonceManager

```golang
const (
  // GapsCapacity is the capacity of the set of gap values.
  GapsCapacity = 1024
  // MaxGap is the maximum gaps a new nonce value can introduce
  MaxGap = 1024
  // GapsExpirationDuration is the duration in seconds for the gaps to expire
  GapsExpirationDuration = 60 * 60 * 24
)

// getGaps returns the gap set, or create a new one if it's expired
func(unm *UnorderedNonceManager) getGaps(blockTime uint64) *IntSet {
  if blockTime > unm.Timestamp + GapsExpirationDuration {
    return NewIntSet(GapsCapacity)
  }

  return &unm.Gaps
}

// CheckNonce checks if the nonce in tx is valid, if yes, also update the internal state.
func(unm *UnorderedNonceManager) CheckNonce(nonce uint64, blockTime uint64) error {
  switch {
    case nonce == unm.Sequence:
      // special case, the current sequence number must have been occupied
      return errors.New("nonce is occupied")

    case nonce > unm.Sequence:
      // the number of gaps introduced by this nonce value, could be zero if it happens to be `unm.Sequence + 1`
      gaps := nonce - unm.Sequence - 1
      if gaps > MaxGap {
        return errors.New("max gap is exceeded")
      }

      gapSet := unm.getGaps(blockTime)
      // record the gaps into the bitmap
      gapSet.AddRange(unm.Sequence + 1, unm.Sequence + gaps + 1)

      // update the latest nonce
      unm.Gaps = *gapSet
      unm.Sequence = nonce
      unm.Timestamp = blockTime

    default:
      // `nonce < unm.Sequence`, the tx try to use a historical nonce
      gapSet := acct.getGaps(blockTime)
      if !gapSet.Contains(nonce) {
        return errors.New("nonce is occupied or expired")
      }

      gapSet.Remove(nonce)
      unm.Gaps = *gapSet
  }
  return nil
}

// IntSet is a set of integers with a capacity, when capacity reached, drop the smallest value
type IntSet struct {
  capacity int
  bitmap roaringbitmap.BitMap
}

func NewIntSet(capacity int) *IntSet {
  return &IntSet{
    capacity: capacity,
    bitmap: *roaringbitmap.New(),
  }
}

func (is *IntSet) Add(n uint64) {
  if is.bitmap.GetCardinality() >= is.capacity {
    // drop the smallest one
    is.bitmap.Remove(is.bitmap.Minimal())
  }

  is.bitmap.Add(n)
}

// AddRange adds the integers in [rangeStart, rangeEnd) to the bitmap.
func (is *IntSet) AddRange(start, end uint64) {
  n := end - start
  if is.bitmap.GetCardinality() + n > is.capacity {
    // drop the smallest ones until the capacity is not exceeded
    toDrop := is.bitmap.GetCardinality() + n - is.capacity
    for i := uint64(0); i < toDrop; i++ {
      is.bitmap.Remove(is.bitmap.Minimal())
    }
  }

  is.bitmap.AddRange(start, end)
}

func (is *IntSet) Remove(n uint64) {
  is.bitmap.Remove(n)
}

func (is *IntSet) Contains(n uint64) bool {
  return is.bitmap.Contains(n)
}
```

## Consequences

### Positive

* Support concurrent transaction inclusion.
* Only optional fields are added to account state, no state migration is needed.
* No runtime overhead when the new feature is not used.

### Negative

- Some runtime overhead when the new feature is used.

## References

* https://github.com/cosmos/cosmos-sdk/issues/13009
