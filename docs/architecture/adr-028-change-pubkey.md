# ADR 028: Change PubKey

## Changelog

- 30-09-2020: Initial Draft

## Status

Proposed

## Context

This msg will update the public key associated with an account to a new public key, while keeping the same address.
		 
This can be used for purposes such as passing ownership of an account to a new key for security reasons or upgrading multisig signers.

## Decision

We will create a module called `changepubkey` that handle all the actions related to pubkey change stuff.

In addition, bonus gas amount for changing pubkey will be configured on auth module as parameter `PubKeyChangeCost`. (This can be part of changepubkey module to make it standalone.)

## Consequences

### Positive

This can be used for purposes such as passing ownership of an account to a new key for security reasons or upgrading multisig signers.

### Negative

### Neutral

## References

