# ADR-71 Bank V2

## Status

DRAFT

## Changelog

* 2024-05-08: Initial Draft (@samricotta, @julienrbrt)

## Abstract

The primary objective of refactoring the bank module is to simplify and enhance the functionality of the Cosmos SDK. Over time the bank module has been burdened with numerous responsibilities including transaction handling, account restrictions, delegation counting, and the minting and burning of coins. 

In addition to the above, the bank module is currently too rigid and handles too many tasks, so this proposal aims to streamline the module by focusing on core functions `Send`, `Mint`, and `Burn`.

Currently, the module is split across different keepers with scattered and duplicates functionalities (with 4 send functions for instance).

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

**Denom Implementation**: A asset interface will be added to standardise interactions such as transfers, balance inquiries, minting, and burning across different tokens. This will allow the bank module to support arbitrary asset types, enabling developers to implement custom, ERC20-like denominations.

For example, currently if a team would like to extend the transfer method the changes would apply universally, affecting all denom’s. With the proposed Asset Interface, it allows teams to customise or extend the transfer method specifically for their own tokens without impacting others.

These improvements are expected to enhance the flexibility of the bank module, allowing for the creation of custom tokens similar to ERC20 standards and assets backed by CosmWasm (CW) contracts. The integration efforts will also aim to unify CW20 with bank coins across the Cosmos chains.

Example of denom interface:

```go
type AssetInterface interface {
    Transfer(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, amount sdk.Coin) error
    Mint(ctx sdk.Context, to sdk.AccAddress, amount sdk.Coin) error
    Burn(ctx sdk.Context, from sdk.AccAddress, amount sdk.Coin) error
    QueryBalance(ctx sdk.Context, account sdk.AccAddress) (sdk.Coin, error)
}
```

Overview of flow:

1. Alice initiates a transfer by entering Bob's address and the amount (100 ATOM)
2. The Bank module verifies that the ATOM token implements the `AssetInterface` by querying the `ATOM_Denom_Account`, which is an `x/account` denom account.
3. The Bank module executes the transfer by subtracting 100 ATOM from Alice’s balance and adding 100 ATOM to Bob’s balance.
4. The Bank module calls the Transfer method on the `ATOM_Denom_Account`. The Transfer method, defined in the `AssetInterface`, handles the logic to subtract 100 ATOM from Alice’s balance and add 100 ATOM to Bob’s balance.
5. The Bank module updates the chain and returns the new balances.
6. Both Alice and Bob successfully receive the updated balances.










## Migration Plans

Bank is a widely used module, so getting a v2 needs to be thought thoroughly. In order to not force all dependencies to immediately migrate to bank/v2, the same _upgrading_ path will be taken as for the `gov` module.

This means `cosmossdk.io/bank` will stay one module and there won't be a new `cosmossdk.io/bank/v2` go module. Instead the bank protos will be versioned from `v1beta1` (current bank) to `v2`.

Bank `v1beta1` endpoints will use the new bank v2 implementation for maximum backward compatibility.

The bank `v1beta1` keepers will be deprecated and potentially eventually removed, but its proto and messages definitions will remain.

Additionally, as bank plans to integrate token factory, migrations functions will be provided to migrate from Osmosis token factory implementation (most widely used implementation) to the new bank/v2 token factory.

## Consequences

### Positive

* Simplified interaction with bank APIs
* Backward compatible changes (no contracts or apis broken)
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
