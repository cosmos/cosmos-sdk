## Vesting

### Intro and Requirements

This paper specifies changes to the auth and bank modules to implement vested accounts for the Cosmos Hub. 
The requirements for this vested account is that it should be capable of being initialized during genesis with
a starting balance X and a vesting blocknumber N. The owner of this account should be able to delegate to validators,
but they cannot send their initial coins to other accounts. However; funds sent to this account, or fees and 
inflation rewards from delegation should be spendable. Thus, the bank module's MsgSend handler should error if 
a vested account is trying to send an amount `x > currentBalance - initialBalance` before block N.

### Implementation

##### Changes to x/auth Module

The first change is to the Account interface to specify both the Account type and any parameters it needs.

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
    // It is upto handler to use these appropriately
    GetParams() map[string]interface{}
    SetParams(map[string]interface{}) error
}
```

The `Type` method will allow handlers to determine what type of account is sending the message, and the 
handler can then call `GetParams` to handle the specific account type using the parameters it expects to 
exist in the parameter map.

The `VestedAccount` will be an implementation of `Account` interface that wraps `BaseAccount` with 
`Type() => "vested` and params, `GetParams() => {"Funds": initialBalance (sdk.Coins), "BlockLock": blockN (int64)}`. 
`SetParams` will be disabled as we do not want to update params after vested account initialization. 
The `VestedAccount` will also maintain an attribute called `FreeCoins`


`auth.AccountMapper` to handle vested accounts as well. Specific changes 
are omitted in this doc for succinctness.


##### Changes to bank MsgSend Handler

Since a vested account should be capable of doing everything but sending, the restriction should be 
handled at the `bank.Keeper` level. Specifically in methods that are explicitly used for sending like 
`sendCoins` and `inputOutputCoins`. These methods must check an account's `Type` method; if it is a vested 
account (i.e. `acc.Type() == "vested"`):

1. Check if `ctx.BlockHeight() < acc.GetParams()["BlockLock"]`
  * If `true`, the account is still vesting 
2. If account is still vesting, check that `(acc.GetCoins() - acc.GetParams()["Funds"] - amount).IsValid()`.
  * This will check that amount trying to be spent will not come from initial balance.
3. If above checks pass, allow transaction to go through. Else, return sdk.Error.

### Initializing at Genesis

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
