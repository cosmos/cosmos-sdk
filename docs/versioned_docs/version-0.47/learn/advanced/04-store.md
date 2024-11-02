---
sidebar_position: 1
---

# Store

:::note Synopsis
A store is a data structure that holds the state of the application.
:::

:::note

### Pre-requisite Readings

* [Anatomy of a Cosmos SDK application](../beginner/00-overview-app.md)

:::

## Introduction

The Cosmos SDK store package provides interfaces, types, and abstractions for managing Merkleized state storage and commitment within a Cosmos SDK application. The package supplies various primitives for developers to work with, including state storage, state commitment, and wrapper KVStores. This document highlights the key abstractions and their significance.

## Multistore

The main store in Cosmos SDK applications is a multistore, a store of stores, that supports modularity. Developers can add any number of key-value stores to the multistore based on their application needs. Each module can declare and manage its own subset of the state, allowing for a modular approach. Key-value stores within the multistore can only be accessed with a specific capability key, which is typically held in the keeper of the module that declared the store.

## Store Interfaces

### KVStore

The `KVStore` interface defines a key-value store that can be used to store and retrieve data. The default implementation of `KVStore` used in `baseapp` is the `iavl.Store`, which is based on an IAVL Tree. KVStores can be accessed by objects that hold a specific key and can provide an `Iterator` method that returns an `Iterator` object, used to iterate over a range of keys.

### CommitKVStore

The `CommitKVStore` interface extends the `KVStore` interface and adds methods for state commitment. The default implementation of `CommitKVStore` used in `baseapp` is also the `iavl.Store`.

### StoreDB

The `StoreDB` interface defines a database that can be used to persist key-value stores. The default implementation of `StoreDB` used in `baseapp` is the `dbm.DB`, which is a simple persistent key-value store.

### DBAdapter

The `DBAdapter` interface defines an adapter for `dbm.DB` that fulfills the `KVStore` interface. This interface is used to provide compatibility between the `dbm.DB` implementation and the `KVStore` interface.

### TransientStore

The `TransientStore` interface defines a base-layer KVStore which is automatically discarded at the end of the block and is useful for persisting information that is only relevant per-block, like storing parameter changes.

## Store Abstractions

The store package provides a comprehensive set of abstractions for managing state commitment and storage in an SDK application. These abstractions include CacheWrapping, KVStore, and CommitMultiStore, which offer a range of features such as CRUD functionality, prefix-based iteration, and state commitment management.

By utilizing these abstractions, developers can create modular applications with independent state management for each module. This approach allows for a more organized and maintainable application structure.

### CacheWrap

CacheWrap is a wrapper around a KVStore that provides caching for both read and write operations. The CacheWrap can be used to improve performance by reducing the number of disk reads and writes required for state storage operations. The CacheWrap also includes a Write method that commits the pending writes to the underlying KVStore.

### HistoryStore

The HistoryStore is an optional feature that can be used to store historical versions of the state. The HistoryStore can be used to track changes to the state over time, allowing developers to analyze changes in the state and roll back to previous versions if necessary.

### IndexStore

The IndexStore is a type of KVStore that is used to maintain indexes of data stored in other KVStores. IndexStores can be used to improve query performance by providing a way to quickly search for data based on specific criteria.

### Queryable

The Queryable interface is used to provide a way for applications to query the state stored in a KVStore. The Queryable interface includes methods for retrieving data based on a key or a range of keys, as well as methods for retrieving data based on specific criteria.

### PrefixIterator

The PrefixIterator interface is used to iterate over a range of keys in a KVStore that share a common prefix. PrefixIterators can be used to efficiently retrieve subsets of data from a KVStore based on a specific prefix.

### RootMultiStore

The RootMultiStore is a Multistore that provides the ability to retrieve a snapshot of the state at a specific height. This is useful for implementing light clients.

### GasKVStore

The GasKVStore is a wrapper around a KVStore that provides gas measurement for read and write operations. The GasKVStore is typically used to measure the cost of executing transactions.

## Implementation Details

While there are many interfaces that the store package provides, there is typically a core implementation for each main interface that modules and developers interact with that are defined in the Cosmos SDK.

The `iavl.Store` provides the core implementation for state storage and commitment by implementing the following interfaces:

-   `KVStore`
-   `CommitStore`
-   `CommitKVStore`
-   `Queryable`
-   `StoreWithInitialVersion`

The `iavl.Store` also provides the ability to remove historical state from the state commitment layer.

An overview of the IAVL implementation can be found [here](https://github.com/cosmos/iavl/blob/master/docs/overview.md).

Other store abstractions include `cachekv.Store`, `gaskv.Store`, `cachemulti.Store`, and `rootmulti.Store`. Each of these stores provide additional functionality and abstractions for developers to work with.

Note that concurrent access to the `iavl.Store` tree is not safe, and it is the responsibility of the caller to ensure that concurrent access to the store is not performed.

## Store Migration

Store migration is the process of updating the structure of a KVStore to support new features or changes in the data model. Store migration can be a complex process, but it is essential for maintaining the integrity of the state stored in a KVStore.
