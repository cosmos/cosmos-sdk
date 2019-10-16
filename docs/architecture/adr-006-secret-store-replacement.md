# ADR 006: Secret Store Replacement

## Changelog

- July 29, 2019: Initial draft

## Context

Currently, an SDK application's CLI directory stores key material and metadata in a plain text database in the user’s home directory.  Key material is encrypted by a passphrase, protected by bcrypt hashing algorithm. Metadata (e.g. addresses, public keys, key storage details) is available in plain text. 

This is not desirable for a number of reasons. Perhaps the biggest reason is insufficient security protection of key material and metadata. Leaking the plain text allows an attacker to surveil what keys a given computer controls via a number of techniques, like compromised dependencies without any privilege execution. This could be followed by a more targeted attack on a particular user/computer.

All modern desktop computers OS (Ubuntu, Debian, MacOS, Windows) provide a built-in secret store that is designed to allow applications to store information that is isolated from all other applications and requires passphrase entry to access the data. 

We are seeking solution that provides a common abstraction layer to the many different backends and reasonable fallback for minimal platforms that don’t provide a native secret store.


## Decision

We recommend replacing the current Keybase backend based on LevelDB with [Keyring](https://github.com/99designs/keyring) by 99 designs. This application is designed to provide a common abstraction and uniform interface between many secret stores and is used by AWS Vault application by 99-designs application.

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

