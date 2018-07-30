## Vesting

### Intro and Requirements

This paper specifies changes to the auth and bank modules to implement vested accounts for the Cosmos Hub. 
The requirements for this vested account is that it should be capable of being initialized during genesis with
a starting balance X coins and a vesting blocktime T. The owner of this account should be able to delegate to validators and vote,
however they cannot send their coins to other accounts until the account has fully vested. Thus, the bank module's `MsgSend` handler 
should error if a vested account is trying to send an amount before time T.

### Implementation

##### Changes to x/auth Module

The `Account` interface will specify both the Account type and any parameters it needs.

```go
// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
    Type() string // returns the type of the account

	GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() int64
	SetAccountNumber(int64) error

	GetSequence() int64
	SetSequence(int64) error

	GetCoins() sdk.Coins
    SetCoins(sdk.Coins) error
    
    // Getter and setter methods for account params
    // Parameters can be understood to be a map[string]interface{} with encoded keys and vals in store
    // It is upto handler to use these appropriately
    GetParams([]byte) []byte
    SetParams([]byte, []byte) error
}
```

The `Type` method will allow handlers to determine what type of account is sending the message, and the 
handler can then call `GetParams` to handle the specific account type using the parameters it expects to 
exist in the parameter map.

The `VestedAccount` will be an implementation of `Account` interface that wraps `BaseAccount` with 
`Type() => "vested` and params, `GetParams() => {"TimeLock": N (int64)}`. 
`SetParams` will be disabled as we do not want to update params after vested account initialization.


`auth.AccountMapper` will be modified handle vested accounts as well. Specific changes 
are omitted in this doc for succinctness.


##### Changes to bank MsgSend Handler

Since a vested account should be capable of doing everything but sending, the restriction should be 
handled at the `bank.Keeper` level. Specifically in methods that are explicitly used for sending like 
`sendCoins` and `inputOutputCoins`. These methods must check an account's `Type` method; if it is a vested 
account (i.e. `acc.Type() == "vested"`):

1. Check if `ctx.BlockHeader().Time < acc.GetParams()["BlockLock"]`
2. If `true`, the account is still vesting, return sdk.Error. Else, allow transaction to be processed as normal.

### Initializing at Genesis

To initialize both vested accounts and base accounts, the `GenesisAccount` struct will be:

```go
type GenesisAccount struct {
	Address  sdk.AccAddress `json:"address"`
    Coins    sdk.Coins      `json:"coins"`
    Type     string         `json:"type"`
    TimeLock int64          `json:"lock"`
}
```

During `InitChain`, the GenesisAccount's are decoded. If they have `Type == "vested`, a vested account with parameters => 
`{"TimeLock": N}` gets created and put in initial state. Otherwise if `Type == "base"` a base account is created 
and the `TimeLock` attribute of corresponding `GenesisAccount` is ignored. `InitChain` will panic on any other account types.

### Pros and Cons

##### Pros

- Easily Extensible. If more account types need to get added in the future or if developers building on top of SDK 
want to handle multiple custom account types, they simply have to implement the `Account` interface with unique `Type` 
and their custom parameters.
- Handlers (and their associated keepers) get to determine what types of accounts they will handle and can use the parameters 
in Account interface to handle different accounts appropriately.

##### Cons

- Changes to `Account` interface
- Slightly more complex code in `bank.Keeper` functions
