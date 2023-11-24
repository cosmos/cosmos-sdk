# ADR 069: Un-Ordered Account

## Changelog

* Nov 24, 2023: Initial Draft

## Abstract

Support gaps in account nonce values to support un-ordered(concurrent) transaction inclusion.

## Context

Right now the transactions from a particular sender must be included in strict order because of the nonce value requirement, which makes it tricky to send many transactions concurrently, for example, when a previous pending transaction fails or timeouts, all the following transactions are all blocked. Relayer and exchanges are some typical examples.

The main purpose of nonce value is to protect against replay attack, but for that purpose we only need to make sure the nonce is unique, so we can relax the orderness requirement without lose the uniqueness, in that way we can improve the user experience of concurrent transaction sending.

## Decision

Change the nonce logic to allow gaps to exist, which can be filled by other transactions later, or never filled at all, the prototype implementation use a bitmap to record these gap values.

It's debatable how we should configure the user accounts, for example, should we change the default behavior directly or let user to turn this feature on explicitly, or should we allow user to set different gap capacity for different accounts.

```golang
const MaxGap = 1024

  type Account struct {
    ...
    SequenceNumber int64
+   Gaps *IntSet
  }

// CheckNonce checks if the input nonce is valid, if yes, modify internal state.
func(acc *Account) CheckNonce(nonce int64) error {
  switch {
    case nonce == acct.SequenceNumber:
    	return errors.New("nonce is occupied")
    case nonce >= acct.SequenceNumber + 1:
      gaps := nonce - acct.SequenceNumber - 1
      if gaps > MaxGap {
        return errors.New("max gap is exceeded")
      }
    	for i := 0; i < gaps; i++ {
      	acct.Gaps.Add(i + acct.SequenceNumber + 1)
    	}
    	acct.SequenceNumber = nonce
    case nonce < acct.SequenceNumber:
    	if !acct.Gaps.Contains(nonce) {
      	return errors.New("nonce is occupied")
    	}
    	acct.Gaps.Remove(nonce)
  }
  return nil
}
```

Prototype implementation of `IntSet`:

```golang
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

func (is *IntSet) Add(n int) {
  if is.bitmap.Length() >= is.capacity {
    // pop the minimal one
    is.Remove(is.bitmap.Minimal())
  }
  
  is.bitmap.Add(n)
}

func (is *IntSet) Remove(n int) {
  is.bitmap.Remove(n)
}

func (is *IntSet) Contains(n int) bool {
  return is.bitmap.Contains(n)
}
```

## Status

Proposed.

## Consequences

### Positive

* Only optional fields are added to `Account`, migration is easy.

### Negative

- Limited runtime overhead.

## References

* https://github.com/cosmos/cosmos-sdk/issues/13009
