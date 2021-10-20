<!--
order: 3
-->

# Messages

In this section we describe the processing of messages for the authz module.

## MsgGrant

An authorization grant is created using the `MsgGrant` message.
If there is already a grant for the `(granter, grantee, Authorization)` triple, then the new grant overwrites the previous one. To update or extend an existing grant, a new grant with the same `(granter, grantee, Authorization)` triple should be created.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L32-L37

The message handling should fail if:

- both granter and grantee have the same address.
- provided `Expiration` time is less than current unix timestamp.
- provided `Grant.Authorization` is not implemented.
- `Authorization.MsgTypeURL()` is not defined in the router (there is no defined handler in the app router to handle that Msg types).

## MsgRevoke

A grant can be removed with the `MsgRevoke` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L60-L64

The message handling should fail if:

- both granter and grantee have the same address.
- provided `MsgTypeUrl` is empty.

NOTE: The `MsgExec` message removes a grant if the grant has expired.

## MsgExec

When a grantee wants to execute a transaction on behalf of a granter, they must send `MsgExec`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L47-L53

The message handling should fail if:

- provided `Authorization` is not implemented.
- grantee doesn't have permission to run the transaction.
- if granted authorization is expired.
