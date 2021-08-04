# Cosmos SDK v0.43.0-RC3 Release Notes

This is a fourth Release Candidate for v0.43. If no issue will be reported this will be the final release.

This release contains two major bug fixes:

- [\#9841](https://github.com/cosmos/cosmos-sdk/pull/9841) Fix the bug where it was impossible to create IBC channels after an on-chain upgrade to 0.43. Please note that this is an **API-Breaking change**: `x/capability`'s `InitializeAndSeal()` has been removed and replaced with `Seal()`.
- [\#9793](https://github.com/cosmos/cosmos-sdk/pull/9793) Fixed ECDSA/secp256r1 transaction malleability.

One **client-breaking change** has also been introduced to fix an emitted event:

- [\#9785](https://github.com/cosmos/cosmos-sdk/issues/9785) Fix missing coin denomination in logs when emitting the `create_validator` event.

It also includes a couple of minor fixes and improvements:

- [\#9750](https://github.com/cosmos/cosmos-sdk/pull/9750) Emit events for tx signature and sequence, so clients can now query txs by signature (`tx.signature='<base64_sig>'`) or by address and sequence combo (`tx.acc_seq='<addr>/<seq>'`).
- Multiple CLI UX improvements.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.43.x/CHANGELOG.md) for more details.
