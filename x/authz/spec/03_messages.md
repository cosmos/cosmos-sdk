<!--
order: 3
-->

# Messages

In this section we describe the processing of messages for the authz module.

## MsgGrantAuthorization

An authorization-grant is created using the `MsgGrantAuthorization` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/authz/v1beta1/tx.proto#L27-L35

This message is expected to fail if:
    
- both granter & grantee have same address.
- provided `Expiration` time less than current unix timestamp.
- provided `Authorization` is not implemented.

## MsgRevokeAuthorization

An allowed authorization can be removed with `MsgRevokeAuthorization` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/authz/v1beta1/tx.proto#L53-L59

This message is expected to fail if:

- both granter & grantee have same address.
- provided `MethodName` is empty.

## MsgExecAuthorizedRequest

When a grantee wants to execute transaction on behalf of a granter, it must send MsgExecAuthorizedRequest.  

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/authz/v1beta1/tx.proto#L42-L48

This message is expected to fail if:

- authorization not implemented for the provided msg.
- grantee don't have permission to run transaction.
- if granted authorization is expired.