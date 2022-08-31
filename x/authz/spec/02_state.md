<!--
order: 2
-->

# State

## Grant

Grants are identified by combining granter address (the address bytes of the granter), grantee address (the address bytes of the grantee) and Authorization type (its type URL). Hence we only allow one grant for the (granter, grantee, Authorization) triple.

* Grant: `0x01 | granter_address_len (1 byte) | granter_address_bytes | grantee_address_len (1 byte) | grantee_address_bytes |  msgType_bytes -> ProtocolBuffer(AuthorizationGrant)`

The grant object encapsulates an `Authorization` type and an expiration timestamp:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/proto/cosmos/authz/v1beta1/authz.proto#L22-L30

## GrantQueue

We are maintaining a queue for authz pruning. Whenever a grant is created, an item will be added to `GrantQueue` with a key of expiration, granter, grantee.

* GrantQueue: `0x02 | expiration_bytes | granter_address_len (1 byte) | granter_address_bytes | grantee_address_len (1 byte) | grantee_address_bytes -> ProtocalBuffer(GrantQueueItem)`

The `expiration_bytes` are the expiration date in UTC with the format `"2006-01-02T15:04:05.000000000"`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/authz/keeper/keys.go#L78-L93

The `GrantQueueItem` object contains the list of type urls between granter and grantee that expire at the time indicated in the key.
