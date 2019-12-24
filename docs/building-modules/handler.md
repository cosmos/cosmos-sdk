<!--
order: 4
synopsis: "A `Handler` designates a function that processes [`message`s](./messages-and-queries.md#messages). `Handler`s are specific to the module in which they are defined, and only process `message`s defined within the said module. They are called from `baseapp` during [`DeliverTx`](../core/baseapp.md#delivertx)."
-->

# Handlers

## Pre-requisite Readings {hide}

- [Module Manager](./module-manager.md) {prereq}
- [Messages and Queries](./messages-and-queries.md) {prereq}

## `handler` type

The `handler` type defined in the Cosmos SDK specifies the typical structure of a `handler` function.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/handler.go#L4

Let us break it down:

- The [`Msg`](./messages-and-queries.md#messages) is the actual object being processed. 
- The [`Context`](../core/context.md) contains all the necessary information needed to process the `msg`, as well as a cache-wrapped copy of the latest state. If the `msg` is succesfully processed, the modified version of the temporary state contained in the `ctx` will be written to the main state.
- The [`Result`] returned to `baseapp`, which contains (among other things) information on the execution of the `handler`, [`gas`](../basics/gas-fees.md) consumption and [`events`](../core/events.md).
	+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/result.go#L15-L40

## Implementation of a module `handler`s

Module `handler`s are typically implemented in a `./handler.go` file inside the module's folder. The
[module manager](./module-manager.md) is used to add the module's `handler`s to the
[application's `router`](../core/baseapp.md#message-routing) via the `NewHandler()` method. Typically,
the manager's `NewHandler()` method simply calls a `NewHandler()` method defined in `handler.go`,
which looks like the following:

```go
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgType1:
			return handleMsgType1(ctx, keeper, msg)

		case MsgType2:
			return handleMsgType2(ctx, keeper, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}
```

This simple switch returns a `handler` function specific to the type of the received `message`. These `handler` functions are the ones that actually process `message`s, and usually follow the following 2 steps:

- First, they perform *stateful* checks to make sure the `message` is valid. At this stage, the `message`'s `ValidateBasic()` method has already been called, meaning *stateless* checks on the message (like making sure parameters are correctly formatted) have already been performed. Checks performed in the `handler` can be more expensive and require access to the state. For example, a `handler` for a `transfer` message might check that the sending account has enough funds to actually perform the transfer. To access the state, the `handler` needs to call the [`keeper`'s](./keeper.md) getter functions. 
- Then, if the checks are successfull, the `handler` calls the [`keeper`'s](./keeper.md) setter functions to actually perform the state transition. 

Before returning, `handler` functions generally emit one or multiple [`events`](../core/events.md) via the `EventManager` held in the `ctx`:

```go
ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,  // e.g. sdk.EventTypeMessage for a message, types.CustomEventType for a custom event defined in the module
			sdk.NewAttribute(attributeKey, attributeValue),
		),
    )
```

These `events` are relayed back to the underlying consensus engine and can be used by service providers to implement services around the application. Click [here](../core/events.md) to learn more about `events`. 

Finally, the `handler` function returns a `sdk.Result` which contains the aforementioned `events` and an optional `Data` field. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/result.go#L15-L40

Next is an example of how to return a `Result` from the `gov` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/gov/handler.go#L59-L62

For a deeper look at `handler`s, see this [example implementation of a `handler` function](https://github.com/cosmos/sdk-application-tutorial/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/x/nameservice/handler.go) from the nameservice tutorial. 

## Next {hide}

Learn about [queriers](./querier.md) {hide}
