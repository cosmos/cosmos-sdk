<!--
order: 7
synopsis: "`Event`s are objects that contain information about the execution of the application. They are mainly used by service providers like block explorers and wallet to track the execution of various messages and index transactions."
-->

# Events

## Pre-Requisite Readings {hide}

- [Anatomy of an SDK application](../basics/app-anatomy.md) {prereq}

## Events

`Event`s are implemented in the Cosmos SDK as an alias of the ABCI `event` type. 

+++ https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/abci/types/types.pb.go#L2661-L2667

They contain:

- A **`type`** of type `string`, which can refer to the type of action that led to the `event`'s emission (e.g. a certain value going above a threshold), or to the type of `message` if the event is triggered at the end of that `message` processing. 
- A list of `attributes`, which are key-value pairs that give more information about the `event`. 
    +++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/events.go#L51-L56

`Event`s are returned to the underlying consensus engine in the response of the following ABCI messages: [`CheckTx`](./baseapp.md#checktx), [`DeliverTx`](./baseapp.md#delivertx), [`BeginBlock`](./baseapp.md#beginblock) and [`EndBlock`](./baseapp.md#endblock). 

Typically, `event` `type`s and `attributes` are defined on a **per-module basis** in the module's `/internal/types/events.go` file, and triggered from the module's [`handler`](../building-modules/handler.md) via the [`EventManager`](#eventmanager).

## EventManager

In Cosmos SDK applications, `event`s are generally managed by an object called the `EventManager`. It is implemented as a simple wrapper around a slice of `event`s: 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/events.go#L16-L20

The `EventManager` comes with a set of useful methods to manage `event`s. Among them, the one that is used the most by module and application developers is the `EmitEvent` method, which registers an `event` in the `EventManager`. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/events.go#L29-L31

Typically, module developers will implement event emission via the `EventManager` in the [`handler`](../building-modules/handler.md) of modules, as well as in the [`BeginBlocker` and/or`EndBlocker` functions](../building-modules/beginblock-endblock.md). The `EventManager` is accessed via the context [`ctx`](./context.md), and event emission generally follows this pattern:

```go
ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,  // e.g. sdk.EventTypeMessage for a message, types.CustomEventType for a custom event defined in the module
			sdk.NewAttribute(attributeKey, attributeValue),
		),
    )
```

See the [`handler` concept doc](../building-modules/handler.md) for a more detailed view on how to typically implement `events` and use the `EventManager` in modules. 

## Subscribing to `events`

It is possible to subscribe to `events` via [Tendermint's Websocket](https://tendermint.com/docs/app-dev/subscribing-to-events-via-websocket.html#subscribing-to-events-via-websocket). This is done by calling the `subscribe` RPC method via Websocket:

```
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "id": "0",
    "params": {
        "query": "tm.event='eventCategory' AND type.attribute='attributeValue'"
    }
}
```

The main `eventCategory` you can subscribe to are:

- `NewBlock`: Contains `events` triggered during `BeginBlock` and `EndBlock`.
- `Tx`: Contains `events` triggered during `DeliverTx` (i.e. transaction processing).
- `ValidatorSetUpdates`: Contains validator set updates for the block. 

These events are triggered from the `state` package after a block is committed. You can get the full list of `event` categories [here](https://godoc.org/github.com/tendermint/tendermint/types#pkg-constants). 

The `type` and `attribute` value of the `query` allow you to filter the specific `event` you are looking for. For example, a `transfer` transaction triggers an `event` of type `Transfer` and has `Recipient` and `Sender` as `attributes` (as defined in the [`events` file of the `bank` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/bank/internal/types/events.go)). Subscribing to this `event` would be done like so:

```
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "id": "0",
    "params": {
        "query": "tm.event='Tx' AND transfer.sender='senderAddress'"
    }
}
```

where `senderAddress` is an address following the [`AccAddress`](../basics/accounts.md#addresses) format. 

## Next {hide}

Learn about [object-capabilities](./ocap.md) {hide}