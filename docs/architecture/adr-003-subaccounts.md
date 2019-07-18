# ADR 3: Module SubAccounts

## Changelog

## Context

Currently `ModuleAccount`s must be declared upon supply keeper initialization. In addition to this they don't allow for separation of fungible coins within an account.

New structs should be added so a `ModuleAccount` can dynamically add accounts.

## Decision

Add the following structs into `x/supply`.

### Implementation Changes

Introduce two new structs:

* `ModuleMultiAccount`
* `SubAccount`

`ModuleMultiAccount` maintains an array of `SubAccount` as well as a list of permissions defined upon initialization of supply keeper (permAddrs).
It has no pubkey and no limit on the number of `SubAccount`s that can be created.
Its constructor returns a `ModuleMultiAccount` with no `SubAccount`s.
`MultiModuleAccount` will implement the Account interface.

```go
// SetCoins will return an error to prevent ModuleMultiAccount address from having a balance.
// SubAccounts can only be appended to the SubAccount array.
// Passively tracks the sum of all SubAccount balances.
type ModuleMultiAccount struct {
    Subaccs []SubAccount
    Permissions []string
    Coins sdk.Coins // passively track all sub account balances

    CreateSubAccount(pubkey, address) int // returns id of subaccount
    GetSubAccount(id int) SubAccount
}
```

To invalidate a `SubAccount` the `ModuleMultiAccount` calls `SetAccountDisabled` for a `SubAccount`

```go
// Implements the Account interface. Address is the ModuleMultiAccount address with the id appended.
// Permissions must be a subset of its ModuleMultiAccount permissions.
// A disabled account can do withdraws, but cannot recieve any coins.
type SubAccount struct {
    Address sdk.AccAddress // MultiAccount (parent) address with index appended
    ID uint // index of subaccount
    Permissions []string
    Disabled bool

    SetAccountDisabled()

    AddPermissions(perms ...string)
    RemovePermissions(perms ...string)

    GetPermissions() []string
}
```

**Other changes**

Add an invariant check for MultiAccount `GetCoins`, which iterates over all `SubAccount`s to see if the sum of the `SubAccount` balances equals the passive tracking which is returned in `GetCoins`

Update BankKeepers SetCoins function to return an error instead of calling panic on the account's SetCoins error.

## Status

Proposed

## Consequences

### Positive

* ModuleAccount can separate fungible coins.
* ModuleAccount can dynamically add accounts.
* ModuleAccount can distribute permissions to SubAccounts.

### Negative

### Neutral

* Adds a new Account types

## References

Issues: [4657] (https://github.com/cosmos/cosmos-sdk/issues/4657)
