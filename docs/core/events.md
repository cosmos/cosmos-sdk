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

- A `type`, which is meant to categorize an event at a high-level, e.g. the SDK uses the `"message"` type to filter events by `Msg`s.
- A list of `attributes`, which are key-value pairs that give more information about the categorized Event. For example, for the `"message"` type, we can filter events by key-value pairs using `message.action={some_action}`, `message.module={some_module}` or `message.sender={some_sender}`.

::: tip
To parse the attribute values as strings, make sure to add `'` (single quotes) around each attribute value.
:::

Events, the `type` and `attributes` are defined on a **per-module basis** in the module's
`/types/events.go` file, and triggered from the module's [`Msg` service](../building-modules/msg-services.md)
by using the [`EventManager`](#eventmanager). In addition, each module documents its events under
`spec/xx_events.md`.

Events are returned to the underlying consensus engine in the response of the following ABCI messages:

- [`BeginBlock`](./baseapp.md#beginblock)
- [`EndBlock`](./baseapp.md#endblock)
- [`CheckTx`](./baseapp.md#checktx)
- [`DeliverTx`](./baseapp.md#delivertx)

### Examples

Below are some examples of Events you can query using the SDK.

| Event                                            | Description                                                                                                                                              |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `tx.height=23`                                   | Query all transactions at height 23                                                                                                                      |
| `message.action='/cosmos.bank.v1beta1.Msg/Send'` | Query all transactions containing a x/bank `Send` [Service `Msg`](../building-modules/msg-services.md). Note the `'`s around the value.                  |
| `message.action='send'`                          | Query all transactions containing a x/bank `Send` [legacy `Msg`](../building-modules/msg-services.md#legacy-amino-msgs). Note the `'`s around the value. |
| `message.module='bank'`                          | Query all transactions containing messages from the x/bank module. Note the `'`s around the value.                                                       |
| `create_validator.validator='cosmosval1...'`     | x/staking-specific event, see [x/staking SPEC](../../../cosmos-sdk/x/staking/spec/07_events.md).                                                         |

## EventManager

In Cosmos SDK applications, events are managed by an abstraction called the `EventManager`.
Internally, the `EventManager` tracks a list of `Events` for the entire execution flow of a
transaction or `BeginBlock`/`EndBlock`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/events.go#L17-L25

The `EventManager` comes with a set of useful methods to manage events. The method
that is used most by module and application developers is `EmitEvent` that tracks
an `event` in the `EventManager`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/events.go#L33-L37

Module developers should handle event emission via the `EventManager#EmitEvent` in each message
`Handler` and in each `BeginBlock`/`EndBlock` handler. The `EventManager` is accessed via
the [`Context`](./context.md), where event emission generally follows this pattern:

```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(eventType, sdk.NewAttribute(attributeKey, attributeValue)),
)
```

Module's `handler` function should also set a new `EventManager` to the `context` to isolate emitted events per `message`:

```go
func NewHandler(keeper Keeper) sdk.Handler {
    return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
        ctx = ctx.WithEventManager(sdk.NewEventManager())
        switch msg := msg.(type) {
```

See the [`Msg` services](../building-modules/msg-services.md) concept doc for a more detailed
view on how to typically implement `Events` and use the `EventManager` in modules.

## Subscribing to Events

You can use Tendermint's [Websocket](https://docs.tendermint.com/master/tendermint-core/subscription.html#subscribing-to-events-via-websocket) to subscribe to `Events` by calling the `subscribe` RPC method:

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

- `NewBlock`: Contains `events` triggered during `BeginBlock` and `EndBlock`.
- `Tx`: Contains `events` triggered during `DeliverTx` (i.e. transaction processing).
- `ValidatorSetUpdates`: Contains validator set updates for the block.

These events are triggered from the `state` package after a block is committed. You can get the
full list of `event` categories [here](https://godoc.org/github.com/tendermint/tendermint/types#pkg-constants).

The `type` and `attribute` value of the `query` allow you to filter the specific `event` you are looking for. For example, a `transfer` transaction triggers an `event` of type `Transfer` and has `Recipient` and `Sender` as `attributes` (as defined in the [`events` file of the `bank` module](https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/x/bank/types/events.go)). Subscribing to this `event` would be done like so:

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

As described above, events are defined on a per-module basis, and it is the responsability of the module developer to define event types and event attributes. Except in the `spec/XX_events.md` file, these types and attributes are unfortunately not easily discoverable, so the SDK proposes to use Protobuf-defined [Typed Events](../architecture/adr-032-typed-events.md) for emitting and querying for events.

The Typed Events proposal has not yet been fully implemented, and its documentation will follow with a future release of the SDK.

## Next {hide}

Learn about SDK [telemetry](./telemetry.md) {hide}
