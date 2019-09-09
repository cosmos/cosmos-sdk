# Messages and Queries

## Pre-requisite Reading

- [Introduction to SDK Modules](./intro.md)

## Synopsis

`Message`s and `Queries` are the two primary objects handled by modules. Most of the core components defined in a module, like `handler`s, `keeper`s and `querier`s, exist to process `message`s and `queries`. 

- [Messages](#messages)
- [Queries](#queries)

## Messages

`Message`s are objects whose end-goal is to trigger state-transitions. They are wrapped in [transactions], which may contain one or multiple of them. 

When a transaction is relayed from the underlying consensus engine to the SDK application, it is first decoded by [`baseapp`](../basics/baseapp.md). Then, each `message` contained in the transaction is extracted and routed to the appropriate module so that it can be processed by the module's [`handler`](./handler.md). For a more detailed explanation of the lifecycle of a transaction, click [here](../basics/tx-lifecycle.md). 

Defining `message`s is the responsibility of module developers. Typically, they are defined in a `types/msgs.go` file inside the module's folder. `message`s need to implement the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/master/types/tx_msg.go#L7-L29) interface, which contains the following methods:

- `Route() string`: Name of the route for this message. Typically all `message`s in a module have the same route, which is most often the module's name.
- `Type() string`: Type of the message, used primarly in [events](./events.md). This should return a message-specific `string`, typically the denomination of the message itself.
- `ValidateBasic() Error`: 

## Queries