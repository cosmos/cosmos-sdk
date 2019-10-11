# Invariants

## Pre-requisite Reading

- [Keepers](./keeper.md)

## Synopsis

An invariant is a property of the application that should always be true. An `Invariant` is a function that checks for a particular invariant. These functions are useful to detect bugs early on and act upon them to limit the potential consequences (e.g. by halting the chain). They are also useful in the development process of the application to detect bugs via simulations. 

- [Implementing `Invariant`s](#implementing-invariants)
- [Invariant Registry](#invariant-registry)

## Implementing `Invariant`s

An `Invariant` is a function that checks for a particular invariant within a module. Module `Invariant`s must follow the [`Invariant`s type](https://github.com/cosmos/cosmos-sdk/blob/master/types/invariant.go#L9):

```go
type Invariant func(ctx Context) (string, bool)
```

where the `string` return value is the invariant message, which can be used when printing logs, and the `bool` return value is the actual result of the invariant check. 

In practice, each module implements `Invariant`s in a `internal/keeper/invariants.go` file within the module's folder. The standard is to implement one `Invariant` function per logical grouping of invariants with the following model:

```go
// Example for an Invariant that checks balance-related invariants

func BalanceInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
        // Implement checks for balance-related invariants
    }
}
```

Additionally, module developers should generally implement an `AllInvariants` function that runs all the `Invariant`s functions of the module:

```go
// AllInvariants runs all invariants of the module.
// In this example, the module implements two Invariants: BalanceInvariants and DepositsInvariants

func AllInvariants(k Keeper) sdk.Invariant {

	return func(ctx sdk.Context) (string, bool) {
		res, stop := BalanceInvariants(k)(ctx)
		if stop {
			return res, stop
		}

		return DepositsInvariant(k)(ctx)
	}
}
```

Finally, module developers need to implement the `RegisterInvariants` method as part of the [`AppModule` interface](./module-manager.md#appmodule). Indeed, the `RegisterInvariants` method of the module, implemented in the `module.go` file, typically only defers the call to a `RegisterInvariants` method implemented in `internal/keeper/invariants.go`. The `RegisterInvariants` method registers a route for each `Invariant` function in the [`InvariantRegistry`](#invariant-registry):


```go
// RegisterInvariants registers all staking invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-accounts",
		BalanceInvariants(k))
	ir.RegisterRoute(types.ModuleName, "nonnegative-power",
		DepositsInvariant(k))
}
```

For more, see an example of [`Invariant`s implementation from the `staking` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/keeper/invariants.go). 

## Invariant Registry

The `InvariantRegistry` is a registry where the `Invariant`s of all the modules of an application are registered. There is only one `InvariantRegistry` per **application**, meaning module developers need not implement their own `InvariantRegistry` when building a module. All module developers need to do is to register their modules' invariants in the `InvariantRegistry`, as explained in the section above. 

Typically, the `InvariantRegistry` is implemented as a specific module (the most used implementation is that of the [`crisis` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/crisis/)). This module must implement an object that follow the [`sdk.InvariantRegistry` interface](https://github.com/cosmos/cosmos-sdk/blob/master/types/invariant.go#L14-L17). 

```go
type InvariantRegistry interface {
	RegisterRoute(moduleName, route string, invar Invariant)
}
```

Typically, this interface is implemented in the module's `keeper`. You can see an example implementation of an `InvariantRegistry` from the `crisis` module [here](https://github.com/cosmos/cosmos-sdk/blob/master/x/crisis/internal/keeper/keeper.go).

`Invariant`s can be checked manually via [`message`s](./messages-and-queries), but most often they are checked automatically at the end of each block (see an example [here](https://github.com/cosmos/cosmos-sdk/blob/master/x/crisis/abci.go)). In both cases, if one of the `Invariant`s returns false, the `InvariantRegistry` can trigger special logic (e.g. have the application panic and print the `Invariant`s message in the log).

## Next

Learn about the recommended [structure for modules](./structure.md). 
