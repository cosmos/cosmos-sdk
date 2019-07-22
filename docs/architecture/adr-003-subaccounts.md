# ADR 3: Module SubAccounts

## Changelog

## Context

Currently `ModuleAccount`s must be declared upon Supply Keeper initialization. In addition to this they don't allow for separation of fungible coins within an account.

We want to support the ability to define and manage sub-module-accounts.

## Decision

We will use the type `ModuleMultiAccount` to manage sub-accounts of type `ModuleAccount`.
A `ModuleMultiAccount` may have zero or more `ModuleAccount`s.
Each sub-account will utilize the existing permissioned properties of `ModuleAccount`.
`ModuleMultiAccount` and `ModuleAccount` have no pubkeys.
There is no limit on the number of `ModuleAccount`s that a `ModuleMultiAccount` can have.
A `ModuleMultiAccount` has no permissions since it cannot hold any coins.
Its constructor returns a `ModuleMultiAccount` with no `ModuleAccount`s.
A sub-account cannot be removed from a `MultiModuleAccount`.
The account number assigned to sub-accounts will begin at 0 and be monotonically auto incrementing.

### Implementation Changes

Introduce a new type into `x/supply`:

* `ModuleMultiAccount`

```go
// Implements the Account interface.
// SetCoins will return an error to prevent ModuleMultiAccount address from having a balance.
// ModuleAccounts are appended to the SubAccounts array.
// Passively tracks the sum of all ModuleAccount balances.
type ModuleMultiAccount struct {
    SubAccounts []ModuleAccount
    Coins sdk.Coins // passively track all sub account balances

    CreateSubAccount(name string, permissions ...string) int // returns account number of sub-account
    GetSubAccount(subAccNumber int64) ModuleAccount
}
```

The `ModuleAccount` implementation will remain unchanged, but we will add the following constructor function:
```go
// NewEmptyModuleSubAccount creates a sub-account ModuleAccount which has an address created from
// the hash of the module's name with the sub-account number appended.
func NewEmptyModuleSubAccount(name string, subAccNumber uint64, permissions ...string) ModuleAccount {
    bz := make([]byte, 8)
    binary.LittleEndian.PutUint64(bz, subAccNumber)
    moduleAddress := append(NewModuleAddress(name), bz...)
    baseAcc := authtypes.NewBaseAccountWithAddress(moduleAddress)

    if err := validatePermissions(permissions...); err != nil {
        panic(err)
    }

    return &ModuleAccount{
        BaseAccount: &baseAcc,
        Name:        name,
        Permissions: permissions,
    } 
}
```

**Permissions**:

A `ModuleMultiAccount` has no permissions.

Since `ModuleAccount`s that are sub-accounts have the same name as its parent `ModuleMultiAccount`, a sub-account should only be granted a subset of the permissions registered with the Supply Keeper under its name.

**Other changes**

We will add an invariant check for the `ModuleMultiAccount` `GetCoins()` function, which will iterate over all SubAccounts to see if the sum of the `ModuleAccount` balances equals the passive tracking which is returned in `GetCoins()`

Bank Keepers `SetCoins()` function will be updated to return an error instead of calling panic on the account's SetCoins error.

## Status

Accepted

## Consequences

### Positive

* ModuleMultiAccount can separate fungible coins.
* ModuleMultiAccount can dynamically add accounts.
* ModuleMultiAccount can distribute permissions to sub-accounts.

### Negative

* sub-accounts cannot be removed from `ModuleMultiAccount`

### Neutral

* Use `ModuleAccount` type as a sub-account for `ModuleMultiAccount`
* Adds a new Account type

## References

Specs: [ModuleAccount](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/supply/01_concepts.md#module-accounts)

Issues: [4657](https://github.com/cosmos/cosmos-sdk/issues/4657)
