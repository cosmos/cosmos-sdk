<!--
order: 2
-->

# State

## AuthorizationGrant

Grants are identified by combining granter address (the address bytes of the granter), grantee address (the address bytes of the grantee) and Msg type (its TypeURL). Hence we only allow one grant for the (granter, grantee, msg type) triple.

- AuthorizationGrant: `0x01 | granter_address_len (1 byte) | granter_address_bytes | grantee_address_len (1 byte) | grantee_address_bytes |  msgType_bytes-> ProtocolBuffer(AuthorizationGrant)`


+++ https://github.com/cosmos/cosmos-sdk/blob/c95de9c4177442dee4c69d96917efc955b5d19d9/proto/cosmos/authz/v1beta1/authz.proto#L32-L37
