# Handlers

## Pre-requisite Reading

- [Module Manager](./module-manager.md)
- [Messages and Queries](./messages-and-queries.md)

## Synopsis

A `Handler` designates a function that processes [`message`s](./messages-and-queries.md#messages). `handler`s are specific to the module in which they are defined, and only process `message`s defined within said module. They are called from `baseapp` during [`DeliverTx`](../core/baseapp.md#delivertx).

- [`handler` type](#handler-type)
- [Implementation of a module `handler`s](#implementation-of-a-module-handlers)

## `handler` type

The [`handler` type](https://github.com/cosmos/cosmos-sdk/blob/master/types/handler.go#L4) defined in the Cosmos SDK specifies the typical structure of a `handler` function:

```go
type Handler func(ctx Context, msg Msg) Result
```

Let us break it down:

- The [`msg`](./messages-and-queries.md#messages) is the actual object being processed. 
- The [`Context`](../core/context.md) contains all the necessary information needed to process the `msg`, as well as a cache-wrapped copy of the latest state. If the `msg` is succesfully processed, the modified version of the temporary state contained in the `ctx` will be written to the main state.
- The [`Result`](https://github.com/cosmos/cosmos-sdk/blob/master/types/result.go#L14-L38) returned to `baseapp`, which contains (among other things) information on the execution of the `handler`, [`gas`](../basics/accounts-fees-gas.md#gas) consumption and [`events`](./events.md).

## Implementation of a module `handler`s

Module `handler`s are typically implemented in a `handler.go` file inside the module's folder. The [module manager](./module-manager.md) is used to add the module's `handler`s to the [application's `router`](../core/baseapp.md#message-routing) via the `NewHandler()` method. Typically, the manager's `NewHandler()` method simply calls a `NewHandler()` method defined in `handler.go`, which looks like the following:

```go
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgType1:
			return handleMsgType1(ctx, keeper, msg)
		case MsgType2:
			return handleMsgType2(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
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

Finally, the `handler` function returns a [`sdk.Result`](https://github.com/tendermint/tendermint/blob/master/abci/types/result.go) which contains the aforementioned `events` and an optional `Data` field. 

For a deeper look at `handler`s, see this [example implementation of a `handler` function](https://github.com/cosmos/sdk-application-tutorial/blob/master/x/nameservice/handler.go) from the nameservice tutorial. 

## Next

Learn about [queriers](./querier.md). 