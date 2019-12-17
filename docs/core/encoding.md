<!--
order: 6
synopsis: The `codec` is used everywhere in the Cosmos SDK to encode and decode structs and interfaces. The specific codec used in the Cosmos SDK is called `go-amino`
-->

# Encoding

## Pre-requisite Readings {hide}

- [Anatomy of an SDK application](../basics/app-anatomy.md) {prereq}

## Encoding

Every Cosmos SDK application exposes a global `codec` to marshal/unmarshal structs and interfaces in order to store and/or transfer them. As of now, the `codec` used in the Cosmos SDK is [go-amino](https://github.com/tendermint/go-amino), which possesses the following important properties:

- Interface support. 
- Deterministic encoding of value (which is required considering that blockchains are deterministic replicated state-machines). 
- Upgradeable schemas. 

The application's `codec` is typically initialized in the [application's constructor function](../basics/app-anatomy.md#constructor-function), where it is also passed to each of the application's modules via the [basic manager](../building-modules/module-manager.md#basic-manager). 

Among other things, the `codec` is used by module's [`keeper`s](../building-modules/keeper.md) to marshal objects into `[]byte` before storing them in the module's [`KVStore`](./store.md#kvstore), or to unmarshal them from `[]byte` when retrieving them:

```go
// typical pattern to marshal an object to []byte before storing it
bz := keeper.cdc.MustMarshalBinaryBare(object)

//typical pattern to unmarshal an object from []byte when retrieving it
keeper.cdc.MustUnmarshalBinaryBare(bz, &object)
```

Alternatively, it is possible to use `MustMarshalBinaryLengthPrefixed`/`MustUnmarshalBinaryLengthPrefixed` instead of `MustMarshalBinaryBare`/`MustUnmarshalBinaryBare` for the same encoding prefixed by a `uvarint` encoding of the object to encode. 

Another important use of the `codec` is the encoding and decoding of [transactions](./transactions.md). Transactions are defined at the Cosmos SDK level, but passed to the underlying consensus engine in order to be relayed to other peers. Since the underlying consensus engine is agnostic to the application, it only accepts transactions in the form of `[]byte`. The encoding is done by an object called `TxEncoder` and the decoding by an object called `TxDecoder`. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/tx_msg.go#L45-L49

A standard implementation of both these objects can be found in the [`auth` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth):

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/auth/types/stdtx.go#L241-L266

## Next {hide}

Learn about [events](./events.md) {hide}