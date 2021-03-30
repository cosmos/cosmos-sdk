<!--
order: 9
-->

# Events

`Event`s are objects that contain information about the execution of the application. They are mainly used by service providers like block explorers and wallet to track the execution of various messages and index transactions. {synopsis}

## Pre-requisite Readings

- [Anatomy of an SDK application](../basics/app-anatomy.md) {prereq}
- [Tendermint Documentation on Events](https://docs.tendermint.com/master/spec/abci/abci.html#events) {prereq}

## Events

Events are implemented in the Cosmos SDK as an alias of the ABCI `Event` type and
take the form of: `{eventType}.{attributeKey}={attributeValue}`.

+++ https://github.com/tendermint/tendermint/blob/v0.34.8/proto/tendermint/abci/types.proto#L304-L313

An Event contains:

- A `type` to categorize the Event at a high-level; for example, the SDK uses the `"message"` type to filter Events by `Msg`s.
- A list of `attributes` are key-value pairs that give more information about the categorized Event. For example, for the `"message"` type, we can filter Events by key-value pairs using `message.action={some_action}`, `message.module={some_module}` or `message.sender={some_sender}`.

::: tip
To parse the attribute values as strings, make sure to add `'` (single quotes) around each attribute value.
:::

Events, the `type` and `attributes` are defined on a **per-module basis** in the module's
`/types/events.go` file, and triggered from the module's [`Msg` service](../building-modules/msg-services.md)
by using the [`EventManager`](#eventmanager). In addition, each module documents its Events under
`spec/xx_events.md`.

Events are returned to the underlying consensus engine in the response of the following ABCI messages:

- [`BeginBlock`](./baseapp.md#beginblock)
- [`EndBlock`](./baseapp.md#endblock)
- [`CheckTx`](./baseapp.md#checktx)
- [`DeliverTx`](./baseapp.md#delivertx)

### Examples

The following examples show how to query Events using the SDK.

| Event                                            | Description                                                                                                                                              |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `tx.height=23`                                   | Query all transactions at height 23                                                                                                                      |
| `message.action='/cosmos.bank.v1beta1.Msg/Send'` | Query all transactions containing a x/bank `Send` [Service `Msg`](../building-modules/msg-services.md). Note the `'`s around the value.                  |
| `message.action='send'`                          | Query all transactions containing a x/bank `Send` [legacy `Msg`](../building-modules/msg-services.md#legacy-amino-msgs). Note the `'`s around the value. |
| `message.module='bank'`                          | Query all transactions containing messages from the x/bank module. Note the `'`s around the value.                                                       |
| `create_validator.validator='cosmosval1...'`     | x/staking-specific Event, see [x/staking SPEC](../../../cosmos-sdk/x/staking/spec/07_events.md).                                                         |

## EventManager

In Cosmos SDK applications, Events are managed by an abstraction called the `EventManager`.
Internally, the `EventManager` tracks a list of Events for the entire execution flow of a
transaction or `BeginBlock`/`EndBlock`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/events.go#L17-L25

The `EventManager` comes with a set of useful methods to manage Events. The method
that is used most by module and application developers is `EmitEvent` that tracks
an Event in the `EventManager`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/events.go#L33-L37

Module developers should handle Event emission via the `EventManager#EmitEvent` in each message
`Handler` and in each `BeginBlock`/`EndBlock` handler. The `EventManager` is accessed via
the [`Context`](./context.md), where Event emission generally follows this pattern:

```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(eventType, sdk.NewAttribute(attributeKey, attributeValue)),
)
```

Module's `handler` function should also set a new `EventManager` to the `context` to isolate emitted Events per `message`:

```go
func NewHandler(keeper Keeper) sdk.Handler {
    return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
        ctx = ctx.WithEventManager(sdk.NewEventManager())
        switch msg := msg.(type) {
```

See the [`Msg` services](../building-modules/msg-services.md) concept doc for a more detailed
view on how to typically implement Events and use the `EventManager` in modules.

## Subscribing to Events

You can use Tendermint's [Websocket](https://docs.tendermint.com/master/tendermint-core/subscription.html#subscribing-to-events-via-websocket) to subscribe to Events by calling the `subscribe` RPC method:

```json
{
  "jsonrpc": "2.0",
  "method": "subscribe",
  "id": "0",
  "params": {
    "query": "tm.event='eventCategory' AND eventType.eventAttribute='attributeValue'"
  }
}
```

The main `eventCategory` you can subscribe to are:

- `NewBlock`: Contains Events triggered during `BeginBlock` and `EndBlock`.
- `Tx`: Contains Events triggered during `DeliverTx` (i.e. transaction processing).
- `ValidatorSetUpdates`: Contains validator set updates for the block.

These Events are triggered from the `state` package after a block is committed. You can get the
full list of Event categories [on the Tendermint Godoc page](https://godoc.org/github.com/tendermint/tendermint/types#pkg-constants).

The `type` and `attribute` value of the `query` allow you to filter the specific Event you are looking for. For example, a `transfer` transaction triggers an Event of type `Transfer` and has `Recipient` and `Sender` as `attributes` (as defined in the [`events.go` file of the `bank` module](https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/x/bank/types/events.go)). Subscribing to this Event would be done like so:

```json
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

## Typed Events (coming soon)

As previously described, Events are defined on a per-module basis. It is the responsibility of the module developer to define Event types and Event attributes. Except in the `spec/XX_events.md` file, these Event types and attributes are unfortunately not easily discoverable, so the SDK proposes to use Protobuf-defined [Typed Events](../architecture/adr-032-typed-events.md) for emitting and querying Events.

The Typed Events proposal has not yet been fully implemented. Documentation is not yet available.

## Next {hide}

Learn about SDK [telemetry](./telemetry.md) {hide}
