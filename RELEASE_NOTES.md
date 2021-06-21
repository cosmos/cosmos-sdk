# Cosmos SDK v0.42.6 "Stargate" Release Notes

This release includes various minor bugfixes and improvments, including:

- x/bank's InitGenesis optimization, which should significantly decrease genesis initialization time,
- bump Tendermint to v0.34.11 to fix state sync issues,
- add `cosmos_sdk_version` to `node_info` to be able to query the SDK version used by a node,
- IBC bugfixes and improvements (see below for more info),
- new fields on `sdk.Context` (see below for more info).

See the [Cosmos SDK v0.42.6 milestone](https://github.com/cosmos/cosmos-sdk/milestone/45?closed=1) on our issue tracker for the exhaustive list of all changes.

### IBC Bugfixes and Improvements

The `[appd] query ibc client header` is fixed and allows querying by height for the header and node-state command. This allows easier venerability of which IBC tokens belong to which chains. IBC's ExportGenesis now exports all fields, including previously missing `NextClientSequence`, `NextConnectionSequence` and `NextChannelSequence`. A new subcommand `[appd] query ibc-transfer escrow-address` has been added to get the escrow address for a channel; it can be used to then query balance of escrowed tokens.

### New Fields on `sdk.Context`

Two fields have been added on `sdk.Context`:

- `ctx.HeaderHash` adds the current block header hash obtained during abci.RequestBeginBlock to the Context,
- `ctx.GasMeter().RefundGas(<amount>, <description>)` adds support for refunding gas directly to the gas meter.
