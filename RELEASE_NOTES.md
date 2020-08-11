# Cosmos SDK v0.39.1 Release Notes

This release fixes the [issue affecting the accounts migration](https://github.com/cosmos/cosmos-sdk/issues/6828) from v0.38 to v0.39.

See the [Cosmos SDK 0.39.1 milestone](https://github.com/cosmos/cosmos-sdk/milestone/29?closed=1) on our issue tracker for details.

## Remove custom JSON serialization for account types

Account types JSON serialization has now changed to Amino. Changes are significant (e.g. integers are treated
as strings) thus it is required to migrate the exported state of an application before restarting the node
with a more recent version of the Cosmos SDK.

## REST server's --unsafe-cors mode

This a UX improvement [back ported from master](https://github.com/cosmos/cosmos-sdk/pull/6853) that allows developers to disable CORS
restrictions during app development and testing by passing the `--unsafe-cors` option to the client's `rest-server` command.

## Tendermint 0.33.7

Tendermint 0.33.7 brings an important regression fix. Please refer to [this bug report](https://github.com/tendermint/tendermint/issues/5112) for more information.
