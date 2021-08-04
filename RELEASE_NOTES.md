# Cosmos SDK v0.42.9 "Stargate" Release Notes

This release includes an important `x/capabiliy` module bug fix for 0.42.7 and 0.42.8 which prohibits IBC to create new channels (issuse [\#9800](https://github.com/cosmos/cosmos-sdk/issues/9800)).
The fix changes the `x/capability/keeper/Keeper.InitializeAndSeal` method behavior and requires to update an app module manager by adding x/capability module to Begin Blockers.

We also fixed `<app> init --recovery` mode where the mnemonic was not handled correctly.

We also changed `<app> tx distribution withdraw-all-rewards` CLI by forcing the broadcast mode if a chunk size is greater than 0. This will ensure that the transactions do not fail even if the user uses invalid broadcast modes for this command (sync and async). This was requested by the community and we consider it as fixing the `withdraw-all-rewards` behavior.

See the [Cosmos SDK v0.42.9 milestone](https://github.com/cosmos/cosmos-sdk/milestone/53?closed=1) on our issue tracker for the exhaustive list of all changes.
