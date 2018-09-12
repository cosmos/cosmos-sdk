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

    // Calculates amount of coins that can be sent to other accounts given the current time
    SendableCoins(sdk.Context) sdk.Coins
}

// Implements Vesting Account
// Continuously vests by unlocking coins linearly with respect to time
type ContinuousVestingAccount struct {
    BaseAccount
    OriginalVestingCoins sdk.Coins // Coins in account on Initialization
    ReceivedCoins        sdk.Coins // Coins received from other accounts
    SentCoins            sdk.Coins // Coins sent to other accounts

    // StartTime and EndTime used to calculate how much of OriginalCoins is unlocked at any given point
    StartTime time.Time
    EndTime   time.Time
}

// Uses time in context to calculate total unlocked coins
SendableCoins(vacc ContinuousVestingAccount, ctx sdk.Context) sdk.Coins:
    
    // Coins unlocked by vesting schedule
    unlockedCoins := ReceivedCoins - SentCoins + OriginalVestingCoins * (Now - StartTime) / (EndTime - StartTime)

    // Must still check for currentCoins constraint since some unlocked coins may have been delegated.
    currentCoins := vacc.BaseAccount.GetCoins()

    // min will return sdk.Coins with each denom having the minimum amount from unlockedCoins and currentCoins
    return min(unlockedCoins, currentCoins)

```

The `VestingAccount` interface is used to assert that an account is a vesting account like so:

```go
vacc, ok := acc.(VestingAccount); ok
```

as well as to calculate the SendableCoins at any given moment.

The `ContinuousVestingAccount` struct implements the Vesting account interface. It uses `OriginalVestingCoins`, `ReceivedCoins`, 
`SentCoins`, `StartTime`, and `EndTime` to calculate how many coins are sendable at any given point. 
Since the vesting restrictions need to be implemented on a per-module basis, the `ContinuousVestingAccount` implements 
the `Account` interface exactly like `BaseAccount`. Thus, `ContinuousVestingAccount.GetCoins()` will return the total of 
both locked coins and unlocked coins currently in the account. Delegated coins are deducted from `Account.GetCoins()`, but do not count against unlocked coins because they are still at stake and will be reinstated (partially if slashed) after waiting the full unbonding period.

##### Changes to Keepers/Handler

Since a vesting account should be capable of doing everything but sending with its locked coins, the restriction should be 
handled at the `bank.Keeper` level. Specifically in methods that are explicitly used for sending like 
`sendCoins` and `inputOutputCoins`. These methods must check that an account is a vesting account using the check described above.

```go
if acc is VestingAccount and Now < vestingAccount.EndTime:
    // Check if amount is less than currently allowed sendable coins
    if msg.Amount > vestingAccount.SendableCoins(ctx) then fail
    else:
        vestingAccount.SentCoins += msg.Amount

else:
    // Account has fully vested, treat like regular account
    if msg.Amount > account.GetCoins() then fail

// All checks passed, send the coins
SendCoins(inputs, outputs)

```

Coins that are sent to a vesting account after initialization by users sending them coins should be spendable 
immediately after receiving them. Thus, handlers (like staking or bank) that send coins that a vesting account did not 
originally own should increment `ReceivedCoins` by the amount sent.
Unlocked coins that are sent to other accounts will increment the vesting account's `SentCoins` attribute.

CONTRACT: Handlers SHOULD NOT update `ReceivedCoins` if they were originally sent from the vesting account. For example, if a vesting account unbonds from a validator, their tokens should be added back to account but staking handlers SHOULD NOT update `ReceivedCoins`.
However when a user sends coins to vesting account, then `ReceivedCoins` SHOULD be incremented.

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
                BaseAccount:          baseAccount,
                OriginalVestingCoins: gacc.GenesisCoins,
                StartTime:            RequestInitChain.Time,
                EndTime:              gacc.EndTime,
            }
            AddAccountToState(vestingAccount)
        else:
            AddAccountToState(baseAccount)

```

### Formulas

`OriginalVestingCoins`: Amount of coins in account at Genesis

`CurrentCoins`: Coins currently in the baseaccount (both locked and unlocked: `vestingAccount.GetCoins`)

`ReceivedCoins`: Coins received from other accounts (always unlocked)

`LockedCoins`: Coins that are currently locked

`Delegated`: Coins that have been delegated (no longer in account; may be locked or unlocked)

`Sent`: Coins sent to other accounts (MUST be unlocked)

Maximum amount of coins vesting schedule allows to be sent:

`ReceivedCoins - SentCoins + OriginalVestingCoins * (Now - StartTime) / (EndTime - StartTime)`

`ReceivedCoins - SentCoins + OriginalVestingCoins - LockedCoins`

Coins currently in Account:

`CurrentCoins = OriginalVestingCoins + ReceivedCoins - Delegated - Sent`

`CurrentCoins = vestingAccount.GetCoins()`

**Maximum amount of coins spendable right now:**

`min( ReceivedCoins - SentCoins + OriginalVestingCoins - LockedCoins, CurrentCoins )`
