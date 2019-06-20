# Concepts

## Supply

The total `Supply` of the network is equal to the sum of all `Account`s stored. The total supply is updated every time a coin is minted (as part of the inflation mechanism)  or burned (due to slashing or if a governance proposal is vetoed).

## Module Accounts

The supply module introduces a new type of `Account`  used by modules to control the flow of coins in and out of a module. This design replaces pools of coins that were stored on each of the modules, such as the `FeeCollectorKeeper` and the  staking `Pool`. The resoning of having this new account type is to calculate the total supply without having to access each of the module's pools, thus reducing the dependencies. 

A `ModuleAccount` is extends the `BaseAccount` functionallity by providing a root `Name` that is used to create the `ModuleAccount` address based on the hash of the `Name`.

```go
moduleAddress = AccAddress(AddressHash([]byte(name)))
```

The `ModuleAccount` interface is defined as follows:

```go
type ModuleAccount interface {
  auth.Account   // same methods as the Account interface
  Name() string  // module's or pool's name; used to obtain the address  
}
```

### Types

A `ModuleAccount` can be one of three types:

- `ModuleHolderAccount`: is allowed to `SendCoinsFromModuleToAccount` and `SendCoinsFromModuleToModule`.
- `ModuleBurnerAccount`: allows for a module to call the `Burn` function to burn a specific amount of `Coins`
- `ModuleMinterAccount`: allows for a module to call the `Mint` function to mint a specific amount of `Coins`
