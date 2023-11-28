# ADR 069: Un-Ordered Nonce Lane

## Changelog

* Nov 24, 2023: Initial Draft

## Status

Proposed

## Abstract

Add an extra nonce "lane" to support un-ordered(concurrent) transaction inclusion.

## Context

Right now the nonce value (account sequence number) prevents replay-attack and make sure the transactions from the same sender are included into blocks and executed in order. But it makes it tricky to send many transactions concurrently in a reliable way. IBC relayer and crypto exchanges would be typical examples of such use cases.

## Decision

Add an extra nonce lane to support optional un-ordered transaction inclusion, the default lane provides ordered semantic, the same as current behavior, the new one provides the un-ordered semantic. The transaction can choose which lane to use.

One of the design goals is to keep minimal overhead and breakage to the existing users who don't use the new feature.

### Transaction Format

It doesn't change the transaction format itself, but re-use the high bit of the exisitng 64bits nonce value to identify the lane, `0` being the default ordered lane, `1` being the new unordered lane.

### Account State

The new nonce lane needs to add some optional fields to the account state:

```golang
type UnOrderedNonce struct {
  Sequence uint64
  Gaps IntSet
}

type Account struct {
  // default to `nil`, only initialized when the new feature is first used.
  unorderedNonce *UnOrderedNonce
}
```

The un-ordered nonce state includes a normal sequence value plus the gap values in recent history, the gap set has a maximum capacity to limit the resource usage, when the capacity is reached, the oldest gap value is simply dropped, which means the pending transaction with that value as nonce will not be accepted anymore.

### Expiration

It would be good to expire the gap nonces after certain timeout is reached, to mitigate the risk that a middleman  intercept an old transaction re-execute it in a long future, which might cause unexpected result.

### Prototype Implementation

The prototype implementation use a roaring bitmap to record these gap values.

```golang
// MaxGap is the maximum gaps a new nonce value can introduce
const MaxGap = 1024

// CheckNonce checks if the nonce in tx is valid, if yes, also update the internal state.
func(u *UnOrderedNonce) CheckNonce(nonce int64) error {
  switch {
    case nonce == u.Sequence:
      // special case, the current sequence number must have been occupied
    	return errors.New("nonce is occupied")
    case nonce >= u.Sequence + 1:
      // the number of gaps introduced by this nonce value, could be zero if it happens to be `u.Sequence + 1`
      gaps := nonce - u.Sequence - 1
      if gaps > MaxGap {
        return errors.New("max gap is exceeded")
      }
      // record the gaps into the bitmap
      for i := 0; i < gaps; i++ {
        acct.Gaps.Add(i + u.Sequence + 1)
      }
      // record the latest nonce
    	u.Sequence = nonce
    case nonce < u.Sequence:
      // try to use a gap value
    	if !u.Gaps.Contains(nonce) {
        return errors.New("nonce is occupied")
    	}
    	u.Gaps.Remove(nonce)
  }
  return nil
}

// IntSet is a set of integers with a capacity, when capacity reached, drop the smallest value
type IntSet struct {
  capacity int
  bitmap roaringbitmap.BitMap
}

func NewIntSet(capacity int) *IntSet {
  return IntSet{
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
    // drop the smallest ones
    toDrop := is.bitmap.GetCardinality() + n - is.capacity
    for i:= 0; i<toDrop; i++ {
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
* Don't need to change transaction format.

### Negative

- Some runtime overhead when the new feature is used.

## References

* https://github.com/cosmos/cosmos-sdk/issues/13009
