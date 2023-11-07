# Cosmos SDK v0.50.1 Release Notes

✨ [**Official Release Announcement**](https://blog.cosmos.network)

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

Cosmos SDK v0.50 is a major release that includes a number of significant new features and improvements. These new features and improvements will make Cosmos SDK applications more performant, scalable, and secure. They will also make it easier for developers to create and integrate new modules into the Cosmos SDK ecosystem.

* **ABCI 2.0 Integration:** Cosmos SDK v0.50 upgrades to CometBFT v0.38 and fully implements ABCI 2.0.
* **Optimistic Execution:** Cosmos SDK v0.50 introduces Optimistic Execution, which allows transactions to be executed and committed without waiting for confirmation from all validators. This can significantly improve the performance of chains with a high volume of transactions.
* **Modular SDK modules:** Cosmos SDK v0.50 starts to extract core modules away from the SDK. These are separately versioned and follow their own release cadence.
* **IAVL v1:** Cosmos SDK v0.50 upgrades the IAVL tree implementation to v1, which provides a number of performance and security improvements.
* **AutoCLI:** Cosmos SDK v0.50 introduces AutoCLI, a library that makes it easier to create CLI commands for SDK applications.
* **Sign Mode Textual:** Cosmos SDK v0.50 introduces a new sign mode that for hardware devices, as a replacement of Amino JSON.
* **Less boilerplate:** Cosmos SDK v0.50 requires less boilerplate in general for modules code and applications. Be sure to read the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) to take advantage of these improvements.

### ABCI 2.0

Cosmos SDK v0.50 upgrades CometBFT to CometBFT [v0.38.0](https://github.com/cometbft/cometbft/blob/v0.38.0/CHANGELOG.md) and integrates [ABCI 2.0](https://github.com/cometbft/cometbft/tree/v0.38.0/spec/abci) semantics. Modules still follow ABCI 1.0 sementics (`BeginBlock`, `EndBlock`).

For instance, applications can now support [Vote Extensions](https://docs.cosmos.network/v0.50/build/building-apps/vote-extensions).

### Optimistic Execution

Cosmos SDK v0.50 introduces Optimistic Execution, which allows transactions to be executed and committed without waiting for confirmation from all validators. This can significantly improve the performance of chains with a high volume of transactions.

Optimistic Execution leverages ABCI 2.0, and is disabled by default. To enable it add `baseapp.SetOptimisticExecution()` to your baseapp options in your `app.go`.

### SDK modules

Cosmos SDK v0.50 starts to extract core modules away from the SDK. These are separately versioned and follow their own release cadence.

It starts with `x/evidence`, `x/feegrant`, `x/nft`, and `x/upgrade`,

Additionally, SDK v0.50 introduces a new core module, `x/circuit` that provides a circuit breaker for the SDK. Read more about it in the [module documentation](https://docs.cosmos.network/v0.50/build/modules/circuit).

Lastly, `x/capability` module has moved to the [IBC repo](https://github.com/cosmos/ibc-go) and is now maintained by the IBC team.

The further decoupling of other core modules is planned for the next release.

### Store v1 and IAVL v1

Cosmos SDK v0.50 has decoupled its store from the SDK. The store is now versioned separately and follows its own release cadence.

Store v1 upgrades the IAVL tree implementation to v1. IAVL v1 provides a number of performance improvements.
Read more about it in the [IAVL repo](https://github.com/cosmos/iavl/releases/tag/v1.0.0).

### AutoCLI

Cosmos SDK v0.50 introduces AutoCLI, a library that makes it easier to create CLI commands for SDK applications.
Forget the boilerplate that was required to create CLI commands for SDK applications. AutoCLI will generate CLI commands for you.
Read more about it in the [AutoCLI docs](https://docs.cosmos.network/v0.50/learn/advanced/autocli)

### Sign Mode Textual

Cosmos SDK v0.50 introduces a new sign mode mainly for hardware wallets, as a replacement of Amino JSON.
Never leak again that you are signing from a Ledger device and sign with Sign Mode Textual everywhere.

### Less Boilerplate

Cosmos SDK v0.50 requires less boilerplate in general for modules code and applications.

Here's a sneak peek of what you can expect:

Next to module CLI code that can be removed thanks to AutoCLI, modules do not need to implement `ValidateBasic()`, or `GetSigners` anymore.
The checks can happen directly on the message server, and the signers can be retrieved from the message itself (thanks to a protobuf annotation).

Be sure to [annotate your proto messages properly](https://docs.cosmos.network/v0.50/build/building-modules/protobuf-annotations) to take advantage of those improvements.

Read the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) for a more exhaustive list of changes.

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.1/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.47.x...v0.50.1) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.51.0, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.

## ❤️ Contributors

* Binary Builders ([@binary_builders](https://twitter.com/binary_builders))
* Crypto.com ([@cronos_chain](https://twitter.com/cronos_chain))
* Orijtech ([@orijtech](https://twitter.com/orijtech))
* VitWit ([@vitwit_](https://twitter.com/vitwit_))
* Zondax ([@\_zondax\_](https://twitter.com/_zondax_))

This list is non exhaustive and ordered alphabetically.  
Thank you to everyone who contributed to this release!
