<!--
order: 2
-->

# State

Currently the x/authz module only stores valid submitted Authorization in state. The Authorization state is also stored and exported in the x/authz module's GenesisState.

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/authz/v1beta1/genesis.proto#L12-L24

All Authorization is retrieved and stored via a prefix KVStore using prefix `0x01<granterAddress_Bytes><granteeAddress_Bytes><msgType_Bytes>` (AuthorizationStoreKey).

## AuthorizationGrant

`AuthorizationGrant` is a space for holding authorization and expiration time.

+++ https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/authz/v1beta1/authz.proto#L32-L37