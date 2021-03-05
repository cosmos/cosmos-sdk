# Cosmos SDK v0.42.0 "Stargate" Release Notes

This release includes an important security fix for all non "Cosmos Hub" chains (e.g. any chain that does not use the default `cosmos` bech32 prefix), and a few performance improvements.

See the [Cosmos SDK v0.42.0 milestone](https://github.com/cosmos/cosmos-sdk/milestone/42?closed=1) on our issue tracker for further details.

# Security fix: validator address conversion in evidence handling

Due to incorrect handling of validators' consensus addresses, Cosmos SDK applications that were not
using the default `cosmos1` bech32 default one could not be able to jail validators that commit
misbehaviours such as double signing.

Although the issue **does not affect the Cosmos Hub**, it still potentially renders the `v0.41` and `v0.40`
release series unsafe for most chains.

# Full header is emitted on IBC UpdateClient message event

The event emitted by the IBC UpdateClient message now contains the full header.
This is meant to make misbehaviour handling, which requires tracking these headers, easier.
