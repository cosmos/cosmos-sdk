# ADR 3: Subaccounts

## Changelog

## Context

Currently module accounts must be declared upon supply keeper initialization. Furthermore, they don't allow for separation of fungible coins within an account.

The account structure should be modified so a `ModuleAccount` can dynamically add accounts.

## Decision

Add the following interfaces and permissions into `x/auth`.

MultiAccounts and SubAccounts could be subkey accounts by having subkey account implement the interface functions.

### Implementation Changes

Introduce two new interfaces:

* `MultiAccount`
* `SubAccount`

MultiAccount implements the `Account` interface.
It maintains a list of its subaccounts as well as a list of permissions defined upon initialization of supply keeper (permAddrs).
`SetCoins` returns an error. This prevents MultiAccount address from having a balance.
MultiAccount has no pubkey for ModuleAccounts. MultiAccount for non ModuleAccounts could perhaps be a subkey account. MultiAccount pubkey would be master pubkey.
Upon initialization of a MultiAccounts, a limit can be set on the max number of sub accounts can be set. There should also be the option to set the max number of sub accounts as unbonded.
MultiAccount Constructor returns a MultiAccount with no sub accounts.
MultiAccount `GetCoins` returns sum of sub account balances.
SubAccounts can only be appended. To invalidate an account we would add `SetAccountDisabled` which sets the `disabled` field to true. 
A disabled account could allow for withdraws, but cannot recieve any coins.
Passively tracks the sum of all account balances.

```
type MultiAccount interface {
    // MultiAccount interface functions
    CreateSubAccount(pubkey, address) int // returns id of subaccount
    GetSubAccount(id int) SubAccount
   
    // account interface functions
    GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error 

	GetPubKey() crypto.PubKey 
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error

	SpendableCoins(blockTime time.Time) sdk.Coins

	String() string

        
}
```

```
type SubAccount interface {
    // SubAccount interface functions    
    SetAccountDisabled()
    SetAccountEnabled()
    
    AddPermissions(perms ...string)
    RemovePermissions(perms ...string)

    GetPermissions() []string

    // account interface functions
    GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error 

	GetPubKey() crypto.PubKey 
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error

	SpendableCoins(blockTime time.Time) sdk.Coins

	String() string

     
}
```

// possible implementation of MultiAccount
```
type ModuleMultiAccount struct {
    subaccs []SubAccount
    permissions []string
    maxNumSubAccs uint
    coins sdk.Coins // passively track all sub account balances
    disabled bool
}
```


SubAccount implements the `Account` interface.
SubAccount's address is the multi account address with the id appended
SubAccount permissions must be a sub set of its multi account's permissions.

// possible implementation of SubAccount
```
type SubAccount struct {
    address sdk.AccAddress // MultiAccount (parent) address with index appended
    pubkey
    id uint // index of subaccount
    permissions []string
}
```

**Other changes**

Add an invariant check for MultiAccount `GetCoins`, which iterates over all subaccs to see if the sum of the subacc balances equals the passive tracking which is returned in `GetCoins`

Update BankKeepers SetCoins function to return an error instead of calling panic on the account's SetCoins error.

## Status

Proposed

## Consequences

### Positive

* Accounts can now separate fungible coins
* Accounts can distribute permissions to sub accounts.

### Negative

* Brings permissions into `x/auth`

### Neutral

* Adds a new Account types

## References

Issues: [4657] (https://github.com/cosmos/cosmos-sdk/issues/4657)

