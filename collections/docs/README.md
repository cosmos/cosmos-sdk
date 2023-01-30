# Collections

Collections is a library meant to simplify the experience with respect to module state handling.

Cosmos-sdk modules handle their state using the `KVStore` interface. The problem with working with
KVStore is that it forces you to think of state as a bytes KV pairings when in reality the majority of
state comes from complex concrete golang objects (strings, ints, structs, etc.).

Collections allows you to work with state as if they were normal golang objects and removes the need
for you to think of your state as raw bytes in your code.

It also allows you to migrate your existing state without causing any state breakage that forces you into
tedious and complex chain state migrations.

## Installation

To install collections in your cosmos-sdk chain project, run the following command:

```shell
go get cosmossdk.io/collections
```

## Documentation

- [Intro](01-intro.md) - an introduction to collections and the basics.
- [Map](02-map.md) - docs and examples of `collections.Map`
- [KeySet](03-keyset.md) - docs and examples of `collections.KeySet`
- [Item](04-item.md) - docs and examples of `collections.Item`
- [Iteration](05-iterating.md) - docs and examples on how to iterate collections.
- [Composite keys](06-composite-keys.md) - docs and examples on working with composite keys (pairs)
- [IndexedMap](07-indexed-map.md) - docs and examples on how to work with `collections.IndexedMap`