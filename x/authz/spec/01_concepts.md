<!--
order: 1
-->

# Concepts

## Authorization and Grant

The `x/authz` module defines interfaces and messages grant authorizations to perform actions
on behalf of one account to other accounts. The design is defined in the [ADR 030](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-030-authz-module.md).

A *grant* is an allowance to execute a Msg by the grantee on behalf of the granter.
Authorization is an interface that must be implemented by a concrete authorization logic to validate and execute grants. Authorizations are extensible and can be defined for any Msg service method even outside of the module where the Msg method is defined. See the `SendAuthorization` example in the next section for more details.

**Note:** The authz module is different from the [auth (authentication)](../modules/auth/) module that is responsible for specifying the base transaction and account types.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/authz/authorizations.go#L11-L25

## Built-in Authorizations

The Cosmos SDK `x/authz` module comes with following authorization types:

### GenericAuthorization

`GenericAuthorization` implements the `Authorization` interface that gives unrestricted permission to execute the provided Msg on behalf of granter's account.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/authz/v1beta1/authz.proto#L13-L20

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/authz/generic_authorization.go#L16-L29

* `msg` stores Msg type URL.

### SendAuthorization

`SendAuthorization` implements the `Authorization` interface for the `cosmos.bank.v1beta1.MsgSend` Msg. It takes a (positive) `SpendLimit` that specifies the maximum amount of tokens the grantee can spend. The `SpendLimit` is updated as the tokens are spent.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/bank/v1beta1/authz.proto#L10-L19

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/bank/types/send_authorization.go#L23-L38

* `spend_limit` keeps track of how many coins are left in the authorization.

### StakeAuthorization

`StakeAuthorization` implements the `Authorization` interface for messages in the [staking module](https://docs.cosmos.network/v0.44/modules/staking/). It takes an `AuthorizationType` to specify whether you want to authorise delegating, undelegating or redelegating (i.e. these have to be authorised seperately). It also takes a required `MaxTokens` that keeps track of a limit to the amount of tokens that can be delegated/undelegated/redelegated. If left empty, the amount is unlimited. Additionally, this Msg takes an `AllowList` or a `DenyList`, which allows you to select which validators you allow or deny grantees to stake with.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/staking/v1beta1/authz.proto#L10-L33

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/staking/types/authz.go#L15-L35

## Gas

In order to prevent DoS attacks, granting `StakeAuthorization`s with `x/authz` incurs gas. `StakeAuthorization` allows you to authorize another account to delegate, undelegate, or redelegate to validators. The authorizer can define a list of validators they allow or deny delegations to. The Cosmos SDK iterates over these lists and charge 10 gas for each validator in both of the lists.

Since the state maintaining a list for granter, grantee pair with same expiration, we are iterating over the list to remove the grant (incase of any revoke of paritcular `msgType`) from the list and we are charging 20 gas per iteration.
