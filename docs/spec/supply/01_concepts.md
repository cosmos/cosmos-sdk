# Concepts

## Supply

The `supply` module introduces a passive tracker for the supply of coins in the chain, which allows to check for invariants of `Coins` with respect to the sum of `Coins` hold the stored `Accounts`.

### Total Supply

The total `Supply` of the network is equal to the sum of all coins from the account. The total supply is updated every time a `Coin` is minted (eg: as part of the inflation mechanism) or burned (eg: due to slashing or if a governance proposal is vetoed).

## Module Accounts

To keep track of the `Supply`, this module introduces a new type of `Account` used by other modules to control the flow of coins that come into and out of each of their modules. This design replaces the pools of coins that were stored on each of the modules, such as the `FeeCollectorKeeper` and the staking `Pool`. The reasoning of having this new `Account` type is to calculate the total supply without having to access each of the modules' pools of coins (previously stored in the `Store`), thus reducing the dependencies.

The `ModuleAccount` interface is defined as follows:

```go
type ModuleAccount interface {
  auth.Account            // same methods as the Account interface
  GetName() string        // name of the module; used to obtain the address
  GetPermission() string  // permission of module account (minter/burner/holder)
}
```

The supply `Keeper` also introduces new wrapper functions for the auth `Keeper` and the bank `Keeper` that are related to `ModuleAccount`s in order to be able to:

- Get and set `ModuleAccount`s by providing the `Name`.
- Send coins from and to other `ModuleAccount`s or standard `Account`s (`BaseAccount` or `VestingAccount`) by passing only the `Name`.
- `Mint` or `Burn` coins for a `ModuleAccount` (restricted to its permissions).

### Permissions

Each `ModuleAccount` has a different set of permissions that provide different object capabilities to perform certain actions. Permissions need to be registered upon the creation of the supply `Keeper` so that every time a `ModuleAccount` calls the allowed functions, the `Keeper` can lookup the permission to that specific account and perform or not the action.

The available permissions are:

- `Basic`: is allowed to only transfer its coins to other accounts.
- `Minter`: allows for a module to mint a specific amount of coins as well as perform the `Holder` permissioned actions.
- `Burner`: allows for a module to burn a specific amount of coins as well as perform the `Holder` permissioned actions.
