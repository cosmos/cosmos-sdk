# Upgrade Reference

This document provides a quick reference for the upgrades from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the **App Wiring Changes** section for more information on application wiring updates.

🚨Upgrading to v0.54.x will require a **coordinated** chain upgrade.🚨

### TLDR

**The only major feature in Cosmos SDK v0.54.x is the upgrade from CometBFT v0.x.x to CometBFT v2.**

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).

#### Deprecation of `TimeoutCommit`

CometBFT v2 has deprecated the use of `TimeoutCommit` for a new field, `NextBlockDelay`, that is part of the
`FinalizeBlockResponse` ABCI message that is returned to CometBFT via the SDK baseapp.  More information from
the CometBFT repo can be found [here](https://github.com/cometbft/cometbft/blob/88ef3d267de491db98a654be0af6d791e8724ed0/spec/abci/abci%2B%2B_methods.md?plain=1#L689).

For SDK application developers and node runners, this means that the `timeout_commit` value in the `config.toml` file
is still used if `NextBlockDelay` is 0 (its default value).  This means that when upgrading to Cosmos SDK v0.54.x, if 
the existing `timout_commit` values that validators have been using will be maintained and have the same behavior.

For setting the field in your application, there is a new `baseapp` option, `SetNextBlockDelay` which can be passed to your application upon
initialization in `app.go`.  Setting this value to any non-zero value will override anything that is set in validators' `config.toml`.
