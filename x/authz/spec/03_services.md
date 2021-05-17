<!--
order: 3
-->

# Tx Msg Service

In this section we describe the Protobuf Msg service for the authz module.

## Grant

An authorization-grant is created using the `MsgGrant` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L32-L37

This message is expected to fail if:

- both granter & grantee have same address.
- provided `Expiration` time less than current unix timestamp.
- provided `Authorization` is not implemented.
-
- Authorization Method doesn't exist (there is no defined handler in the app router to handle that Msg types)

If there is already a grant for the `(granter, grantee, Authorization)` triple, then the new grant will overwrite the previous one. To update or extend an existing grant, a new grant with the same `(granter, grantee, Authorization)` triple should be created.

## Revoke

An grant can be removed with `MsgRevoke` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L60-L64

This message is expected to fail if:

- both granter & grantee have same address.
- provided `MsgTypeUrl` is empty.

## Exec

When a grantee wants to execute transaction on behalf of a granter, it must send `MsgExec`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L47-L53

This message is expected to fail if:

- authorization not implemented for the provided msg.
- grantee don't have permission to run transaction.
- if granted authorization is expired.
