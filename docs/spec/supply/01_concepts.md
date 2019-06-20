# Concepts

## Supply

The `supply` module introduces a passive tracker for the `Supply` of coins in the chain, which allows to check for invariants of `Coins` with respect to the sum of `Coins` hold the stored `Accounts`.

### Total Supply

The total `Supply` of the network is equal to the sum of all coins from the `Account`s. The total supply is updated every time a `Coin` is minted (as part of the inflation mechanism) or burned (due to slashing or if a governance proposal is vetoed).

## Module Accounts

To keep track of the `Supply`, this module introduces a new type of `Account` used by other modules to control the flow of coins that come into and out of each of their modules. This design replaces the pools of coins that were stored on each of the modules, such as the `FeeCollectorKeeper` and the  staking `Pool`. The resoning of having this new `Account` type is to calculate the total supply without having to access each of the modules' pools of coins (previously stored in the `Store`), thus reducing the dependencies.

A `ModuleAccount` extends the `BaseAccount` functionallity by providing a root `Name` that is used to create the `ModuleAccount` address based on the hash of the `Name`:

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

The supply `Keeper` also introduces new wrapper functions for the `AccountKeeper` and the bank `Keeper` to be able to:

- Get and set `ModuleAccount`s by providing the `Name`.
- Send coins from and to other `ModuleAccount`s or standard `Account`s (`BaseAccount` or `VestingAccount`) by passing only the `Name`.
- `Mint` and `Burn` coins. These methods are restricted to the different `ModuleAccount` [types](./01_concepts#types).

### Types

A `ModuleAccount` can be one of three types:

- `ModuleHolderAccount`: is allowed to only hold and transfer `Coins`.
- `ModuleBurnerAccount`: allows for a module to call the `Burn` function to burn a specific amount of `Coins`.
- `ModuleMinterAccount`: allows for a module to call the `Mint` function to mint a specific amount of `Coins`.
