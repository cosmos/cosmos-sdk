# cosmos-sdk

This repo is a fork of [cosmos/cosmos-sdk](https://github.com/cosmos/cosmos-sdk) with a few modifications.

## Modifications

Larger modifications include:

1. Early adoption of `PrepareProposal` and `ProcessProposal`. This was added to the fork because at the time of development, a Cosmos SDK release was not available with these ABCI methods. Ref: https://github.com/celestiaorg/cosmos-sdk/commit/233a229cabf0599aed91b6b6697c268753731b2c
1. The addition of `chainID` to baseapp so that a branch of state can be used in PrepareProposal and ProcessProposal. Ref: https://github.com/celestiaorg/cosmos-sdk/pull/326
1. The consensus params version is overriden to the `AppVersion` to enable EndBlocker to update the `AppVersion`. Ref: https://github.com/celestiaorg/cosmos-sdk/pull/321

Smaller modifications include:

1. The addition of a `SetTxDecoder` on tx config so that celestia-app can override the default tx decoder with one that supports decoding `BlobTx`s. Ref: https://github.com/celestiaorg/cosmos-sdk/pull/311
1. The addition of a `start_time` to the vesting module's `MsgCreateVestingAccount` so that vesting accounts can be created with a delayed start time. Ref: https://github.com/celestiaorg/cosmos-sdk/pull/342
1. Allow celestia-app to override the default consensus params via the `init` command. Ref: https://github.com/celestiaorg/cosmos-sdk/pull/317

Modifications that make it easier to maintain this fork:

1. Modify CODEOWNERS to Celestia maintainers
1. Modify Github CI workflows to include `release/**` branches
1. Modify Github CI workflows to not run some workflows
1. Delete cosmovisor

Modifications that may be revertable:

1. Override the default keyringBackend from `os` to `test`. Maybe move to celestia-app
1. Increase `DefaultGasLimit` from 200000 to 210000.
1. Remove `Evidence` from grpc/tmservice/types.pb.go.
1. Override simapp test helpers `DefaultGenTxGas` from 10000000 to 2600000.
1. Disable staticcheck golangci lint after fixing lint errors.
1. In auth/tx/query.go disable the prove flag when querying transactions
1. In server/util.go remove `conf.Consensus.TimeoutCommit = 5 * time.Second`

## Branches

1. [v0.46.x-celestia](https://github.com/celestiaorg/cosmos-sdk/tree/release/v0.46.x-celestia) is based on the `v0.46.x` release branch from upstream

## Contributing

This repo intends on preserving the minimal possible diff with [cosmos/cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to make fetching upstream changes easy. If the proposed contribution is

* specific to Celestia: consider if [celestia-app](https://github.com/celestiaorg/celestia-app) is a better target
* not specific to Celestia: consider making the contribution upstream in [cosmos/cosmos-sdk](https://github.com/cosmos/cosmos-sdk)
