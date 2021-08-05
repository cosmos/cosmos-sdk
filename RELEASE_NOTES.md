# Cosmos SDK v0.43.0-RC3 Release Notes

This is a fourth Release Candidate for v0.43. If no issue will be reported this will be the final release.

This release contains two major bug fixes:

- This release includes an important `x/capabiliy` module bug fix which prohibited IBC to create new channels (issue [\#9800](https://github.com/cosmos/cosmos-sdk/issues/9800)).
  The fix introduces an API-breaking change by removing the `x/capability/keeper/Keeper.InitializeAndSeal` method and replaces it with `Seal()`. It also requires app developers to update their app module manager by adding x/capability module to `BeginBlocker`s before any other module's `BeginBlocker`.
- [\#9793](https://github.com/cosmos/cosmos-sdk/pull/9793) Fixed ECDSA/secp256r1 transaction malleability.

One **client-breaking change** has also been introduced to fix an emitted event:

- [\#9785](https://github.com/cosmos/cosmos-sdk/issues/9785) Fix missing coin denomination in logs when emitting the `create_validator` event.

It also includes a couple of minor fixes and improvements:

- [\#9750](https://github.com/cosmos/cosmos-sdk/pull/9750) Emit events for tx signature and sequence, so clients can now query txs by signature (`tx.signature='<base64_sig>'`) or by address and sequence combo (`tx.acc_seq='<addr>/<seq>'`).
- Multiple CLI UX improvements. Notably, we changed `<app> tx distribution withdraw-all-rewards` CLI by forcing the broadcast mode if a chunk size is greater than 0. This will ensure that the transactions do not fail even if the user uses invalid broadcast modes for this command (sync and async). This was requested by the community and we consider it as fixing the `withdraw-all-rewards` behavior.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.43.x/CHANGELOG.md) for more details.
