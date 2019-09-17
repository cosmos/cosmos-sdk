# Events

## Pre-Requisite Reading

- [Anatomy of an SDK application](../basics/app-anatomy.md)

## Synopsis

`Event`s are objects that contain information about the execution of the application. They are mainly used by service providers like block explorers and wallet to track the execution of various messages and index transactions. 

- [Events](#events)
- [EventManager](#eventmanager)

## Events

`Event`s are implemented in the Cosmos SDK as an alias of the [ABCI `event` type](https://github.com/tendermint/tendermint/blob/master/abci/types/types.pb.go#L2661-L2667). They contain:

- A **type** of type `string`, which can refer to the type of action that led to the `event`'s emission (e.g. a certain value going above a threshold), or to the type of `message` if the event is triggered at the end of that `message` processing. 
- A list of [`attributes`](https://github.com/cosmos/cosmos-sdk/blob/master/types/events.go#L53-L56), which are key-value pairs that give more information about the `event`. 

`Event`s are returned to the underlying consensus engine in the response of the following ABCI messages: [`CheckTx`](./baseapp.md#checktx), [`DeliverTx`](./baseapp.md#delivertx), [`BeginBlock`](./baseapp.md#beginblock) and [`EndBlock`](./baseapp.md#endblock). 

## EventManager

In Cosmos SDK applications, `event`s are generally managed by an object called the [`EventManager`](https://github.com/cosmos/cosmos-sdk/blob/master/types/events.go#L18-L20). It is implemented as a simple wrapper around a slice of `event`s: 

```go
type EventManager struct {
	events Events
}
```

The `EventManager` comes with a set of useful methods to manage `event`s. Among them, the one that is used the most by module and application developers is the [`EmitEvent`](https://github.com/cosmos/cosmos-sdk/blob/master/types/events.go#L29-L31) method, which registers an `event` in the `EventManager`. 

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

## Next

Learn about [encoding](./encoding.md)