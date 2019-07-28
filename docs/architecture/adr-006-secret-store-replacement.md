# # ADR 006: Replace Keybase with platform specific secret store provided by the Keyring library

## Changelog

- July 29, 2019: Initial Draft

## Context

Currently, gaiacli stores key material and metadata in a plaintext database in the user’s home directory.  Key material is encrypted by a passphrase. Metadata is available in plaintext.

This is not desirable for a number of reasons. Perhaps the biggest is that leaking the plain allows an attacker to surveil what keys a given computer controls via any number of techniques like compromised dependencies without any privilege execution. This would be followed by a more target attack on a particular user/computer.

All modern desktop computers OS (Ubuntu,Debian, MacOS, Windows) provide a built in secret store that is designed to allow applications to store information that is isolated from all other applications and requires passphrase entry to access the data. 

We are seeking solution that provides a common abstraction layer to the many different backends and reasonable fallback for minimal platforms that don’t provide a native secret store.


## Decision

We recommend replacing the current Keybase backend based on leveldb with [Keyring](https://github.com/99designs/keyring) by 99 designs. This application is designed to provide a common abstraction between many secret stores and is used by aws-vault application by 99-designs application.

This appears to fulfill the requirement of protecting both key material and metadata from rouge software on a user’s machine.



## Status
> Proposed

This change is implemented in [Switch secret store to the keyring secret store by poldsam · Pull Request #4754 · cosmos/cosmos-sdk · GitHub](https://github.com/cosmos/cosmos-sdk/pull/4754)

The audit status on Keyring is currently unknown and may require input from security.

## Consequences

### Positive

Increased safety for users.

### Negative

Users must manually migrate.

Testing against all supported backends is difficult.

Running tests locally on a Mac require numerous repetitive password entries.

### Neutral

{neutral consequences}

## References

- {reference link}

