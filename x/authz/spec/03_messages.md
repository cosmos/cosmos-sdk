<!--
order: 3
-->

# Messages

In this section we describe the processing of messages for the authz module.

## Msg/Grant

An authorization-grant is created using the `MsgGrant` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/proto/cosmos/authz/v1beta1/tx.proto#L27-L35

This message is expected to fail if:

- both granter & grantee have same address.
- provided `Expiration` time less than current unix timestamp.
- provided `Authorization` is not implemented.
- Authorization Method doesn't exist (there is no defined handler in the app router to handle that Msg types)

## Msg/Revoke

An allowed authorization can be removed with `MsgRevoke` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/proto/cosmos/authz/v1beta1/tx.proto#L53-L59

This message is expected to fail if:

- both granter & grantee have same address.
- provided `MsgTypeUrl` is empty.

## Msg/Exec

When a grantee wants to execute transaction on behalf of a granter, it must send MsgExecRequest.

+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/proto/cosmos/authz/v1beta1/tx.proto#L42-L48

This message is expected to fail if:

- authorization not implemented for the provided msg.
- grantee don't have permission to run transaction.
- if granted authorization is expired.
