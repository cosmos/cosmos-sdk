# Cosmos SDK v0.39.1 Release Notes

See the [Cosmos SDK 0.39.1 milestone](https://github.com/cosmos/cosmos-sdk/milestone/29?closed=1) on our issue tracker for details.

## Remove custom JSON serialization for account types

Account types JSON serialization has now changed to Amino. Changes are significant (e.g. integers are treated
as strings) thus it is required to migrate the exported state of an application before restarting the node
with a more recent version of the Cosmos SDK.
