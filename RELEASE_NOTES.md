# Cosmos SDK v0.41.4 "Stargate" Release Notes

This release includes the addition of the multisign-batch command, minor bug fixes, and performance improvements.

See the [Cosmos SDK v0.41.4 milestone](https://github.com/cosmos/cosmos-sdk/milestone/40?closed=1) on our issue tracker for details.

## multisign-batch command

Multisign-batch command was added and it allows generating multiple multisig transactions by merging batches of signatures.

## Query tx with multisig addresses

Now the rest endpoint allows to query transactions with multisig addresses.

## Improvements

Major performance improvements in store and balance which will speed up genesis verification and initialization.

Genesis now allows 0 coin account balances. This means that genesis initialization will not fail if an address with no balance will be included.

## Bugfixes

The keys migration command (i.e. `keys migrate`) is now functional for offline, multisign, and ledger keys.

Minor bugfixes were included regarding missing errors and fields on some responses.

## Tendermint new release

Tendermint was upgraded to v0.34.8. This release of Tendermint introduces various changes that should make the logs much, much quieter. See [Tendermint's changelog](https://github.com/tendermint/tendermint/blob/v0.34.8/CHANGELOG.md#v0.34.8) for more information.
