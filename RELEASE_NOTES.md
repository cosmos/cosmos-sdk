# Cosmos SDK v0.47.0 Release Notes

✨ [**Official Release Announcenment**](https://blog.cosmos.network)

💬 [**Release Discussion**](https://github.com/cosmos/community)

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0/CHANGELOG.md) for an exhaustive list of changes or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.46.x...v0.47.0) from last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md) when migrating from `v0.46.x` to `v0.47.0`.

## 🚀 Highlights

* Upgrade to CometBFT [v0.37.0](https://github.com/cometbft/cometbft/blob/v0.37.0/CHANGELOG.md).
    * With the notable introduction of [ABCI 1.0](https://medium.com/the-interchain-foundation/tendermints-new-application-blockchain-interface-abci-86d46bd6f987).
    * Changes of events keys and values from `[]byte` to `string`.

* Support of [ABCI 1.0](https://medium.com/the-interchain-foundation/tendermints-new-application-blockchain-interface-abci-86d46bd6f987) in the SDK.
    * Allows chains to set their own mempool implementation. Follow the guide [here](https://docs.cosmos.network/v0.47/building-apps/app-mempool).
    * Support of the new `PrepareProposal` and `ProcessProposal` ABCI methods.

* Deprecation of the [`x/params`](https://docs.cosmos.network/v0.47/modules/params) module.
    * Modules params are now handled directly by the modules themselves, via the `MsgUpdateParams` message.
    * All core SDK modules have migrated away from using `x/params`. It is recommended to migrate your custom modules as well.

* Migration from `gogo/protobuf` to `cosmos/gogoproto`.
    * The SDK was using the now unmaintained `gogo/protobuf` library. This has been replaced by [`cosmos/gogoproto`](https://github.com/cosmos/gogoproto) which is a fork of `gogo/protobuf` with some improvements and fixes, that is maintained by the Cosmos SDK team.
    * This change is not transparent for applications developers. All proto files should be regenerated with the new library.
    * Use the `ghcr.io/cosmos/proto-builder` image (version >= `0.11.5`) for generating protobuf files.

* App Wiring with dependency injection.
    * [App Wiring](https://docs.cosmos.network/v0.47/building-apps/app-go-v2) is ready for community feedback. It allows developers to build a chain with less boilerplate by removing the need to manually wire a chain.
    * Community feedback will be implemented in the following releases which can lead to API breakage (`runtime` and [`depinject`](https://docs.cosmos.network/v0.47/tooling/depinject) are `pre-1.0`).
    * Manually wiring an application is still possible and will always remain supported.

* Removal of the [proposer-based rewards](https://github.com/cosmos/cosmos-sdk/issues/12667) from `x/distribution`.
    * This removes unfairness towards smaller validators.

* Re-addition of `title` and `summary` fields on group and gov proposals.
    * In `v0.46` with `x/gov` v1, these fields were not present (while present in `v1beta1`). After community feedback, they have been added in `x/gov` v1.

* Refactoring of tests in the SDK and addition of the [`simtestutil` package](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/testutil/sims), for facilitating testing without depending on simapp.
    * Any dependencies on `simapp` in an application must be removed going forward.

## ❤️ Contributors

* Binary Builders ([@binary_builders](https://twitter.com/binary_builders))
* Crypto.com ([@cronos_chain](https://twitter.com/cronos_chain))
* Interchain GmbH ([@interchain_io](https://twitter.com/interchain_io))
* Notional ([@notionaldao](https://twitter.com/notionaldao))
* Osmosis ([@osmosiszone](https://twitter.com/osmosiszone))
* Regen Network ([@regen_network](https://twitter.com/RegenNetworkDev))
* VitWit ([@vitwit_](https://twitter.com/vitwit_))

This list is non exhaustive and ordered alphabetically.  
Thank you to everyone who contributed to this release!
