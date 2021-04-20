<!--
order: 5
-->

# Gas and Fees

This document describes the default strategies to handle gas and fees within a Cosmos SDK application. {synopsis}

### Pre-requisite Readings

- [Anatomy of an SDK Application](./app-anatomy.md) {prereq}

## Introduction to `Gas` and `Fees`

In the Cosmos SDK, `gas` is a special unit that is used to track the consumption of resources during execution. `gas` is typically consumed whenever read and writes are made to the store, but it can also be consumed if expensive computation needs to be done. It serves two main purposes:

- Make sure blocks are not consuming too many resources and will be finalized. This is implemented by default in the SDK via the [block gas meter](#block-gas-meter).
- Prevent spam and abuse from end-user. To this end, `gas` consumed during [`message`](../building-modules/messages-and-queries.md#messages) execution is typically priced, resulting in a `fee` (`fees = gas * gas-prices`). `fees` generally have to be paid by the sender of the `message`. Note that the SDK does not enforce `gas` pricing by default, as there may be other ways to prevent spam (e.g. bandwidth schemes). Still, most applications will implement `fee` mechanisms to prevent spam. This is done via the [`AnteHandler`](#antehandler).

## Gas Meter

In the Cosmos SDK, `gas` is a simple alias for `uint64`, and is managed by an object called a _gas meter_. Gas meters implement the `GasMeter` interface

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/store/types/gas.go#L34-L43

where:

- `GasConsumed()` returns the amount of gas that was consumed by the gas meter instance.
- `GasConsumedToLimit()` returns the amount of gas that was consumed by gas meter instance, or the limit if it is reached.
- `Limit()` returns the limit of the gas meter instance. `0` if the gas meter is infinite.
- `ConsumeGas(amount Gas, descriptor string)` consumes the amount of `gas` provided. If the `gas` overflows, it panics with the `descriptor` message. If the gas meter is not infinite, it panics if `gas` consumed goes above the limit.
- `IsPastLimit()` returns `true` if the amount of gas consumed by the gas meter instance is strictly above the limit, `false` otherwise.
- `IsOutOfGas()` returns `true` if the amount of gas consumed by the gas meter instance is above or equal to the limit, `false` otherwise.

The gas meter is generally held in [`ctx`](../core/context.md), and consuming gas is done with the following pattern:

```go
ctx.GasMeter().ConsumeGas(amount, "description")
```

By default, the Cosmos SDK makes use of two different gas meters, the [main gas meter](#main-gas-metter[) and the [block gas meter](#block-gas-meter).

### Main Gas Meter

`ctx.GasMeter()` is the main gas meter of the application. The main gas meter is initialized in `BeginBlock` via `setDeliverState`, and then tracks gas consumption during execution sequences that lead to state-transitions, i.e. those originally triggered by [`BeginBlock`](../core/baseapp.md#beginblock), [`DeliverTx`](../core/baseapp.md#delivertx) and [`EndBlock`](../core/baseapp.md#endblock). At the beginning of each `DeliverTx`, the main gas meter **must be set to 0** in the [`AnteHandler`](#antehandler), so that it can track gas consumption per-transaction.

Gas consumption can be done manually, generally by the module developer in the [`BeginBlocker`, `EndBlocker`](../building-modules/beginblock-endblock.md) or [`Msg` service](../building-modules/msg-services.md), but most of the time it is done automatically whenever there is a read or write to the store. This automatic gas consumption logic is implemented in a special store called [`GasKv`](../core/store.md#gaskv-store).

### Block Gas Meter

`ctx.BlockGasMeter()` is the gas meter used to track gas consumption per block and make sure it does not go above a certain limit. A new instance of the `BlockGasMeter` is created each time [`BeginBlock`](../core/baseapp.md#beginblock) is called. The `BlockGasMeter` is finite, and the limit of gas per block is defined in the application's consensus parameters. By default Cosmos SDK applications use the default consensus parameters provided by Tendermint:

+++ https://github.com/tendermint/tendermint/blob/v0.34.0-rc6/types/params.go#L34-L41

When a new [transaction](../core/transactions.md) is being processed via `DeliverTx`, the current value of `BlockGasMeter` is checked to see if it is above the limit. If it is, `DeliverTx` returns immediately. This can happen even with the first transaction in a block, as `BeginBlock` itself can consume gas. If not, the transaction is processed normally. At the end of `DeliverTx`, the gas tracked by `ctx.BlockGasMeter()` is increased by the amount consumed to process the transaction:

```go
ctx.BlockGasMeter().ConsumeGas(
	ctx.GasMeter().GasConsumedToLimit(),
	"block gas meter",
)
```

## AnteHandler

The `AnteHandler` is run for every transaction during `CheckTx` and `DeliverTx`, before the `Msg` service of each `Msg` in the transaction. `AnteHandler`s have the following signature:

```go
// AnteHandler authenticates transactions, before their internal messages are handled.
// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx, simulate bool) (newCtx Context, result Result, abort bool)
```

The `anteHandler` is not implemented in the core SDK but in a module. This gives the possibility to developers to choose which version of `AnteHandler` fits their application's needs. That said, most applications today use the default implementation defined in the [`auth` module](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth). Here is what the `anteHandler` is intended to do in a normal Cosmos SDK application:

- Verify that the transaction are of the correct type. Transaction types are defined in the module that implements the `anteHandler`, and they follow the transaction interface:
  +++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/tx_msg.go#L49-L57
  This enables developers to play with various types for the transaction of their application. In the default `auth` module, the default transaction type is `Tx`:
  +++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/proto/cosmos/tx/v1beta1/tx.proto#L12-L25
- Verify signatures for each [`message`](../building-modules/messages-and-queries.md#messages) contained in the transaction. Each `message` should be signed by one or multiple sender(s), and these signatures must be verified in the `anteHandler`.
- During `CheckTx`, verify that the gas prices provided with the transaction is greater than the local `min-gas-prices` (as a reminder, gas-prices can be deducted from the following equation: `fees = gas * gas-prices`). `min-gas-prices` is a parameter local to each full-node and used during `CheckTx` to discard transactions that do not provide a minimum amount of fees. This ensure that the mempool cannot be spammed with garbage transactions.
- Verify that the sender of the transaction has enough funds to cover for the `fees`. When the end-user generates a transaction, they must indicate 2 of the 3 following parameters (the third one being implicit): `fees`, `gas` and `gas-prices`. This signals how much they are willing to pay for nodes to execute their transaction. The provided `gas` value is stored in a parameter called `GasWanted` for later use.
- Set `newCtx.GasMeter` to 0, with a limit of `GasWanted`. **This step is extremely important**, as it not only makes sure the transaction cannot consume infinite gas, but also that `ctx.GasMeter` is reset in-between each `DeliverTx` (`ctx` is set to `newCtx` after `anteHandler` is run, and the `anteHandler` is run each time `DeliverTx` is called).

As explained above, the `anteHandler` returns a maximum limit of `gas` the transaction can consume during execution called `GasWanted`. The actual amount consumed in the end is denominated `GasUsed`, and we must therefore have `GasUsed =< GasWanted`. Both `GasWanted` and `GasUsed` are relayed to the underlying consensus engine when [`DeliverTx`](../core/baseapp.md#delivertx) returns.

## Next {hide}

Learn about [baseapp](../core/baseapp.md) {hide}
