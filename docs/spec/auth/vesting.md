## Vesting

### Intro and Requirements

This paper specifies vesting account implementation for the Cosmos Hub. 
The requirements for this vesting account is that it should be initialized during genesis with
a starting balance X coins and a vesting endtime T. The owner of this account should be able to delegate to validators 
and vote with locked coins, however they cannot send locked coins to other accounts until those coins have been unlocked. 
The vesting account should also be able to spend any coins it receives from other users. 
Thus, the bank module's `MsgSend` handler should error if a vesting account is trying to send an amount that exceeds their 
unlocked coin amount.

### Implementation

##### Vesting Account implementation

NOTE:  `Now = ctx.BlockHeader().Time`

```go
type VestingAccount interface {
    Account
    AssertIsVestingAccount() // existence implies that account is vesting.
    ConvertAccount(sdk.Context) BaseAccount

    // Calculates total amount of unlocked coins released by vesting schedule
    // May be larger than total coins in account right now
    TotalUnlockedCoins(sdk.Context) sdk.Coins
}

// Implements Vesting Account
// Continuously vests by unlocking coins linearly with respect to time
type ContinuousVestingAccount struct {
    BaseAccount
    OriginalCoins sdk.Coins // Coins in account on Initialization
    ReceivedCoins sdk.Coins // Coins received from other accounts

    // StartTime and EndTime used to calculate how much of OriginalCoins is unlocked at any given point
    StartTime     int64
    EndTime       int64
}

// ConvertAccount converts VestingAccount into BaseAccount
// Will convert only after account has fully vested
ConvertAccount(vacc ContinuousVestingAccount, ctx sdk.Context) (BaseAccount):
    if Now > vacc.EndTime:
        return vacc.BaseAccount

// Uses time in context to calculate total unlocked coins
TotalUnlockedCoins(vacc ContinuousVestingAccount, ctx sdk.Context) sdk.Coins:
    return ReceivedCoins + OriginalCoins * (Now - StartTime) / (EndTime - StartTime)

```

The `VestingAccount` interface is used to assert that an account is a vesting account like so:

```go
vacc, ok := acc.(VestingAccount); ok
```

as well as to convert to BaseAccount again once the account has fully vested.

The `ContinuousVestingAccount` struct implements the Vesting account interface. It uses `OriginalCoins`, `ReceivedCoins`, 
`StartTime`, and `EndTime` to calculate how many coins are sendable at any given point. Once the account has fully vested, 
the next `bank.MsgSend` will convert the account into a `BaseAccount` and store it in state as such from that point on. 
Since the vesting restrictions need to be implemented on a per-module basis, the `ContinuouosVestingAccount` implements 
the `Account` interface exactly like `BaseAccount`.

##### Changes to Keepers/Handler

Since a vesting account should be capable of doing everything but sending with its locked coins, the restriction should be 
handled at the `bank.Keeper` level. Specifically in methods that are explicitly used for sending like 
`sendCoins` and `inputOutputCoins`. These methods must check that an account is a vesting account using the check described above.

```go
if Now < vestingAccount.EndTime:
    // NOTE: SendableCoins may be greater than total coins in account 
    // because coins can be subtracted by staking module
    // SendableCoins denotes maximum coins allowed to be spent.
    if msg.Amount > vestingAccount.TotalUnlockedCoins() then fail

// Account fully vested, convert to BaseAccount
else:
    account = ConvertAccount(account) 

// Must still check if account has enough coins,
// since SendableCoins does not check this.
if msg.Amount > account.GetCoins() then fail

// All checks passed, send the coins
SendCoins(inputs, outputs)

```

Coins that are sent to a vesting account after initialization by users sending them coins should be spendable 
immediately after receiving them. Thus, handlers (like staking or bank) that send coins that a vesting account did not 
originally own should increment `ReceivedCoins` by the amount sent.

CONTRACT: Handlers SHOULD NOT update `ReceivedCoins` if they were originally sent from the vesting account. For example, if a vesting account unbonds from a validator, their tokens should be added back to account but `ReceivedCoins` SHOULD NOT be incremented.
However when the staking handler is handing out fees or inflation rewards, then `ReceivedCoins` SHOULD be incremented.

### Initializing at Genesis

To initialize both vesting accounts and base accounts, the `GenesisAccount` struct will include an EndTime. Accounts meant to be 
BaseAccounts will have `EndTime = 0`. The `initChainer` method will parse the GenesisAccount into BaseAccounts and VestingAccounts 
as appropriate.

```go
type GenesisAccount struct {
    Address        sdk.AccAddress `json:"address"`
    GenesisCoins   sdk.Coins      `json:"coins"`
    EndTime        int64          `json:"lock"`
}

initChainer:
    for gacc in GenesisAccounts:
        baseAccount := BaseAccount{
            Address: gacc.Address,
            Coins:   gacc.GenesisCoins,
        }
        if gacc.EndTime != 0:
            vestingAccount := ContinuouslyVestingAccount{
                BaseAccount:   baseAccount,
                OriginalCoins: gacc.GenesisCoins,
                StartTime:     RequestInitChain.Time,
                EndTime:       gacc.EndTime,
            }
            AddAccountToState(vestingAccount)
        else:
            AddAccountToState(baseAccount)

```

### Formulas

`OriginalCoins`: Amount of coins in account at Genesis

`CurrentCoins`: Coins currently in the baseaccount (both locked and unlocked)

`ReceivedCoins`: Coins received from other accounts (always unlocked)

`LockedCoins`: Coins that are currently locked

`Delegated`: Coins that have been delegated (no longer in account; may be locked or unlocked)

`Sent`: Coins sent to other accounts (MUST be unlocked)

Maximum amount of coins vesting schedule allows to be sent:

`ReceivedCoins + OriginalCoins * (Now - StartTime) / (EndTime - StartTime)`

`ReceivedCoins + OriginalCoins - LockedCoins`

Coins currently in Account:

`CurrentCoins = OriginalCoins + ReceivedCoins - Delegated - Sent`

`CurrentCoins = vestingAccount.BaseAccount.GetCoins()`

**Maximum amount of coins spendable right now:**

`min( ReceivedCoins + OriginalCoins - LockedCoins, CurrentCoins )`
