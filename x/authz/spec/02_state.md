<!--
order: 2
-->

# State

## Grant

Grants are identified by combining granter address (the address bytes of the granter), grantee address (the address bytes of the grantee) and Authorization type (its type URL). Hence we only allow one grant for the (granter, grantee, Authorization) triple.

* Grant: `0x01 | granter_address_len (1 byte) | granter_address_bytes | grantee_address_len (1 byte) | grantee_address_bytes |  msgType_bytes-> ProtocolBuffer(AuthorizationGrant)`

The grant object encapsulates an `Authorization` type and an expiration timestamp:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/authz/v1beta1/authz.proto#L22-L30

## GrantQueue

We are maintaining a queue for authz pruning, whenever a grant created an item will be added to `GrantQueue` with a key of granter, grantee, expiration and value added as array of msg type urls.

* GrantQueue: `0x02 | granter_address_len (1 byte) | granter_address_bytes | grantee_address_len (1 byte) | grantee_address_bytes | expiration_bytes -> ProtocalBuffer([]string{msgTypeUrls})`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/authz/keeper/keys.go#L78-L93
