# ADR-066: Keyring plugin

## Changelog

* Jun 20, 2023: Initial Draft (@JulianToledano & @bizk)

## Status

DRAFT

## Abstract

This ADR describes a keyring implementation based on the hashicorp plugins over gRPC.

## Context

Currently, in the cosmos-sdk, the keyring implementation depends on
[99designs/keyring](https://github.com/99designs/keyring) module that isn't under actively maintenance.

For that reason, it is proposed to develop a new keyring implementation that leverages
HashiCorp plugins over gRPC. These can be abstracted and implemented in any language,
while maintaining its capacity to use the rest of the cosmos-sdk features.
This approach adds extendability to the system itself, and provides more adoption.

## Alternatives

There are several other options available, such as plugins over RPC or utilizing the  Go standard
library plugin package. However, in the end, all these plugin alternatives share similarities.

Another alternative is to reimplement the keystore [db](https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/crypto/keyring/keyring.go#L207)
interface. This db is where the `99desing` dependency lays. 

## Decision

We will define the gRPC service with all the necessary messages. Additionally, we will
implement one plugin for each backend currently supported in the current keyring.

## Consequences

### Backwards Compatibility

Backwards compatibility is guaranteed as the current keyring implementation and this new one can coexist.

### Positive

As plugins communicate with the main process using gRPC, plugins can be written in any language.

Teams can easily develop plugins to meet their specific requirements,
which opens the door to new functionalities in key management.

### Negative

The `Record` relies on its `cachedValue` field to retrieve the address, public and private keys. This
field  cannot be sent over gRPC. This will generate occasions when the `Record` must be deserialized
both in the plugin  and the SDK (main process), leading to some overhead.

### Neutral

Since plugins are separate subprocesses initiated from the main process, it is important to close
these subprocesses properly. To achieve this, the current keyring interface should be extended with
a `Close()` method.

Some work may be needed to provide a way to migrate keys between the current keyring implementation and 
this new one.

## Further Discussions


## References

* https://github.com/cosmos/cosmos-sdk/issues/14940
* [Keyring Plugins Poc Implementation](https://github.com/Zondax/keyringPoc)
* [Hashicorp plugins](https://github.com/hashicorp/go-plugin)

