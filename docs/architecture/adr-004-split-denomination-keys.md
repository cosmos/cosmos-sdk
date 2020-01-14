# ADR 004: Split Denomination Keys

## Changelog

- 2020-01-08: Initial version
- 2020-01-09: Alterations to handle vesting accounts

## Context

With permissionless IBC, anyone will be able to send arbitrary denominations to any other account. Currently, all non-zero balances are stored along with the account in an `sdk.Coins` struct, which creates a potential denial-of-service concern, as too many denominations will become expensive to load & store each time
the account is modified. See issues [5467](https://github.com/cosmos/cosmos-sdk/issues/5467) and [4982](https://github.com/cosmos/cosmos-sdk/issues/4982) for additional context.

Simply rejecting incoming deposits after a denomination count limit doesn't work, since it opens up a griefing vector: someone could send a user lots of nonsensical coins over IBC, and then prevent the user from receiving real denominations (such as staking rewards).

## Decision

Balances shall be stored per-account & per-denomination under a denomination- and account-unique key, thus enabling O(1) read & write access to the balance of a particular account in a particular denomination.

### Account interface (x/auth)

`GetCoins()` and `SetCoins()` will be removed from the account interface, since coin balances will now be stored in & managed by the bank module.

`GetVestedCoins()`, `GetVestingCoins()`, `GetOriginalVesting()`, `GetDelegatedFree()`, and `GetDelegatedVesting()` will be altered to take a bank keeper and a denomination as two additional arguments, which will be used to lookup the balances from the base account as necessary.

Vesting accounts will continue to store original vesting, delegated free, and delegated vesting coins (which is safe since these cannot contain arbitrary denominations).

### Bank keeper (x/bank)

`GetBalance(addr AccAddress, denom string) Int` and `SetBalance(addr AccAddress, denom string, balance Int)` methods will be added to the bank keeper to retrieve & set balances, respectively.

`DelegateCoins()` and `UndelegateCoins()` will be altered to take a single `sdk.Coin` (one denomination & amount) instead of `sdk.Coins`. They will read balances directly instead of calling `GetCoins()` (which no longer exists).

`SubtractCoins()` and `AddCoins()` will be altered to read & write the balances directly instead of calling `GetCoins()` / `SetCoins()` (which no longer exist).

`trackDelegation()` and `trackUndelegation()` will be altered to read & write the balances directly instead of calling `GetCoins()` / `SetCoins()` (which no longer exist).

External APIs will need to scan all balances under an account to retain backwards-compatibility - additional methods should be added to fetch a balance for a single denomination only.

## Status

Proposed.

## Consequences

### Positive

- O(1) reads & writes of balances (with respect to the number of denominations for which an account has non-zero balances)

### Negative

- Slighly less efficient reads/writes when reading & writing all balances of a single account in a transaction.

### Neutral

None in particular.

## References

Ref https://github.com/cosmos/cosmos-sdk/issues/5492
Ref https://github.com/cosmos/cosmos-sdk/issues/5467
