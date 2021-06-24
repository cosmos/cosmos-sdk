# Cosmos SDK v0.43.0 Release Notes

This release introduces several new important updates to the Cosmos SDK. The release notes below provide an overview of the larger high-level changes introduced in the v0.43 release series.

That being said, this release does contain many more minor and module-level changes besides those mentioned below. For a comprehsive list of all breaking changes and improvements since the v0.42 "Stargate" release series, please see the
[changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-rc0/CHANGELOG.md).

## Two new modules: `x/authz` and `x/feegrant`

The v0.43 release focused on simplifying [keys and fee management](https://github.com/cosmos/cosmos-sdk/issues/7074) for SDK users, by introducing the two following modules:

- `x/feegrant` allows one account, the "granter" to grant another account, the "grantee" an allowance to spend the granter's account balance for fees within certain well-defined limits. It solves the problem of signing accounts needing to possess a sufficient balance in order to pay fees.
  - See [ADR-029](https://github.com/cosmos/cosmos-sdk/blob/8f21e26a6506c9ee81686bad6cf9be3f0e8e11d7/docs/architecture/adr-029-fee-grant-module.md) introducing the module.
  - Read the [documentation](https://docs.cosmos.network/master/modules/feegrant/).
  - Check out its [Protobuf definitions](https://github.com/cosmos/cosmos-sdk/tree/master/proto/cosmos/feegrant/v1beta1).
- `x/authz` provides functionality for granting arbitrary privileges from one account (the "granter") to another account (the "grantee"). These privileges, called [`Authorization`](https://github.com/cosmos/cosmos-sdk/blob/f2cea6a137ce19ad8987fa8a0cb99f4b37c4484d/x/authz/authorizations.go#L11)s in the code, can for example allow grantees to execute `Msg`s on behalf of the granter.
  - See [ADR-030](https://github.com/cosmos/cosmos-sdk/blob/8f21e26a6506c9ee81686bad6cf9be3f0e8e11d7/docs/architecture/adr-030-authz-module.md) introducing the module.
  - Read the [documentation](https://docs.cosmos.network/master/modules/authz/).
  - Check out its [Protobuf definitions](https://github.com/cosmos/cosmos-sdk/tree/master/proto/cosmos/authz/v1beta1).

## ADR-028 Address derivation from public keys

## In-place store migrations
