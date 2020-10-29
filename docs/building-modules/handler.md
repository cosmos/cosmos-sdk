<!--
order: 4
-->

# Handlers

A `Handler` designates a function that processes [`message`s](./messages-and-queries.md#messages). `Handler`s are specific to the module in which they are defined, and only process `message`s defined within the said module. They are called from `baseapp` during [`DeliverTx`](../core/baseapp.md#delivertx). {synopsis}

## Pre-requisite Readings

- [Module Manager](./module-manager.md) {prereq}
- [Messages and Queries](./messages-and-queries.md) {prereq}

## `handler` type

The `handler` type defined in the Cosmos SDK specifies the typical structure of a `handler` function.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/handler.go#L4

Let us break it down:

- The [`Msg`](./messages-and-queries.md#messages) is the actual object being processed. 
- The [`Context`](../core/context.md) contains all the necessary information needed to process the `msg`, as well as a cache-wrapped copy of the latest state. If the `msg` is succesfully processed, the modified version of the temporary state contained in the `ctx` will be written to the main state.
- The [`*Result`] returned to `baseapp`, which contains (among other things) information on the execution of the `handler` and [`events`](../core/events.md).
	+++ https://github.com/cosmos/cosmos-sdk/blob/d55c1a26657a0af937fa2273b38dcfa1bb3cff9f/proto/cosmos/base/abci/v1beta1/abci.proto#L81-L95

## Implementation of a module `handler`s

Module `handler`s are typically implemented in a `./handler.go` file inside the module's folder. The [module manager](./module-manager.md) is used to add the module's `handler`s to the
[application's `router`](../core/baseapp.md#message-routing) via the `Route()` method. Typically,
the manager's `Route()` method simply constructs a Route that calls a `NewHandler()` method defined in `handler.go`,

### Using [`Msg` Services](messages-and-queries.md#msg-services)

Here's an example of a `NewHandler` from `x/bank` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc1/x/bank/handler.go#L10-L30

All `Msg` processing is done by a `MsgServer`](messages-and-queries.md#msg-services) protobuf service. Each module should define service, which will be responsible for request and response serialization. Each service defines an RPC which we will use in the `Handler`.  
When possible, the existing module's [`Keeper`](keeper.md) should implement `MsgServer`, else a `msgServer` struct that embeds the `Keeper` can be created, typically in  `./keeper/msg_server.go`.

`Handler` dispatches a `Msg` to appropriate `MsgServer` RPC function, usually by using a switch statement:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc1/x/bank/keeper/msg_server.go#L14-L16

First, the `handler` function sets a new `EventManager` to the context to isolate events per `msg`.
Then, a simple switch calls the appropriate `msgServer` method based on the `Msg` type. These methods are the ones that actually process `message`s. They can retrieve the `sdk.Context` from the `context.Context` parameter method using the `sdk.UnwrapSDKContext`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc1/x/bank/keeper/msg_server.go#L27

They usually follow the following 2 steps:

- First, they perform *stateful* checks to make sure the `message` is valid. At this stage, the `message`'s `ValidateBasic()` method has already been called, meaning *stateless* checks on the message (like making sure parameters are correctly formatted) have already been performed. Checks performed in the `handler` can be more expensive and require access to the state. For example, a `handler` for a `transfer` message might check that the sending account has enough funds to actually perform the transfer. To access the state, the `handler` needs to call the [`keeper`'s](./keeper.md) getter functions. 
- Then, if the checks are successfull, the `handler` calls the [`keeper`'s](./keeper.md) setter functions to actually perform the state transition. 

Before returning, `msgServer` methods generally emit one or multiple [`events`](../core/events.md) via the `EventManager` held in the `ctx`:

```go
ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,  // e.g. sdk.EventTypeMessage for a message, types.CustomEventType for a custom event defined in the module
			sdk.NewAttribute(attributeKey, attributeValue),
		),
    )
```

These `events` are relayed back to the underlying consensus engine and can be used by service providers to implement services around the application. Click [here](../core/events.md) to learn more about `events`. 

Finally, the invoked `msgServer` method returns a `proto.Message` response and an `error` which are then wrapped in `*sdk.Result` which contains the aforementioned `events` and an optional `Data` field, using `sdk.WrapServiceResult` function. 

+++ https://github.com/cosmos/cosmos-sdk/blob/d55c1a26657a0af937fa2273b38dcfa1bb3cff9f/proto/cosmos/base/abci/v1beta1/abci.proto#L81-L95

The `handler` can then be registered from [`AppModule.Route()`](./module-manager.md#appmodule) as shown in the example below:

+++ https://github.com/cosmos/cosmos-sdk/blob/228728cce2af8d494c8b4e996d011492139b04ab/x/gov/module.go#L143-L146

### Using [Legacy `Msg`s](messages-and-queries.md#legacy-msgs)

In this case, `handler`s functions need to be implemented for each module `Msg` and should be used in `NewHandler` in the place of `msgServer` methods.
`handler`s functions should return a `*Result` and an `error`.

## Telemetry

New [telemetry metrics](../core/telemetry.md) can be created from the `handler` when handling messages for instance. 

This is an example from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/d55c1a26657a0af937fa2273b38dcfa1bb3cff9f/x/auth/vesting/handler.go#L68-L80

## Next {hide}

Learn about [query services](./query-services.md) {hide}
