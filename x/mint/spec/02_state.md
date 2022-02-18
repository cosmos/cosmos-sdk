<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current inflation information.

* Minter: `0x00 -> ProtocolBuffer(minter)`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc7/proto/cosmos/mint/v1beta1/mint.proto#L8-L19

## Params

Minting params are held in the global params store.

* Params: `mint/params -> legacy_amino(params)`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc7/proto/cosmos/mint/v1beta1/mint.proto#L21-L53
