# LockUp


* [Vesting Account Types](#vesting-account-types)
    * [BaseVestingAccount](#basevestingaccount)
    * [ContinuousVestingAccount](#continuousvestingaccount)
    * [DelayedVestingAccount](#delayedvestingaccount)
    * [Period](#period)
    * [PeriodicVestingAccount](#periodicvestingaccount)
    * [PermanentLockedAccount](#permanentlockedaccount)

The x/accounts/lockup module provides the implementation for lockup accounts within the x/accounts module.

## Vesting Account Types

### BaseVestingAccount

The base vesting account is used by all default lockup accounts. It contains the basic information for a vesting account. The Base vesting account keeps knowledge of the staking delegations from the  account.

```go
type BaseLockup struct {
	// Owner is the address of the account owner.
	Owner            collections.Item[[]byte]
	OriginalLocking  collections.Map[string, math.Int]
	DelegatedFree    collections.Map[string, math.Int]
	DelegatedLocking collections.Map[string, math.Int]
	WithdrawedCoins  collections.Map[string, math.Int]
	addressCodec     address.Codec
	headerService    header.Service
	// lockup end time.
	EndTime collections.Item[time.Time]
}
```

### ContinuousVestingAccount

The continuous vesting account has a future start time and begins unlocking continuously until the specified enddate.

```go
type ContinuousLockingAccount struct {
	*BaseLockup
	StartTime collections.Item[time.Time]
}
```

### DelayedVestingAccount

The delayed vesting account unlocks all tokens at a specific time. The account can receive coins and send coins. The account can be used to lock coins for a long period of time.

```go
type DelayedLockingAccount struct {
	*BaseLockup
}
```

### PeriodicVestingAccount

The periodic vesting account locks tokens for a series of periods. The account can receive coins and send coins. After all the periods all the coins are unlocked and the account can send coins.

```go
type PeriodicLockingAccount struct {
	*BaseLockup
	StartTime      collections.Item[time.Time]
	LockingPeriods collections.Vec[lockuptypes.Period]
}
```

### PermanentLockedAccount

The permanent vesting account permentally locks the coins in the account. The account can only receive coins and cannot send coins. The account can be used to lock coins for a long period of time.

```go
type PermanentLockingAccount struct {
	*BaseLockup
}
```
