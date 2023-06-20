# ADR-066: Keyring plugin

## Changelog

* Jun 20, 2023: Initial Draft (@bizk)

## Status

DRAFT

## Abstract

The keyring can interact with external plugins wrote in any language rough a GRPC
connection using the Hashicorp plugin system. 

## Context

Keyring can be abstracted and implemented in any language, while maintaining it's
capacity to use the rest of the cosmos-sdk features. This approach adds extendability
to the system itself, and provides more adoption.

## Consequences

We will have a keystone client / server inside the cosmos-sdk and the respective plugins used
in the users repositories. We will also have a mechanism to interact and trigger keyring
operations in any language regardless of the implementation.

## References

* [Keyring Poc Implementation](https://github.com/Zondax/keyringPoc)
