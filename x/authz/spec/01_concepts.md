<!--
order: 0
-->

# Concepts

## Authorization
Any concrete type of authorization defined in the `x/authz` module must fulfill `Authorization` interface outlined below. Authorizations determine exactly what privileges are granted. They are extensible and can be defined for any Msg service method even outside of the module where the Msg method is defined. Authorizations use the new `ServiceMsg` type from [ADR 031](docs/architecture/adr-31-msg-service.md).


+++ https://github.com/cosmos/cosmos-sdk/blob/master/x/authz/types/authorizations.go#L15-L24


## Built-in Authorizations

### SendAuthorization



### GenericAuthorization
