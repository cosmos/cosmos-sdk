# Lockup Accounts


* [Lockup Account Types](#lockup-account-types)
    * [BaseLockup](#baselockup)
    * [ContinuousLockup](#continuouslockup)
    * [DelayedLockup](#delayedlockup)
    * [PeriodicLockup](#periodiclockup)
    * [PermanentLocked](#permanentlocked)
* [Genesis Initialization](#genesis-initialization)
* [In An Event Of Slashing](#in-an-event-of-slashing)
* [Examples](#examples)
    * [Simple](#simple)
    * [Slashing](#slashing)
    * [Periodic Lockup](#periodic-lockup)

The x/accounts/defaults/lockup module provides the implementation for lockup accounts within the x/accounts module.

## Lockup Account Types

### BaseLockup

The base lockup account is used by all default lockup accounts. It contains the basic information for a lockup account. The Base lockup account keeps knowledge of the staking delegations from the account.

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

### ContinuousLockup

The continuous lockup account has a future start time and begins unlocking continuously until the specified end date.

To determine the amount of coins that are vested for a given block time `T`, the
following is performed:

1. Compute `X := T - StartTime`
2. Compute `Y := EndTime - StartTime`
3. Compute `V' := OV * (X / Y)`
4. Compute `V := OV - V'`

Thus, the total amount of _vested_ coins is `V'` and the remaining amount, `V`,
is _lockup_.

```go
type ContinuousLockingAccount struct {
	*BaseLockup
	StartTime collections.Item[time.Time]
}
```

### DelayedLockup

The delayed lockup account unlocks all tokens at a specific time. The account can receive coins and send coins. The account can be used to lock coins for a long period of time.

```go
type DelayedLockingAccount struct {
	*BaseLockup
}
```

### PeriodicLockup

The periodic lockup account locks tokens for a series of periods. The account can receive coins and send coins. After all the periods, all the coins are unlocked and the account can send coins.

Periodic lockup accounts require calculating the coins released during each period for a given block time `T`. Note that multiple periods could have passed when calling `GetVestedCoins`, so we must iterate over each period until the end of that period is after `T`.

1. Set `CT := StartTime`
2. Set `V' := 0`

For each Period P:

  1. Compute `X := T - CT`
  2. IF `X >= P.Length`
      1. Compute `V' += P.Amount`
      2. Compute `CT += P.Length`
      3. ELSE break
  3. Compute `V := OV - V'`

```go
type PeriodicLockingAccount struct {
	*BaseLockup
	StartTime      collections.Item[time.Time]
	LockingPeriods collections.Vec[lockuptypes.Period]
}
```

### PermanentLocked

The permanent lockup account permanently locks the coins in the account. The account can only receive coins and cannot send coins. The account can be used to lock coins for a long period of time.

```go
type PermanentLockingAccount struct {
	*BaseLockup
}
```

## Genesis Initialization

<!-- TODO: once implemented -->

## In An Event Of Slashing

As defined, base lockup store `DelegatedLocking` by amount. In an event of a validator that the lockup account delegate to is slash which affect the actual delegation amount, this will leave the `DelegatedLocking` have an excess amount even if user undelegate all of the 
account delegated amount. This excess amount would affect the spendable amount, further details are as below:

The spendable amount is calculated as:
`spendableAmount` = `balance` - `notBondedLockedAmount`
where `notBondedLockedAmount` = `lockedAmount` - `Min(lockedAmount, delegatedLockedAmount)`

As seen in the formula `notBondedLockedAmount` can only be 0 or a positive value when `DelegatedLockedAmount` < `LockedAmount`. Let call `NewDelegatedLockedAmount` is the `delegatedLockedAmount` when applying N slash

1. Case 1: Originally `DelegatedLockedAmount` > `lockedAmount` but when applying the slash amount the `NewDelegatedLockedAmount` < `lockedAmount` then 
    * When not applying slash  `notBondedLockedAmount` will be 0 
    * When apply slash `notBondedLockedAmount` will be `lockedAmount` - `NewDelegatedLockedAmount` =  a positive amount
2. Case 2: where originally `DelegatedLockedAmount` < `lockedAmount` when applying the slash amount the `NewDelegatedLockedAmount` < `lockedAmount` then 
    * When not applying slash `lockedAmount` - `DelegatedLockedAmount`
    * When apply slash `notBondedLockedAmount` will be `lockedAmount` - `NewDelegatedLockedAmount` = `lockedAmount` - `(DelegatedLockedAmount - N)` = `lockedAmount` - `DelegatedLockedAmount` + N 
3. Case 3:  where originally `DelegatedLockedAmount` > `lockedAmount` when applying the slash amount still the `NewDelegatedLockedAmount` > `lockedAmount` then `notBondedLockedAmount` will be 0 applying slash or not

In cases 1 and 2, `notBondedLockedAmount` decreases when not applying the slash, resulting in a higher `spendableAmount`.

Due to the nature of x/accounts, as other modules cannot assume certain account types exist so the handling of slashing event must be done internally within x/accounts's accounts. For lockup accounts, this would make the logic overcomplicated. Since these effects are only an edge case that affect a small number of users, so here we would accept the trade off for a simpler design. This design decision aligns with the legacy vesting account implementation.

## Examples

### Simple

Given a continuous lockup account with 10 vested coins.

```text
OV = 10
DF = 0
DV = 0
BC = 10
V = 10
V' = 0
```

1. Immediately receives 1 coin

    ```text
    BC = 11
    ```

2. Time passes, 2 coins vest

    ```text
    V = 8
    V' = 2
    ```

3. Delegates 4 coins to validator A

    ```text
    DV = 4
    BC = 7
    ```

4. Sends 3 coins

    ```text
    BC = 4
    ```

5. More time passes, 2 more coins vest

    ```text
    V = 6
    V' = 4
    ```

6. Sends 2 coins. At this point, the account cannot send anymore until further
coins vest or it receives additional coins. It can still, however, delegate.

    ```text
    BC = 2
    ```

### Slashing

Same initial starting conditions as the simple example.

1. Time passes, 5 coins vest

    ```text
    V = 5
    V' = 5
    ```

2. Delegate 5 coins to validator A

    ```text
    DV = 5
    BC = 5
    ```

3. Delegate 5 coins to validator B

    ```text
    DF = 5
    BC = 0
    ```

4. Validator A gets slashed by 50%, making the delegation to A now worth 2.5 coins
5. Undelegate from validator A (2.5 coins)

    ```text
    DF = 5 - 2.5 = 2.5
    BC = 0 + 2.5 = 2.5
    ```

6. Undelegate from validator B (5 coins). The account at this point can only
send 2.5 coins unless it receives more coins or until more coins vest.
It can still, however, delegate.

    ```text
    DV = 5 - 2.5 = 2.5
    DF = 2.5 - 2.5 = 0
    BC = 2.5 + 5 = 7.5
    ```

    Notice how we have an excess amount of `DV`. This is explained in [In An Event Of Slashing](#in-an-event-of-slashing)

### Periodic Lockup

A lockup account is created where 100 tokens will be released over 1 year, with
1/4 of tokens vesting each quarter. The lockup schedule would be as follows:

```yaml
Periods:
- amount: 25stake, length: 7884000
- amount: 25stake, length: 7884000
- amount: 25stake, length: 7884000
- amount: 25stake, length: 7884000
```

```text
OV = 100
DF = 0
DV = 0
BC = 100
V = 100
V' = 0
```

1. Immediately receives 1 coin

    ```text
    BC = 101
    ```

2. Lockup period 1 passes, 25 coins vest

    ```text
    V = 75
    V' = 25
    ```

3. During lockup period 2, 5 coins are transferred and 5 coins are delegated

    ```text
    DV = 5
    BC = 91
    ```

4. Lockup period 2 passes, 25 coins vest

    ```text
    V = 50
    V' = 50
    ```
