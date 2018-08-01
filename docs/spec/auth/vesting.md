## Vesting

### Intro and Requirements

This paper specifies changes to the auth and bank modules to implement vesting accounts for the Cosmos Hub. 
The requirements for this vesting account is that it should be capable of being initialized during genesis with
a starting balance X coins and a vesting blocktime T. The owner of this account should be able to delegate to validators 
and vote with locked coins, however they cannot send locked coins to other accounts until those coins have been unlocked. 
The vesting account should also be able to spend any coins it receives from other users or from fees/inflation rewards. 
Thus, the bank module's `MsgSend` handler should error if a vesting account is trying to send an amount that exceeds their 
unlocked coin amount.

### Implementation

##### Vesting Account implementation

```go
type VestingAccount interface {
    Account
    AssertIsVestingAccount() // existence implies that account is vesting.
}

// Implements Vesting Account
// Continuously vests by unlocking coins linearly with respect to time
type ContinuousVestingAccount struct {
    BaseAccount
    OriginalCoins sdk.Coins
    ReceivedCoins sdk.Coins
    StartTime     int64
    EndTime       int64
}

func (vacc ContinuousVestingAccount) ConvertAccount() BaseAccount {
    if T > vacc.EndTime {
        // Convert to BaseAccount
    }
}
```

The `VestingAccount` interface is used purely to assert that an account is a vesting account like so:

```go
vacc, ok := acc.(VestingAccount); ok
```

The `ContinuousVestingAccount` struct implements the Vesting account interface. It uses `OriginalCoins`, `ReceivedCoins`, 
`StartTime`, and `EndTime` to calculate how many coins are sendable at any given point. Once the account has fully vested, 
the next `bank.MsgSend` will convert the account into a `BaseAccount` and store it in state as such from that point on. 
Since the vesting restrictions need to be implemented on a per-module basis, the `ContinuouosVestingAccount` implements 
the `Account` interface exactly like `BaseAccount`.

##### Changes to Keepers/Handler

Since a vesting account should be capable of doing everything but sending with its locked coins, the restriction should be 
handled at the `bank.Keeper` level. Specifically in methods that are explicitly used for sending like 
`sendCoins` and `inputOutputCoins`. These methods must check that an account is a vesting account using the check described above.
NOTE: `Now = ctx.BlockHeader().Time`

1. If `Now < vacc.EndTime`
   1. Calculate `SendableCoins := ReceivedCoins + OriginalCoins * (Now - StartTime)/(EndTime - StartTime))`
      - NOTE: `SendableCoins` may be greater than total coins in account. This is because coins can be subtracted by staking module.
        `SendableCoins` denotes maximum coins allowed to be spent right now.
   2. If `msg.Amount > SendableCoins`, return sdk.Error. Else, allow transaction to process normally.
2. Else:
   1. Convert account to `BaseAccount` and process normally.

Coins that are sent to a vesting account after initialization either through users sending them coins or through fees/inflation rewards 
should be spendable immediately after receiving them. Thus, handlers (like staking or bank) that send coins that a vesting account did not 
originally own should increment `ReceivedCoins` by the amount sent.

WARNING: Handlers SHOULD NOT update `ReceivedCoins` if they were originally sent from the vesting account. For example, if a vesting account 
unbonds from a validator, their tokens should be added back to account but `ReceivedCoins` SHOULD NOT be incremented.
However when the staking handler is handing out fees or inflation rewards, then `ReceivedCoins` SHOULD be incremented.

### Initializing at Genesis

To initialize both vesting accounts and base accounts, the `GenesisAccount` struct will be:

```go
type GenesisAccount struct {
	Address sdk.AccAddress `json:"address"`
    Coins   sdk.Coins      `json:"coins"`
    EndTime int64          `json:"lock"`
}
```

During `InitChain`, the GenesisAccounts are decoded. If `EndTime == 0`, a BaseAccount gets created and put in Genesis state. 
Otherwise a vesting account is created with `StartTime = RequestInitChain.Time`, `EndTime = gacc.EndTime`, and `OriginalCoins = Coins`. 
