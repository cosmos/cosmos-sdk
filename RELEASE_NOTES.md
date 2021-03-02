# Cosmos SDK v0.41.4 "Stargate" Release Notes

This release includes the addition of the multisign-batch command, minor bug fixes, and performance improvements.

See the [Cosmos SDK v0.41.4 milestone](https://github.com/cosmos/cosmos-sdk/milestone/40?closed=1) on our issue tracker for details.

## multisign-batch command

Multisign-batch command was added and it allows generating multiple musltisig transactions by merging batches of signatures.

## Query tx with multisig addresses

Now the rest endpoint allows to query transactions with multisig addresses.

## Improvements

Major performance improvements in store and balance which will speed up genesis verification and initialization.

Tendermint was upgraded to v0.34.8.

Genesis now allows 0 coin account balances.

## Bugfixes

Minor bugfixes were included regarding missing errors and fields on some responses.

