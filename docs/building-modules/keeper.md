<!--
order: 7
synopsis: "`Keeper`s refer to a Cosmos SDK abstraction whose role is to manage access to the subset of the state defined by various modules. `Keeper`s are module-specific, i.e. the subset of state defined by a module can only be accessed by a `keeper` defined in said module. If a module needs to access the subset of state defined by another module, a reference to the second module's internal `keeper` needs to be passed to the first one. This is done in `app.go` during the instantiation of module keepers."
-->

# Keepers

## Pre-requisite Readings {hide}

- [Introduction to SDK Modules](./intro.md) {prereq}

## Motivation

The Cosmos SDK is a framework that makes it easy for developers to build complex decentralised applications from scratch, mainly by composing modules together. As the ecosystem of open source modules for the Cosmos SDK expands, it will become increasingly likely that some of these modules contain vulnerabilities, as a result of the negligence or malice of their developer. 

The Cosmos SDK adopts an [object-capabilities-based approach](../core/ocap.md) to help developers better protect their application from unwanted inter-module interactions, and `keeper`s are at the core of this approach. A `keeper` can be thought of quite literally as the gatekeeper of a module's store(s). Each store (typically an [`IAVL` Store](../core/store.md#iavl-store)) defined within a module comes with a `storeKey`, which grants unlimited access to it. The module's `keeper` holds this `storeKey` (which should otherwise remain unexposed), and defines [methods](#implementing-methods) for reading and writing to the store(s). 

The core idea behind the object-capabilities approach is to only reveal what is necessary to get the work done. In practice, this means that instead of handling permissions of modules through access-control lists, module `keeper`s are passed a reference to the specific instance of the other modules' `keeper`s that they need to access (this is done in the [application's constructor function](../basics/app-anatomy.md#constructor-function)). As a consequence, a module can only interact with the subset of state defined in another module via the methods exposed by the instance of the other module's `keeper`. This is a great way for developers to control the interactions that their own module can have with modules developed by external developers. 

## Type Definition 

`keeper`s are generally implemented in a `internal/keeper/keeper.go` file located in the module's folder. By convention, the type `keeper` of a module is simply named `Keeper` and usually follows the following structure:

```go
type Keeper struct {
    // External keepers, if any

    // Store key(s)

    // codec
}
```

For example, here is the [type definition of the `keeper` from the nameservice tutorial:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/internal/keeper/keeper.go#L10-L17

Let us go through the different parameters:

- An expected `keeper` is a `keeper` external to a module that is required by the internal `keeper` of said module. External `keeper`s are listed in the internal `keeper`'s type definition as interfaces. These interfaces are themselves defined in a `internal/types/expected_keepers.go` file within the module's folder. In this context, interfaces are used to reduce the number of dependencies, as well as to facilitate the maintenance of the module itself. 
- `storeKey`s grant access to the store(s) of the [multistore](../core/store.md) managed by the module. They should always remain unexposed to external modules. 
- A [codec `cdc`](../core/encoding.md), used to marshall and unmarshall struct to/from `[]byte`. 

Of course, it is possible to define different types of internal `keeper`s for the same module (e.g. a read-only `keeper`). Each type of `keeper` comes with its own constructor function, which is called from the [application's constructor function](../basics/app-anatomy.md). This is where `keeper`s are instantiated, and where developers make sure to pass correct instances of modules' `keeper`s to other modules that require it. 

## Implementing Methods 

`Keeper`s primarily expose getter and setter methods for the store(s) managed by their module. These methods should remain as simple as possible and strictly be limited to getting or setting the requested value, as validity checks should have already been performed via the `ValidateBasic()` method of the [`message`](./messages-and-queries.md#messages) and the [`handler`](./handler.md) when `keeper`s' methods are called. 

Typically, a *getter* method will present with the following signature 

```go
func (k Keeper) Get(ctx sdk.Context, key string) returnType
```

and go through the following steps:

1. Retrieve the appropriate store from the `ctx` using the `storeKey`. This is done through the `KVStore(storeKey sdk.StoreKey` method of the `ctx`.
2. If it exists, get the `[]byte` value stored at location `[]byte(key)` using the `Get(key []byte)` method of the store. 
3. Unmarshall the retrieved value from `[]byte` to `returnType` using the codec `cdc`. Return the value.

Similarly, a *setter* method will present with the following signature 

```go
func (k Keeper) Set(ctx sdk.Context, key string, value valueType) 
```

and go through the following steps:

1. Retrieve the appropriate store from the `ctx` using the `storeKey`. This is done through the `KVStore(storeKey sdk.StoreKey` method of the `ctx`.
2. Marhsall `value` to `[]byte` using the codec `cdc`. 
3. Set the encoded value in the store at location `key` using the `Set(key []byte, value []byte)` method of the store. 

For more, see an example of `keeper`'s [methods implementation from the nameservice tutorial](https://github.com/cosmos/sdk-application-tutorial/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/x/nameservice/internal/keeper/keeper.go). 

## Next {hide}

Learn about [invariants](./invariants.md) {hide}
