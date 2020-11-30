# Cosmos SDK v0.39.2 Release Notes

This release fixes various bugs and brings coin's denom validation to the *Cosmos SDK 0.39* release series.

See the [Cosmos SDK 0.39.2 milestone](https://github.com/cosmos/cosmos-sdk/milestone/30?closed=1) on our issue tracker for details.

## Allow ValidateDenom() to be customised per application

Applications can now customise `types.Coin` denomination validation by passing
their application-specific validation function to `types.SetCoinDenomRegex()`.

## Upgrade queries don't work after upgrade

New stores can now be registered during an on-chain upgrade. This is to
prevent blockchain state queries from stopping working after a successful upgrade.

## ApproxRoot() infinite looping

The `types.Dec.ApproxRoot()` function has now a maximum number 100 iterations as backup boundary
condition to prevent the client's code from entering an endless loop.

## Go 1.15

This is the first release of the Launchpad series that has been tested and built with **go 1.15**.

## Tendermint's updates

Tendermint has received a few updates in the last development cycle.

The pings frequency for remote private validators and the number of GetPubKey requests
have been reduced to prevent validators from [failing to sync when using remote signers](https://github.com/tendermint/tendermint/issues/5550).

A security vulnerability that affected the Go's `encoding/binary` package was reported.
Tendermint's `v0.33.8` release was published with the objective to aid users in using the correct version of Go.
Please refer to [this bug report](https://github.com/golang/go/issues/40618) for more information.

## Known issues

Keyrings using the `test` backend that were created with applications built with `Cosmos SDK v0.39.1`
and `go 1.15` may break with the following error after re-compiling with `Cosmos SDK v0.39.2`:

```
ERROR: aes.KeyUnwrap(): integrity check failed.
```

This is due to [the update](https://github.com/99designs/keyring/pull/75) that the `jose2go` dependency
has received that made it [fully compatible with go 1.15](https://github.com/dvsekhvalnov/jose2go/issues/26).
