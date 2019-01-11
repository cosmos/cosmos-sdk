# Vesting

<!-- TOC -->

- [Vesting](#vesting)
  - [Intro and Requirements](#intro-and-requirements)
  - [Vesting Account Types](#vesting-account-types)
  - [Vesting Account Specification](#vesting-account-specification)
    - [Determining Vesting & Vested Amounts](#determining-vesting--vested-amounts)
      - [Continuously Vesting Accounts](#continuously-vesting-accounts)
      - [Delayed/Discrete Vesting Accounts](#delayeddiscrete-vesting-accounts)
    - [Transferring/Sending](#transferringsending)
      - [Continuously Vesting Accounts](#continuously-vesting-accounts-1)
        - [Delayed/Discrete Vesting Accounts](#delayeddiscrete-vesting-accounts-1)
        - [Keepers/Handlers](#keepershandlers)
    - [Delegating](#delegating)
      - [Continuously Vesting Accounts](#continuously-vesting-accounts-2)
        - [Delayed/Discrete Vesting Accounts](#delayeddiscrete-vesting-accounts-2)
        - [Keepers/Handlers](#keepershandlers-1)
    - [Undelegating](#undelegating)
      - [Continuously Vesting Accounts](#continuously-vesting-accounts-3)
        - [Delayed/Discrete Vesting Accounts](#delayeddiscrete-vesting-accounts-3)
        - [Keepers/Handlers](#keepershandlers-2)
  - [Keepers & Handlers](#keepers--handlers)
  - [Initializing at Genesis](#initializing-at-genesis)
  - [Examples](#examples)
    - [Simple](#simple)
    - [Slashing](#slashing)
  - [Glossary](#glossary)

<!-- /TOC -->

## Intro and Requirements

This paper specifies vesting account implementation for the Cosmos Hub.
The requirements for this vesting account is that it should be initialized
during genesis with a starting balance `X` coins and a vesting end time `T`.

The owner of this account should be able to delegate to validators
and vote with locked coins, however they cannot send locked coins to other
accounts until those coins have been unlocked. When it comes to governance, it
is yet undefined if we want to allow a vesting account to be able to deposit
vesting coins into proposals.

In addition, a vesting account vests all of its coin denominations at the same
rate. This may be subject to change.

**Note**: A vesting account could have some vesting and non-vesting coins. To
support such a feature, the `GenesisAccount` type will need to be updated in
order to make such a distinction.

## Vesting Account Types

```go
// VestingAccount defines an interface that any vesting account type must
// implement.
type VestingAccount interface {
    Account
    AssertIsVestingAccount() // existence implies that account is vesting

    // Calculates the amount of coins that can be sent to other accounts given
    // the current time.
    SpendableCoins(Context) Coins
    // Performs delegation accounting.
    TrackDelegation(amount)
    // Performs undelegation accounting.
    TrackUndelegation(amount)
}

// BaseVestingAccount implements the VestingAccount interface. It contains all
// the necessary fields needed for any vesting account implementation.
type BaseVestingAccount struct {
    BaseAccount

    OriginalVesting  Coins // coins in account upon initialization
    DelegatedFree    Coins // coins that are vested and delegated
    EndTime          Time // when the coins become unlocked
}

// ContinuousVestingAccount implements the VestingAccount interface. It
// continuously vests by unlocking coins linearly with respect to time.
type ContinuousVestingAccount struct {
    BaseAccount
    BaseVestingAccount

    DelegatedVesting Coins // coins that vesting and delegated
    StartTime        Time // when the coins start to vest
}

// DelayedVestingAccount implements the VestingAccount interface. It vests all
// coins after a specific time, but non prior. In other words, it keeps them
// locked until a specified time.
type DelayedVestingAccount struct {
    BaseAccount
    BaseVestingAccount
}
```

## Vesting Account Specification

Given a vesting account, we define the following in the proceeding operations:

- `OV`: The original vesting coin amount. It is a constant value.
- `V`: The number of `OV` coins that are still _vesting_. It is derived by `OV`, `StartTime` and `EndTime`. This value is computed on demand and not on a per-block basis.
- `V'`: The number of `OV` coins that are _vested_ (unlocked). This value is computed on demand and not a per-block basis.
- `DV`: The number of delegated _vesting_ coins. It is a variable value. It is stored and modified directly in the vesting account.
- `DF`: The number of delegated _vested_ (unlocked) coins. It is a variable value. It is stored and modified directly in the vesting account.
- `BC`: The number of `OV` coins less any coins that are transferred, which can be negative, or delegated (`DV + DF`). It is considered to be balance of the embedded base account. It is stored and modified directly in the vesting account.

### Determining Vesting & Vested Amounts

It is important to note that these values are computed on demand and not on a
mandatory per-block basis.

#### Continuously Vesting Accounts

To determine the amount of coins that are vested for a given block `B`, the
following is performed:

1. Compute `X := B.Time - StartTime`
2. Compute `Y := EndTime - StartTime`
3. Compute `V' := OV * (X / Y)`
4. Compute `V := OV - V'`

Thus, the total amount of _vested_ coins is `V'` and the remaining amount, `V`,
is _vesting_.

```go
func (cva ContinuousVestingAccount) GetVestedCoins(b Block) Coins {
    // We must handle the case where the start time for a vesting account has
    // been set into the future or when the start of the chain is not exactly
    // known.
    if b.Time < va.StartTime {
        return ZeroCoins
    }

    x := b.Time - cva.StartTime
    y := cva.EndTime - cva.StartTime

    return cva.OriginalVesting * (x / y)
}

func (cva ContinuousVestingAccount) GetVestingCoins(b Block) Coins {
    return cva.OriginalVesting - cva.GetVestedCoins(b)
}
```

#### Delayed/Discrete Vesting Accounts

Delayed vesting accounts are easier to reason about as they only have the full
amount vesting up until a certain time, then they all become vested (unlocked).

```go
func (dva DelayedVestingAccount) GetVestedCoins(b Block) Coins {
    if b.Time >= dva.EndTime {
        return dva.OriginalVesting
    }

    return ZeroCoins
}

func (dva DelayedVestingAccount) GetVestingCoins(b Block) Coins {
    return cva.OriginalVesting - cva.GetVestedCoins(b)
}
```

### Transferring/Sending

#### Continuously Vesting Accounts

At any given time, a continuous vesting account may transfer: `min((BC + DV) - V, BC)`.

In other words, a vesting account may transfer the minimum of the base account
balance and the base account balance plus the number of currently delegated
vesting coins less the number of coins vested so far.

```go
func (cva ContinuousVestingAccount) SpendableCoins() Coins {
    bc := cva.GetCoins()
    return min((bc + cva.DelegatedVesting) - cva.GetVestingCoins(), bc)
}
```

##### Delayed/Discrete Vesting Accounts

A delayed vesting account may send any coins it has received. In addition, if it
has fully vested, it can send any of it's vested coins.

```go
func (dva DelayedVestingAccount) SpendableCoins() Coins {
    bc := dva.GetCoins()
    return bc - dva.GetVestingCoins()
}
```

##### Keepers/Handlers

The corresponding `x/bank` keeper should appropriately handle sending coins
based on if the account is a vesting account or not.

```go
func SendCoins(from Account, to Account amount Coins) {
    if isVesting(from) {
        sc := from.SpendableCoins()
    } else {
        sc := from.GetCoins()
    }

    if amount <= sc {
        from.SetCoins(sc - amount)
        to.SetCoins(amount)
        // save accounts...
    }
}
```

### Delegating

#### Continuously Vesting Accounts

For a continuous vesting account attempting to delegate `D` coins, the following
is performed:

1. Verify `BC >= D > 0`
2. Compute `X := min(max(V - DV, 0), D)` (portion of `D` that is vesting)
3. Compute `Y := D - X` (portion of `D` that is free)
4. Set `DV += X`
5. Set `DF += Y`
6. Set `BC -= D`

```go
func (cva ContinuousVestingAccount) TrackDelegation(amount Coins) {
    x := min(max(cva.GetVestingCoins() - cva.DelegatedVesting, 0), amount)
    y := amount - x

    cva.DelegatedVesting += x
    cva.DelegatedFree += y
}
```

##### Delayed/Discrete Vesting Accounts

For a delayed vesting account, it can only delegate with received coins and
coins that are fully vested so we only need to update `DF`.

```go
func (dva DelayedVestingAccount) TrackDelegation(amount Coins) {
    dva.DelegatedFree += amount
}
```

##### Keepers/Handlers

```go
func DelegateCoins(from Account, amount Coins) {
    // canDelegate checks different semantics for continuous and delayed vesting
    // accounts
    if isVesting(from) && canDelegate(from) {
        sc := from.GetCoins()

        if amount <= sc {
            from.TrackDelegation(amount)
            from.SetCoins(sc - amount)
            // save account...
        }
    } else {
        sc := from.GetCoins()

        if amount <= sc {
            from.SetCoins(sc - amount)
            // save account...
        }
    }
}
```

### Undelegating

#### Continuously Vesting Accounts

For a continuous vesting account attempting to undelegate `D` coins, the
following is performed:

1. Verify `(DV + DF) >= D > 0` (this is simply a sanity check)
2. Compute `Y := min(DF, D)` (portion of `D` that should become free, prioritizing free coins)
3. Compute `X := D - Y` (portion of `D` that should remain vesting)
4. Set `DV -= X`
5. Set `DF -= Y`
6. Set `BC += D`

```go
func (cva ContinuousVestingAccount) TrackUndelegation(amount Coins) {
    y := min(cva.DelegatedFree, amount)
    x := amount - y

    cva.DelegatedVesting -= x
    cva.DelegatedFree -= y
}
```

**Note**: If a delegation is slashed, the continuous vesting account will end up
with excess an `DV` amount, even after all its coins have vested. This is because
undelegating free coins are prioritized.

##### Delayed/Discrete Vesting Accounts

For a delayed vesting account, it only needs to add back the `DF` amount since
the account is fully vested.

```go
func (dva DelayedVestingAccount) TrackUndelegation(amount Coins) {
    dva.DelegatedFree -= amount
}
```

##### Keepers/Handlers

```go
func UndelegateCoins(to Account, amount Coins) {
    if isVesting(to) {
        if to.DelegatedFree + to.DelegatedVesting >= amount {
            to.TrackUndelegation(amount)
            AddCoins(to, amount)
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

## Initializing at Genesis

To initialize both vesting accounts and base accounts, the `GenesisAccount`
struct will include an `EndTime`. Accounts meant to be of type `BaseAccount` will
have `EndTime = 0`. The `initChainer` method will parse the GenesisAccount into
BaseAccounts and VestingAccounts as appropriate.

```go
type GenesisAccount struct {
    Address        sdk.AccAddress
    GenesisCoins   sdk.Coins
    EndTime        int64
}

func initChainer() {
    for genAcc in GenesisAccounts {
        baseAccount := BaseAccount{
            Address: genAcc.Address,
            Coins:   genAcc.GenesisCoins,
        }

        if genAcc.EndTime != 0 {
            vestingAccount := ContinuousVestingAccount{
                BaseAccount:      baseAccount,
                OriginalVesting:  genAcc.GenesisCoins,
                StartTime:        RequestInitChain.Time,
                EndTime:          genAcc.EndTime,
            }

            AddAccountToState(vestingAccount)
        } else {
            AddAccountToState(baseAccount)
        }
    }
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
6. Sends 2 coins. At this point the account cannot send anymore until further coins vest or it receives additional coins. It can still however, delegate.
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
6. Undelegate from validator B (5 coins). The account at this point can only send 2.5 coins unless it receives more coins or until more coins vest. It can still however, delegate.
    ```
    DV = 5 - 2.5 = 2.5
    DF = 2.5 - 2.5 = 0
    BC = 2.5 + 5 = 7.5
    ```

Notice how we have an excess amount of `DV`.

## Glossary

- OriginalVesting: The amount of coins (per denomination) that are initially part of a vesting account. These coins are set at genesis.
- StartTime: The BFT time at which a vesting account starts to vest.
- EndTime: The BFT time at which a vesting account is fully vested.
- DelegatedFree: The tracked amount of coins (per denomination) that are delegated from a vesting account that have been fully vested at time of delegation.
- DelegatedVesting: The tracked amount of coins (per denomination) that are delegated from a vesting account that were vesting at time of delegation.
- ContinuousVestingAccount: A vesting account implementation that vests coins linearly over time.
- DelayedVestingAccount: A vesting account implementation that only fully vests all coins at a given time.
