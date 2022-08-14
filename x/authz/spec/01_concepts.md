<!--
order: 1
-->

# Concepts

## Authorization and Grant

`x/authz` module defines interfaces and messages grant authorizations to perform actions
on behalf of one account to other accounts. The design is defined in the [ADR 030](../../../architecture/adr-030-authz-module.md).

Grant is an allowance to execute a Msg by the grantee on behalf of the granter.
Authorization is an interface which must be implemented by a concrete authorization logic to validate and execute grants. They are extensible and can be defined for any Msg service method even outside of the module where the Msg method is defined. See the `SendAuthorization` example in the next section for more details.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/authz/authorizations.go#L11-L25

## Built-in Authorizations

Cosmos-SDK `x/authz` module comes with following authorization types

### SendAuthorization

`SendAuthorization` implements the `Authorization` interface for the `cosmos.bank.v1beta1.MsgSend` Msg. It takes a `SpendLimit` that specifies the maximum amount of tokens the grantee can spend, which is updated as the tokens are spent.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/bank/v1beta1/authz.proto#L10-L19

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/bank/types/send_authorization.go#L25-L40

- `spent_limit` keeps track of how many coins are left in the authorization.

### GenericAuthorization

`GenericAuthorization` implements the `Authorization` interface, that gives unrestricted permission to execute the provided Msg on behalf of granter's account.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/authz.proto#L14-L19

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/authz/generic_authorization.go#L18-L31

- `msg` stores Msg type URL.

## Gas

In order to prevent DoS attacks, granting `StakeAuthorizaiton`s with `x/authz` incur gas. `StakeAuthorizaiton` allows you to authorize another account to delegate, undelegate, or redelegate to validators. The authorizer can define a list of validators they will allow and/or deny delegations to. The SDK will iterate over these lists and charge 10 gas for each validator in both of the lists.
