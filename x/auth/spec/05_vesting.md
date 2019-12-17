<!--
order: 6
-->

# Vesting

- [Vesting](#vesting)
  - [Intro and Requirements](#intro-and-requirements)
  - [Vesting Account Types](#vesting-account-types)
  - [Vesting Account Specification](#vesting-account-specification)
    - [Determining Vesting & Vested Amounts](#determining-vesting--vested-amounts)
      - [Continuously Vesting Accounts](#continuously-vesting-accounts)
      - [Delayed/Discrete Vesting Accounts](#delayeddiscrete-vesting-accounts)
    - [Transferring/Sending](#transferringsending)
      - [Keepers/Handlers](#keepershandlers)
    - [Delegating](#delegating)
      - [Keepers/Handlers](#keepershandlers-1)
    - [Undelegating](#undelegating)
      - [Keepers/Handlers](#keepershandlers-2)
  - [Keepers & Handlers](#keepers--handlers)
  - [Genesis Initialization](#genesis-initialization)
  - [Examples](#examples)
    - [Simple](#simple)
    - [Slashing](#slashing)
  - [Glossary](#glossary)

## Intro and Requirements

This specification defines the vesting account implementation that is used by
the Cosmos Hub. The requirements for this vesting account is that it should be
initialized during genesis with a starting balance `X` and a vesting end
time `ET`. A vesting account may be initialized with a vesting start time `ST`
and a number of vesting periods `P`. If a vesting start time is included, the
vesting period will not begin until start time is reached. If vesting periods
are included, the vesting will occur over the specified number of periods.

For all vesting accounts, the owner of the vesting account is able to delegate
and undelegate from validators, however they cannot transfer coins to another
account until those coins are vested. This specification allows for three
different kinds of vesting:

- Delayed vesting, where all coins are vested once `ET` is reached.
- Continous vesting, where coins begin to vest at `ST` and vest linearly with
respect to time until `ET` is reached
- Periodic vesting, where coins begin to vest at `ST` and vest periodically
according to number of periods and the vesting amount per period.
The number of periods, length per period, and amount per period are
configurable. A periodic vesting account is distinguished from a continuous
vesting account in that coins can be released in staggered tranches. For
example, a periodic vesting account could be used for vesting arrangements
where coins are relased quarterly, yearly, or over any other function of
tokens over time.

## Note

Vesting accounts can be initialized with some vesting and non-vesting coins.
The non-vesting coins would be immediately transferable. The current
specification does not allow for vesting accounts to be created with normal
messages after genesis. All vesting accounts must be created at genesis, or as
part of a manual network upgrade. The current specification only allows
for _unconditional_ vesting (ie. there is no possibility of reaching `ET` and
having coins fail to vest).

## Vesting Account Types

```go
// VestingAccount defines an interface that any vesting account type must
// implement.
type VestingAccount interface {
    Account

    GetVestedCoins(Time)  Coins
    GetVestingCoins(Time) Coins

    // Delegation and undelegation accounting that returns the resulting base
    // coins amount.
    TrackDelegation(Time, Coins)
    TrackUndelegation(Coins)

    GetStartTime() int64
    GetEndTime()   int64
}

// BaseVestingAccount implements the VestingAccount interface. It contains all
// the necessary fields needed for any vesting account implementation.
type BaseVestingAccount struct {
    BaseAccount

    OriginalVesting  Coins // coins in account upon initialization
    DelegatedFree    Coins // coins that are vested and delegated
    DelegatedVesting Coins // coins that vesting and delegated

    EndTime  int64 // when the coins become unlocked
}

// ContinuousVestingAccount implements the VestingAccount interface. It
// continuously vests by unlocking coins linearly with respect to time.
type ContinuousVestingAccount struct {
    BaseVestingAccount

    StartTime  int64 // when the coins start to vest
}

// DelayedVestingAccount implements the VestingAccount interface. It vests all
// coins after a specific time, but non prior. In other words, it keeps them
// locked until a specified time.
type DelayedVestingAccount struct {
    BaseVestingAccount
}

// VestingPeriod defines a length of time and amount of coins that will vest
type Period struct {
  Length int64 // length of the period, in seconds
  Amount Coins // amount of coins vesting during this period
}

// Stores all vesting periods passed as part of a PeriodicVestingAccount
type Periods []Period

// PeriodicVestingAccount implements the VestingAccount interface. It
// periodically vests by unlocking coins during each specified period
type PeriodicVestingAccount struct {
  BaseVestingAccount
  StartTime int64
  Periods Periods // the vesting schedule
}
```

In order to facilitate less ad-hoc type checking and assertions and to support
flexibility in account usage, the existing `Account` interface is updated to contain
the following:

```go
type Account interface {
    // ...

    // Calculates the amount of coins that can be sent to other accounts given
    // the current time.
    SpendableCoins(Time) Coins
}
```

## Vesting Account Specification

Given a vesting account, we define the following in the proceeding operations:

- `OV`: The original vesting coin amount. It is a constant value.
- `V`: The number of `OV` coins that are still _vesting_. It is derived by
`OV`, `StartTime` and `EndTime`. This value is computed on demand and not on a
per-block basis.
- `V'`: The number of `OV` coins that are _vested_ (unlocked). This value is
computed on demand and not a per-block basis.
- `DV`: The number of delegated _vesting_ coins. It is a variable value. It is
stored and modified directly in the vesting account.
- `DF`: The number of delegated _vested_ (unlocked) coins. It is a variable
value. It is stored and modified directly in the vesting account.
- `BC`: The number of `OV` coins less any coins that are transferred
(which can be negative or delegated). It is considered to be balance of the
embedded base account. It is stored and modified directly in the vesting account.

### Determining Vesting & Vested Amounts

It is important to note that these values are computed on demand and not on a
mandatory per-block basis (e.g. `BeginBlocker` or `EndBlocker`).

#### Continuously Vesting Accounts

To determine the amount of coins that are vested for a given block time `T`, the
following is performed:

1. Compute `X := T - StartTime`
2. Compute `Y := EndTime - StartTime`
3. Compute `V' := OV * (X / Y)`
4. Compute `V := OV - V'`

Thus, the total amount of _vested_ coins is `V'` and the remaining amount, `V`,
is _vesting_.

```go
func (cva ContinuousVestingAccount) GetVestedCoins(t Time) Coins {
    if t <= cva.StartTime {
        // We must handle the case where the start time for a vesting account has
        // been set into the future or when the start of the chain is not exactly
        // known.
        return ZeroCoins
    } else if t >= cva.EndTime {
        return cva.OriginalVesting
    }

    x := t - cva.StartTime
    y := cva.EndTime - cva.StartTime

    return cva.OriginalVesting * (x / y)
}

func (cva ContinuousVestingAccount) GetVestingCoins(t Time) Coins {
    return cva.OriginalVesting - cva.GetVestedCoins(t)
}
```

### Periodic Vesting Accounts

Periodic vesting accounts require calculating the coins released during each
period for a given block time `T`. Note that multiple periods could have passed
when calling `GetVestedCoins`, so we must iterate over each period until the
end of that period is after `T`.

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
func (pva PeriodicVestingAccount) GetVestedCoins(t Time) Coins {
  if t < pva.StartTime {
    return ZeroCoins
  }
  ct := pva.StartTime // The start of the vesting schedule
  vested := 0
  periods = pva.GetPeriods()
  for _, period  := range periods {
    if t - ct < period.Length {
      break
    }
    vested += period.Amount
    ct += period.Length // increment ct to the start of the next vesting period
  }
  return vested
}

func (pva PeriodicVestingAccount) GetVestingCoins(t Time) Coins {
    return pva.OriginalVesting - cva.GetVestedCoins(t)
}
```

#### Delayed/Discrete Vesting Accounts

Delayed vesting accounts are easier to reason about as they only have the full
amount vesting up until a certain time, then all the coins become vested (unlocked).
This does not include any unlocked coins the account may have initially.

```go
func (dva DelayedVestingAccount) GetVestedCoins(t Time) Coins {
    if t >= dva.EndTime {
        return dva.OriginalVesting
    }

    return ZeroCoins
}

func (dva DelayedVestingAccount) GetVestingCoins(t Time) Coins {
    return dva.OriginalVesting - dva.GetVestedCoins(t)
}
```

### Transferring/Sending

At any given time, a vesting account may transfer: `min((BC + DV) - V, BC)`.

In other words, a vesting account may transfer the minimum of the base account
balance and the base account balance plus the number of currently delegated
vesting coins less the number of coins vested so far.

```go
func (va VestingAccount) SpendableCoins(t Time) Coins {
    bc := va.GetCoins()
    return min((bc + va.DelegatedVesting) - va.GetVestingCoins(t), bc)
}
```

#### Keepers/Handlers

The corresponding `x/bank` keeper should appropriately handle sending coins
based on if the account is a vesting account or not.

```go
func SendCoins(t Time, from Account, to Account, amount Coins) {
    bc := from.GetCoins()

    if isVesting(from) {
        sc := from.SpendableCoins(t)
        assert(amount <= sc)
    }

    newCoins := bc - amount
    assert(newCoins >= 0)

    from.SetCoins(bc - amount)
    to.SetCoins(amount)

    // save accounts...
}
```

### Delegating

For a vesting account attempting to delegate `D` coins, the following is performed:

1. Verify `BC >= D > 0`
2. Compute `X := min(max(V - DV, 0), D)` (portion of `D` that is vesting)
3. Compute `Y := D - X` (portion of `D` that is free)
4. Set `DV += X`
5. Set `DF += Y`

```go
func (va VestingAccount) TrackDelegation(t Time, amount Coins) {
    x := min(max(va.GetVestingCoins(t) - va.DelegatedVesting, 0), amount)
    y := amount - x

    va.DelegatedVesting += x
    va.DelegatedFree += y
}
```

**Note** `TrackDelegation` only modifies the `DelegatedVesting` and `DelegatedFree`
fields, so upstream callers MUST modify the `Coins` field by subtracting `amount`.

#### Keepers/Handlers

```go
func DelegateCoins(t Time, from Account, amount Coins) {
    bc := from.GetCoins()
    assert(amount <= bc)

    if isVesting(from) {
        from.TrackDelegation(t, amount)
    } else {
        from.SetCoins(sc - amount)
    }

    // save account...
}
```

### Undelegating

For a vesting account attempting to undelegate `D` coins, the following is performed:
NOTE: `DV < D` and `(DV + DF) < D` may be possible due to quirks in the rounding of
delegation/undelegation logic.

1. Verify `D > 0`
2. Compute `X := min(DF, D)` (portion of `D` that should become free, prioritizing free coins)
3. Compute `Y := min(DV, D - X)` (portion of `D` that should remain vesting)
4. Set `DF -= X`
5. Set `DV -= Y`

```go
func (cva ContinuousVestingAccount) TrackUndelegation(amount Coins) {
    x := min(cva.DelegatedFree, amount)
    y := amount - x

    cva.DelegatedFree -= x
    cva.DelegatedVesting -= y
}
```

**Note** `TrackUnDelegation` only modifies the `DelegatedVesting` and `DelegatedFree`
fields, so upstream callers MUST modify the `Coins` field by adding `amount`.

**Note**: If a delegation is slashed, the continuous vesting account will end up
with an excess `DV` amount, even after all its coins have vested. This is because
undelegating free coins are prioritized.

**Note**: The undelegation (bond refund) amount may exceed the delegated
vesting (bond) amount due to the way undelegation truncates the bond refund,
which can increase the validator's exchange rate (tokens/shares) slightly if the
undelegated tokens are non-integral.

#### Keepers/Handlers

```go
func UndelegateCoins(to Account, amount Coins) {
    if isVesting(to) {
        if to.DelegatedFree + to.DelegatedVesting >= amount {
            to.TrackUndelegation(amount)
            // save account ...
        }
    } else {
        AddCoins(to, amount)
        // save account...
    }
}
```

## Keepers & Handlers

The `VestingAccount` implementations reside in `x/auth`. However, any keeper in
a module (e.g. staking in `x/staking`) wishing to potentially utilize any vesting
coins, must call explicit methods on the `x/bank` keeper (e.g. `DelegateCoins`)
opposed to `SendCoins` and `SubtractCoins`.

In addition, the vesting account should also be able to spend any coins it
receives from other users. Thus, the bank module's `MsgSend` handler should
error if a vesting account is trying to send an amount that exceeds their
unlocked coin amount.

See the above specification for full implementation details.

## Genesis Initialization

To initialize both vesting and non-vesting accounts, the `GenesisAccount` struct will
include new fields: `Vesting`, `StartTime`, and `EndTime`. Accounts meant to be
of type `BaseAccount` or any non-vesting type will have `Vesting = false`. The
genesis initialization logic (e.g. `initFromGenesisState`) will have to parse
and return the correct accounts accordingly based off of these new fields.

```go
type GenesisAccount struct {
    // ...

    // vesting account fields
    OriginalVesting  sdk.Coins `json:"original_vesting"`
    DelegatedFree    sdk.Coins `json:"delegated_free"`
    DelegatedVesting sdk.Coins `json:"delegated_vesting"`
    StartTime        int64     `json:"start_time"`
    EndTime          int64     `json:"end_time"`
}

func ToAccount(gacc GenesisAccount) Account {
    bacc := NewBaseAccount(gacc)

    if gacc.OriginalVesting > 0 {
        if ga.StartTime != 0 && ga.EndTime != 0 {
            // return a continuous vesting account
        } else if ga.EndTime != 0 {
            // return a delayed vesting account
        } else {
            // invalid genesis vesting account provided
            panic()
        }
    }

    return bacc
}
```

## Examples

### Simple

Given a continuous vesting account with 10 vesting coins.

```
OV = 10
DF = 0
DV = 0
BC = 10
V = 10
V' = 0
```

1. Immediately receives 1 coin

    ```
    BC = 11
    ```

2. Time passes, 2 coins vest

    ```
    V = 8
    V' = 2
    ```

3. Delegates 4 coins to validator A

    ```
    DV = 4
    BC = 7
    ```

4. Sends 3 coins

    ```
    BC = 4
    ```

5. More time passes, 2 more coins vest

    ```
    V = 6
    V' = 4
    ```

6. Sends 2 coins. At this point the account cannot send anymore until further
coins vest or it receives additional coins. It can still however, delegate.

    ```
    BC = 2
    ```

### Slashing

Same initial starting conditions as the simple example.

1. Time passes, 5 coins vest

    ```
    V = 5
    V' = 5
    ```

2. Delegate 5 coins to validator A

    ```
    DV = 5
    BC = 5
    ```

3. Delegate 5 coins to validator B

    ```
    DF = 5
    BC = 0
    ```

4. Validator A gets slashed by 50%, making the delegation to A now worth 2.5 coins
5. Undelegate from validator A (2.5 coins)

    ```
    DF = 5 - 2.5 = 2.5
    BC = 0 + 2.5 = 2.5
    ```

6. Undelegate from validator B (5 coins). The account at this point can only
send 2.5 coins unless it receives more coins or until more coins vest.
It can still however, delegate.

    ```
    DV = 5 - 2.5 = 2.5
    DF = 2.5 - 2.5 = 0
    BC = 2.5 + 5 = 7.5
    ```

    Notice how we have an excess amount of `DV`.

### Periodic Vesting

A vesting account is created where 100 tokens will be released over 1 year, with
1/4 of tokens vesting each quarter. The vesting schedule would be as follows:

```yaml
Periods:
- amount: 25stake, length: 7884000
- amount: 25stake, length: 7884000
- amount: 25stake, length: 7884000
- amount: 25stake, length: 7884000
```

```
OV = 100
DF = 0
DV = 0
BC = 100
V = 100
V' = 0
```

1. Immediately receives 1 coin

    ```
    BC = 101
    ```

2. Vesting period 1 passes, 25 coins vest

    ```
    V = 75
    V' = 25
    ```

3. During vesting period 2, 5 coins are transfered and 5 coins are delegated

    ```
    DV = 5
    BC = 91
    ```

4. Vesting period 2 passes, 25 coins vest

    ```
    V = 50
    V' = 50
    ```

## Glossary

- OriginalVesting: The amount of coins (per denomination) that are initially
part of a vesting account. These coins are set at genesis.
- StartTime: The BFT time at which a vesting account starts to vest.
- EndTime: The BFT time at which a vesting account is fully vested.
- DelegatedFree: The tracked amount of coins (per denomination) that are
delegated from a vesting account that have been fully vested at time of delegation.
- DelegatedVesting: The tracked amount of coins (per denomination) that are
delegated from a vesting account that were vesting at time of delegation.
- ContinuousVestingAccount: A vesting account implementation that vests coins
linearly over time.
- DelayedVestingAccount: A vesting account implementation that only fully vests
all coins at a given time.
- PeriodicVestingAccount: A vesting account implementation that vests coins
according to a custom vesting schedule.
