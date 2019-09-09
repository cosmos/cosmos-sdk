# Messages and Queries

## Pre-requisite Reading

- [Introduction to SDK Modules](./intro.md)

## Synopsis

`Message`s and `Queries` are the two primary objects handled by modules. Most of the core components defined in a module, like `handler`s, `keeper`s and `querier`s, exist to process `message`s and `queries`. 

- [Messages](#messages)
- [Queries](#queries)

## Messages

`Message`s are objects defined by modules, instantiated by end-users and included in [transactions](../core/transactions.md) in order to modify the state. 

## Queries