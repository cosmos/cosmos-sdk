<!--
order: 1
-->

# Concepts

## Authorization
Any concrete type of authorization defined in the `x/authz` module must fulfill the `Authorization` interface outlined below. Authorizations determine exactly what privileges are granted. They are extensible and can be defined for any Msg service method even outside of the module where the Msg method is defined. Authorizations use the new `ServiceMsg` type from [ADR 031](../../../architecture/adr-031-msg-service.md).


+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/x/authz/types/authorizations.go#L15-L24


## Built-in Authorizations

Cosmos-SDK `x/authz` module comes with following authorization types

### SendAuthorization

`SendAuthorization` implements `Authorization` interface for the `cosmos.bank.v1beta1.Msg/Send` ServiceMsg, that takes a `SpendLimit` and updates it down to zero.

+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/proto/cosmos/authz/v1beta1/authz.proto#L12-L19

+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/x/authz/types/send_authorization.go#L23-L45

- `spent_limit` keeps track of how many coins left in the authorization.


### GenericAuthorization

`GenericAuthorization` implements the `Authorization` interface, that gives unrestricted permission to execute the provided ServiceMsg on behalf of granter's account.

+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/proto/cosmos/authz/v1beta1/authz.proto#L21-L30

+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/x/authz/types/generic_authorization.go#L20-L28

- `method_name` holds ServiceMsg type.
