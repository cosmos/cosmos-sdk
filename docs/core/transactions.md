<!--
order: 2
-->

# Transactions

`Transactions` are objects created by end-users to trigger state changes in the application. {synopsis}

## Pre-requisite Readings

- [Anatomy of an SDK Application](../basics/app-anatomy.md) {prereq}

## Transactions

Transactions are comprised of metadata held in [contexts](./context.md) and [`Msg`s](../building-modules/messages-and-queries.md) that trigger state changes within a module through the module's [`Msg` service](../building-modules/msg-services.md).

When users want to interact with an application and make state changes (e.g. sending coins), they create transactions. Each of a transaction's `Msg`s must be signed using the private key associated with the appropriate account(s), before the transaction is broadcasted to the network. A transaction must then be included in a block, validated, and approved by the network through the consensus process. To read more about the lifecycle of a transaction, click [here](../basics/tx-lifecycle.md).

## Type Definition

Transaction objects are SDK types that implement the `Tx` interface

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/tx_msg.go#L49-L57

It contains the following methods:

- **GetMsgs:** unwraps the transaction and returns a list of its `Msg`s - one transaction may have one or multiple [`Msg`s](../building-modules/messages-and-queries.md#messages), which are defined by module developers.
- **ValidateBasic:** includes lightweight, [_stateless_](../basics/tx-lifecycle.md#types-of-checks) checks used by ABCI messages [`CheckTx`](./baseapp.md#checktx) and [`DeliverTx`](./baseapp.md#delivertx) to make sure transactions are not invalid. For example, the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth) module's `StdTx` `ValidateBasic` function checks that its transactions are signed by the correct number of signers and that the fees do not exceed what the user's maximum. Note that this function is to be distinct from the `ValidateBasic` functions for `Msg`s, which perform basic validity checks on messages only. For example, when [`runTx`](./baseapp.md#runtx) is checking a transaction created from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth/spec) module, it first runs `ValidateBasic` on each message, then runs the `auth` module AnteHandler which calls `ValidateBasic` for the transaction itself.

As a developer, you should rarely manipulate `Tx` directly, as `Tx` is really an intermediate type used for transaction generation. Instead, developers should prefer the `TxBuilder` interface, which you can learn more about [below](#transaction-generation).

### Signing Transactions

Every message in a transaction must be signed by the addresses specified by its `GetSigners`. The SDK currently allows signing transactions in two different ways.

#### `SIGN_MODE_DIRECT` (preferred)

The most used implementation of the `Tx` interface is the Protobuf `Tx` message, which is used in `SIGN_MODE_DIRECT`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/proto/cosmos/tx/v1beta1/tx.proto#L12-L25

Because Protobuf serialization is not deterministic, the SDK uses an additional `TxRaw` type to denote the pinned bytes over which a transaction is signed. Any user can generate a valid `body` and `auth_info` for a transaction, and serialize these two messages using Protobuf. `TxRaw` then pins the user's exact binary representation of `body` and `auth_info`, called respectively `body_bytes` and `auth_info_bytes`. The document that is signed by all signers of the transaction is `SignDoc` (deterministically serialized using [ADR-027](../architecture/adr-027-deterministic-protobuf-serialization.md)):

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/proto/cosmos/tx/v1beta1/tx.proto#L47-L64

Once signed by all signers, the `body_bytes`, `auth_info_bytes` and `signatures` are gathered into `TxRaw`, whose serialized bytes are broadcasted over the network.

#### `SIGN_MODE_LEGACY_AMINO_JSON`

The legacy implemention of the `Tx` interface is the `StdTx` struct from `x/auth`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/x/auth/legacy/legacytx/stdtx.go#L120-L130

The document signed by all signers is `StdSignDoc`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/x/auth/legacy/legacytx/stdsign.go#L20-L33

which is encoded into bytes using Amino JSON. Once all signatures are gathered into `StdTx`, `StdTx` is serialized using Amino JSON, and these bytes are broadcasted over the network.

#### Other Sign Modes

Other sign modes, most notably `SIGN_MODE_TEXTUAL`, are being discussed. If you wish to learn more about them, please refer to [ADR-020](../architecture/adr-020-protobuf-transaction-encoding.md).

## Transaction Process

The process of an end-user sending a transaction is:

- decide on the messages to put into the transaction,
- generate the transaction using the SDK's `TxBuilder`,
- broadcast the transaction using one of the available interfaces.

The next paragraphs will describe each of these components, in this order.

### Messages

::: tip
Module `Msg`s are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between the Tendermint and application layers.
:::

**Messages** (or `Msg`s) are module-specific objects that trigger state transitions within the scope of the module they belong to. Module developers define the messages for their module by adding methods to the Protobuf [`Msg` service](../building-modules/msg-services.md), and also implement the corresponding `MsgServer`.

`Msg`s in a module are defined as methods in the [`Msg` service] inside each module's `tx.proto` file. Since `Msg`s are module-specific, each module needs a to process all of its message types and trigger state changes within the module's scope. This design puts more responsibility on module developers, allowing application developers to reuse common functionalities without having to implement state transition logic repetitively. To achieve this, Protobuf generates a `MsgServer` interface for each module, and the module developer needs to implement this interface. The methods on the `MsgServer` interface corresponds on how to handle each of the different `Msg`.

To learn more about `Msg` services and how to implement `MsgServer`, click [here](../building-modules/msg-services.md).

While messages contain the information for state transition logic, a transaction's other metadata and relevant information are stored in the `TxBuilder` and `Context`.

### Transaction Generation

The `TxBuilder` interface contains data closely related with the generation of transactions, which an end-user can freely set to generate the desired transaction:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/client/tx_config.go#L32-L45

- `Msg`s, the array of [messages](#messages) included in the transaction.
- `GasLimit`, option chosen by the users for how to calculate how much gas they will need to pay.
- `Memo`, to send with the transaction.
- `FeeAmount`, the maximum amount the user is willing to pay in fees.
- `TimeoutHeight`, block height until which the transaction is valid.
- `Signatures`, the array of signatures from all signers of the transaction.

As there are currently two sign modes for signing transactions, there are also two implementations of `TxBuilder`:

- [wrapper](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/x/auth/tx/builder.go#L19-L33) for creating transactions for `SIGN_MODE_DIRECT`,
- [StdTxBuilder](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/x/auth/legacy/legacytx/stdtx_builder.go#L14-L20) for `SIGN_MODE_LEGACY_AMINO_JSON`.

However, the two implementation of `TxBuilder` should be hidden away from end-users, as they should prefer using the overarching `TxConfig` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/client/tx_config.go#L21-L30

`TxConfig` is an app-wide configuration for managing transactions. Most importantly, it holds the information about whether to sign each transaction with `SIGN_MODE_DIRECT` or `SIGN_MODE_LEGACY_AMINO_JSON`. By calling `txBuilder := txConfig.NewTxBuilder()`, a new `TxBuilder` will be created with the appropriate sign mode.

Once `TxBuilder` is correctly populated with the setters exposed above, `TxConfig` will also take care of correctly encoding the bytes (again, either using `SIGN_MODE_DIRECT` or `SIGN_MODE_LEGACY_AMINO_JSON`). Here's a pseudo-code snippet of how to generate and encode a transaction, using the `TxEncoder()` method:

```go
txBuilder := txConfig.NewTxBuilder()
txBuilder.SetMsgs(...) // and other setters on txBuilder

bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
// bz are bytes to be broadcasted over the network
```

### Broadcasting the Transaction

Once the transaction bytes are generated, there are currently three ways of broadcasting it.

#### CLI

Application developers create entrypoints to the application by creating a [command-line interface](../core/cli.md), [gRPC and/or REST interface](../core/grpc_rest.md), typically found in the application's `./cmd` folder. These interfaces allow users to interact with the application through command-line.

For the [command-line interface](../building-modules/module-interfaces.md#cli), module developers create subcommands to add as children to the application top-level transaction command `TxCmd`. CLI commands actually bundle all the steps of transaction processing into one simple command: creating messages, generating transactions and broadcasting. For concrete examples, see the [Interacting with a Node](../run-node/interact-node.md) section. An example transaction made using CLI looks like:

```bash
simd tx send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake
```

#### gRPC

[gRPC](https://grpc.io) is introduced in Cosmos SDK 0.40 as the main component for the SDK's RPC layer. The principal usage of gRPC is in the context of modules' [`Query` services](../building-modules). However, the SDK also exposes a few other module-agnostic gRPC services, one of them being the `Tx` service:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/proto/cosmos/tx/v1beta1/service.proto

The `Tx` service exposes a handful of utility functions, such as simulating a transaction or querying a transaction, and also one method to broadcast transactions.

Examples of broadcasting and simulating a transaction are shown [here](../run-node/txs.md#programmatically-with-go).

#### REST

Each gRPC method has its corresponding REST endpoint, generated using [gRPC-gateway](https://github.com/grpc-ecosystem/grpc-gateway). Therefore, instead of using gRPC, you can also use HTTP to broadcast the same transaction, on the `POST /cosmos/tx/v1beta1/txs` endpoint.

An example can be seen [here](../run-node/txs.md#using-rest)

#### Tendermint RPC

The three methods presented above are actually higher abstractions over the Tendermint RPC `/broadcast_tx_{async,sync,commit}` endpoints, documented [here](https://docs.tendermint.com/master/rpc/#/Tx). This means that you can use the Tendermint RPC endpoints directly to broadcast the transaction, if you wish so.

## Next {hide}

Learn about the [context](./context.md) {hide}
