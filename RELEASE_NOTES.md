# Cosmos SDK v0.47.0-rc2 Release Notes

Cosmos SDK `v0.47.0-rc2` contains all the features and changes that are planned for the final v0.47.0 release.
This release candidate is intended to give application developers and validator operators a chance to test the release candidate before the final release.

The main changes in this release are:

* Upgrade to Tendermint [v0.37.0](https://github.com/tendermint/tendermint/blob/v0.37.0-rc2/CHANGELOG_PENDING.md).
    * With the notable introduction of [ABCI 1.0](https://medium.com/the-interchain-foundation/tendermints-new-application-blockchain-interface-abci-86d46bd6f987).
    * Changes of events keys and values from `[]byte` to `string`.

* Support of [ABCI 1.0 in the SDK](https://docs.cosmos.network/v0.47/building-apps/app-mempool).
    * Allows chains to set their own mempool implementation.
    * Support of the new `PrepareProposal` and `ProcessProposal` ABCI methods.

* Deprecation of the `x/params` module.
    * Modules params are now handled directly by the modules themselves, with the message `MsgUpdateParams`.
    * All core modules have migrated away from `x/params`. It is recommended to migrate your custom modules as well.

* Migration from `gogo/protobuf` to `cosmos/gogoproto`.
    * The SDK was using the now unmaintained `gogo/protobuf` library. This has been replaced by [`cosmos/gogoproto`](https://github.com/cosmos/gogoproto) which is a fork of `gogo/protobuf` with some improvements and fixes, that is maintained by the Cosmos SDK team.
    * This change is not transparent for applications developers. All proto files should be regenerated with the new library.
    * Please use the `ghcr.io/cosmos/proto-builder` image (version >= `0.11.2`) for generating protobuf files.

* Dependency Injection / App Wiring
    * [App Wiring](https://docs.cosmos.network/v0.47/building-apps/app-go-v2) is ready for community feedback and testing. It allows to build a chain with less boilerplate by removing the need to manually wire a chain.
    * Community feedback will be implemented in the following releases which can lead to API breakage (`runtime` and [`depinject`](https://docs.cosmos.network/v0.47/tooling/depinject) are `pre-1.0`).
    * Manually wiring an application is still possible and will always remain supported.

* Removal of the [proposer-based rewards](https://github.com/cosmos/cosmos-sdk/issues/12667) from `x/distribution`.
    * This removes unfairness towards smaller validators.

* Re-addition of `title` and `summary` fields on group and gov proposals.
    * In `v0.46` with `x/gov` v1, these fields were not present (while present in `v1beta1`). After community feedback, they have been added in `x/gov` v1.

* Refactoring of tests in the SDK and addition of the [`simtestutil` package](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/testutil/sims), for facilitating testing without depending on simapp.
    * Any dependencies on `simapp` in an application must be removed going forward.

Refer to the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md) for upgrading your application.
Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History (previous version): https://github.com/cosmos/cosmos-sdk/compare/release/v0.46.x...release/v0.47.x
Full Commit History (`rc1..rc2`): https://github.com/cosmos/cosmos-sdk/compare/v0.47.0-rc1...v0.47.0-rc2
