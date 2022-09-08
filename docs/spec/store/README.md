# Store

The store package defines the interfaces, types and abstractions for Cosmos SDK
modules to read and write to merkleized state within a Cosmos SDK application.
The store package provides many primitives for developers to use in order to
work with state. Below we describe the various abstractions.

## Types

### Store

The bulk of the store interfaces are defined [here](https://github.com/cosmos/cosmos-sdk/blob/main/store/types/store.go),
where the base primitive interface, for which other interfaces build off of, is
the `Store` type. The `Store` interface defines the ability to tell the type of
the implementing store and the ability to cache wrap via the `CacheWrapper` interface.

### CacheWrapper
