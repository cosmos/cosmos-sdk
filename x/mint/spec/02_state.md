<!--
order: 2
-->

# State

## Minter

The minter is a space for holding current inflation information.

* Minter: `0x00 -> ProtocolBuffer(minter)`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/mint/v1beta1/mint.proto#L9-L23

## Params

The mint module stores it's params in state with the prefix of `0x01`,
it can be updated with governance or the address with authority.

* Params: `mint/params -> legacy_amino(params)`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/mint/v1beta1/mint.proto#L25-L57
