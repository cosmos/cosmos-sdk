# Cosmos SDK v0.42.7 "Stargate" Release Notes

This release includes various minor bugfixes and improvments, including:

- a x/capability initialization fix, which fixes the consensus error when using statesync,
- CLI improvements such as fixing the `{appd} keys parse` subcommand and a better user error when `--chain-id` is not passed in the `{appd} tx multisign` subcommand,
- add a new `Trace()` method on BaseApp to return the `trace` value for logging error stack traces.

See the [Cosmos SDK v0.42.7 milestone](https://github.com/cosmos/cosmos-sdk/milestone/48?closed=1) on our issue tracker for the exhaustive list of all changes.
