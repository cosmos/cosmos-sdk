# ADR 000: Generalize Genesis Accounts

## Changelog

- 2019-08-30: initial draft

## Context

Summary: The `auth` module allows custom account types, but the `genaccounts` module does not.

Currently the SDK allows for custom account types; the `auth` keeper stores any type fulfilling its Account interface. However `auth` does not handle exporting or loading accounts to/from a genesis file, this is done by `genaccounts`, which only handles one of 4 concrete account types.

Projects wanting to use custom vesting accounts need to fork and modify `genaccounts`.


## Decision

We will
 - marshal/unmarshal accounts (interface types) directly using amino
 - remove `genaccounts`'s custom genesis account type
 - and since the above removes the majority of `genaccounts`’s code, move all logic into auth and remove the `genaccounts` module.

## Status

Proposed

## Consequences

### Positive

 - custom accounts can be used without needing to fork `genaccounts`
 - reduced LoC

### Negative


### Neutral

- genaccounts module no longer exists
- accounts in genesis files are stored under `auth.accounts` rather than `accounts`
- the `add-genesis-account` cli command is now in auth

## References
