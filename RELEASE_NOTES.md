# Cosmos SDK v0.39.2 Release Notes

This release fixes various bugs and brings coin's denom validation to the *Cosmos SDK 0.39* release series.

See the [Cosmos SDK 0.39.2 milestone](https://github.com/cosmos/cosmos-sdk/milestone/30?closed=1) on our issue tracker for details.

## Allow ValidateDenom() to be customised per application

Applications can now customise `types.Coin` denomination validation by
replacing `types.CoinDenomRegex` with their application-specific validation function.

## Upgrade queries don't work after upgrade

New stores can now be registered during an on-chain upgrade. This is to
prevent blockchain state queries from stopping working after a successful upgrade.

## ApproxRoot() infinite looping

The `types.Dec.ApproxRoot()` function has now a maximum number 100 iterations as backup boundary
condition to prevent the client's code from entering an endless loop.

## Go 1.15

This is the first release of the Launchpad series that has been tested and built with **go 1.15**.

## Tendermint 0.33.8

A security vulnerability that affected the Go's `encoding/binary` package was reported.
Tendermint's new `v0.33.8` is meant to aid users in using the correct version of Go.

Please refer to [this bug report](https://github.com/golang/go/issues/40618) for more information.
