<!--
order: 2
-->

# State

## AuthorizationGrant

Authorizations are identified by combining granter address (the address bytes of the granter), grantee address (the address bytes of the grantee) and ServiceMsg type.

- AuthorizationGrant: `0x01 | granter_address_bytes | grantee_address_bytes |  msgType_bytes-> ProtocolBuffer(AuthorizationGrant)`


+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/authz/v1beta1/authz.proto#L32-L37