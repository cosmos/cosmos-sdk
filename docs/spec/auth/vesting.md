# Vesting

## Intro and Requirements

This paper specifies vesting account implementation for the Cosmos Hub.
The requirements for this vesting account is that it should be initialized
during genesis with a starting balance `X` coins and a vesting end time `T`.

The owner of this account should be able to delegate to validators
and vote with locked coins, however they cannot send locked coins to other
accounts until those coins have been unlocked.

When it comes to governance, it is yet undefined if we want to allow a vesting
account to be able to deposit vesting coins into proposals.

It is also important to note that a vesting account vests all of it's coin
denominations at the same rate. This may be subject to change.

__Note__: A vesting account could have some vesting and non-vesting coins at
genesis, however, the latter is unlikely.

## Vesting Account Definition

```go
type VestingAccount interface {
    Account
    AssertIsVestingAccount() // existence implies that account is vesting

    // Calculates the amount of coins that can be sent to other accounts given
    // the current time.
    SendableCoins(sdk.Context) sdk.Coins
}

// ContinuousVestingAccount implements the VestingAccount interface. It
// continuously vests by unlocking coins linearly with respect to time.
type ContinuousVestingAccount struct {
    BaseAccount

    OriginalVesting  sdk.Coins // coins in account upon initialization
    DelegatedVesting sdk.Coins // coins that vesting and delegated
    DelegatedFree    sdk.Coins // coins that are vested and delegated

    // StartTime and EndTime are used to calculate how much of OriginalVesting
    // is unlocked at any given point.
    StartTime time.Time
    EndTime   time.Time
}

// DelayedVestingAccount implements the VestingAccount interface. It vests all
// coins after a specific time, but non prior. In other words, it keeps them
// locked until a specified time.
type DelayedVestingAccount struct {
    BaseAccount

    OriginalVesting  sdk.Coins // coins in account upon initialization

    EndTime time.Time // when the coins become unlocked
}
```

## Vesting Account Implementation

Given a vesting account, we define the following in the proceeding operations:

```
OV = OriginalVesting (constant)
V = Number of OV coins that are still vesting (derived by OV and the start/end times)
DV = DelegatedVesting (variable)
DF = DelegatedFree (variable)
BC (BaseAccount.Coins) = OV - transferred (can be negative) - delegated (DV + DF)
```

__Note__: The above are explicitly stored and modified on the vesting account.

### Operations

#### Determining Vesting Amount

To determine the amount of coins that are vested for a given block `B`, the
following is performed:

1. Compute `X := B.Time - StartTime`
2. Compute `Y := EndTime - StartTime`
3. Compute `V' := OV * (X / Y)`

Thus, the number of vested coins is defined by `V'` and as a result the number of
vesting coins equates to `V := OV - V'`. It is important to note these values are
calculated on demand and not on a per-block basis.

#### Transferring/Sending

At any given time, a vesting account may transfer: `min((BC + DV) - V, BC)`.

#### Delegating

For a vesting account attempting to delegate `D` coins, the following is performed:

1. Verify `BC >= D > 0`
2. Compute `X := min(max(V - DV, 0), D)` (portion of `D` that is vesting)
3. Compute `Y := D - X` (portion of `D` that is free)
4. Set `DV += X`
5. Set `DF += Y`
6. Set `BC -= D`

#### Undelegating

For a vesting account attempting to undelegate `D` coins, the following is performed:

1. Verify `(DV + DF) >= D > 0` (this is simply a sanity check)
2. Compute `Y := min(DF, D)` (portion of `D` that should become free, prioritizing free coins)
3. Compute `X := D - Y` (portion of `D` that should remain vesting)
4. Set `DV -= X`
5. Set `DF -= Y`
6. Set `BC += D`

## Keepers & Handlers

The `VestingAccount` implementations reside in `x/auth`. However, any keeper in
a module (e.g. staking in `x/stake`) wishing to potentially utilize any vesting
coins, must call explicit methods on the `BankKeeper` (e.g. `DelegateCoins`)
opposed to `SendCoins` and `SubtractCoins`.

In addition, the vesting account should also be able to spend any coins it
receives from other users. Thus, the bank module's `MsgSend` handler should
error if a vesting account is trying to send an amount that exceeds their
unlocked coin amount.

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
