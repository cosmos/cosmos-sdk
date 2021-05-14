<!--
order: 3
-->

# Messages

In this section we describe the Msg Protobuf service for the authz module.

## Grant

An authorization-grant is created using the `MsgGrant` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/authz/v1beta1/tx.proto#L33-L37

This message is expected to fail if:

- both granter & grantee have same address.
- provided `Expiration` time less than current unix timestamp.
- provided `Authorization` is not implemented.
- there is already a grant for `(granter, grantee, Authorization)` triple.
- Authorization Method doesn't exist (there is no defined handler in the app router to handle that Msg types)

## Revoke

An allowed authorization can be removed with `MsgRevoke` message.

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
