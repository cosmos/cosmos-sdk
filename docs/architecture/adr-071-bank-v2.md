# ADR-71 Bank V2

## Status

DRAFT

## Changelog

* 2024-05-08: Initial Draft (@samricotta, @julienrbrt)

## Abstract

The primary objective of refactoring the bank module is to simplify and enhance the functionality of the Cosmos SDK. Over time the bank module has been burdened with numerous responsibilities including transaction handling, account restrictions, delegation counting, and the minting and burning of coins. 

In addition to the above, the bank module is currently too rigid and handles too many tasks, so this proposal aims to streamline the module by focusing on core functions `Send`, `Mint`, and `Burn`.

Currently, the module is split accross different keepers with scattered and duplicates functionalities (with 4 send functions for instance).

Additionally, the integration of the token factory into the bank module allows for standardization, and better integration within the core modules.

This rewrite will reduce complexity and enhance the efficiency and UX of the bank module.

## Context

The current implementation of the bank module is characterised by its handling of a broad array of functions, leading to significant complexity in using and extending the bank module. 

These issues have underscored the need for a refactoring strategy that simplifies the module’s architecture and focuses on its most essential operations.

Additionally, there is an overlap in functionality with a Token Factory module, which could be integrated to streamline oper.

## Decision

**Permission Tightening**: Access to the module can be restricted to selected denominations only, ensuring that it operates within designated boundaries and does not exceed its intended scope. Currently, the permissions allow all denoms, so this should be changed. Send restrictions functionality will be maintained.

**Simplification of Logic**: The bank module will focus on core functionalities `Send`, `Mint`, and `Burn`. This refinement aims to streamline the architecture, enhancing both maintainability and performance.

**Integration of Token Factory**: The Token Factory will be merged into the bank module. This consolidation of related functionalities aims to reduce redundancy and enhance coherence within the system. Migrations functions will be provided for migrating from Osmosis' Token Factory module to bank/v2.

**Legacy Support**: A legacy wrapper will be implemented to ensure compatibility with about 90% of existing functions. This measure will facilitate a smooth transition while keeping older systems functional.

**Callback Functions**: We propose to to integrate callbacks which are handled by an`x/account` and will handle the callback message type. This allows for  for customisable behaviour during minting, burning and transferring operations without complicating statemanagement.It also helps overwrite the default accounting mechanism by allowing to overwrite the balance method (this is particularly useful for rebasing tokens).

Each denom will be an `x/account` and will implement `OnTransfer`, `OnBurn`, `OnMint` and `OnBalanceOf` hooks. The bank module will check if the denom is an x/account and implement the respective hook. Depending on the hook, th bank module will send a message to the denom account. From there the denom account will return the new balances of the accounts.

We would have the following messages `MsgMint`, `MsgBurn`, `MsgOnTransfer`, `QueryBalance` that would be sent to the denom account. 

How the `x/account` denom implementation would look is as per below:

```go
func (a Account) RegisterExecuteHandlers(r accountsd.Router) {
accountsd.RegisterHandler(r, a.OnTransfer)
accountsd.RegisterHAndler(r. a.Burn)
}

func (a Account) OnTransfer(ctx context.Context, msg MsgOnTransfer) (MsgOnTransferResponse, error) {
    ... check if account KYCd...
}
```

For example if alice sends `MsgSend{to: bob, amount: 100ATOM}` to bank, bank checks if `ATOM` is an `x/account` implementing the `OnTransfer` hook.

If `ATOM` is an `x/account` implementing the `OnTransfer` hook then bank sends a `MsgTransfer` to the `ATOM` denom account. The `ATOM` denom account returns the new balances of alice and bob.

If `ATOM` is not an `x/account` implementing the `OnTransfer` hook then bank will simply substract amount from alice and add it to bob.

[flow diagram of the above process](flow.png)


## Migration Plans

Bank is a widely used module, so getting a v2 needs to be thought thoroughly. In order to not force all dependencies to immediately migrate to bank/v2, the same _upgrading_ path will be taken as for the `gov` module.

This means `cosmossdk.io/bank` will stay one module and there won't be a new `cosmossdk.io/bank/v2` go module. Instead the bank protos will be versioned from `v1beta1` (current bank) to `v2`.

Bank `v1beta1` endpoints will use the new bank v2 implementation for maximum backward compatibility.

The bank `v1beta1` keepers will be deprecated and potentially eventually removed, but its proto and messages definitions will remain.

Additionally, as bank plans to integrate token factory, migrations functions will be provided to migrate from Osmosis token factory implementation (most widely used implementation) to the new bank/v2 token factory.

## Consequences

### Positive

* Simplified interaction with bank APIs
* Backward comptible changes (no contracts or apis broken)
* Optional migration (note: bank `v1beta1` won't get any new feature after bank `v2` release)

### Neutral

* Asset implementation not available cross-chain (IBC-ed custom asset should possibly fallback to the default implementation)
* Many assets may slow down bank balances requests

### Negative

* Temporarily duplicate functionalities as bank `v1beta1` are `v2` are living alongside
* Difficultity to ever completely remove bank `v1beta1`

### References

* Current bank module implementation: https://github.com/cosmos/cosmos-sdk/blob/v0.50.6/x/bank/keeper/keeper.go#L22-L53
* Osmosis token factory: https://github.com/osmosis-labs/osmosis/tree/v25.0.0/x/tokenfactory/keeper
