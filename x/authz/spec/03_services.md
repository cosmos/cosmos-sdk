<!--
order: 3
-->

# Msg Service

In this section we describe the Protobuf Msg service for the authz module.

## Grant

The `Grant` method creates a new grant using `MsgGrant` message.
If there is already a grant for the `(granter, grantee, Authorization)` triple, then the new grant will overwrite the previous one. To update or extend an existing grant, a new grant with the same `(granter, grantee, Authorization)` triple should be created.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L32-L37

The method is expected to fail if:

- both granter and grantee have the same address.
- provided `Expiration` time is less than current unix timestamp.
- provided `Grant.Authorization` is not implemented.
- `Authorization.MsgTypeURL()` is not defined in the router (there is no defined handler in the app router to handle that Msg types).


## Revoke

`Revoke` method removes a grant using `MsgRevoke` message:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L60-L64

The method is expected to fail if:

- both granter and grantee have the same address.
- provided `MsgTypeUrl` is empty.

NOTE: `Exec` method removes a grant if it is exhausted.

## Exec

When a grantee wants to execute transaction on behalf of a granter, it must send `MsgExec`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L47-L53

The method is expected to fail if:

- provided `Authorization` is not implemented.
- grantee doesn't have permission to run the transaction.
- if granted authorization is expired.
