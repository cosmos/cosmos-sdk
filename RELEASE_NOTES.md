# Cosmos SDK v0.46.0 Release Notes

This release introduces several new important updates to the Cosmos SDK. The release notes below provide an overview of the larger high-level changes introduced in the v0.46 release series.

That being said, this release does contain many more minor and module-level changes besides those mentioned below. For a comprehsive list of all breaking changes and improvements since the v0.45 release series, please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md).

## New Module: `x/group`

The previous v0.43 series focused on simplifying [keys and fee management](https://github.com/cosmos/cosmos-sdk/issues/7074) for SDK users, by adding `x/feegrant` and `x/authz`. v0.46 finishes this work by introducing `x/group`.

`x/group` provides functionality to define on-chain groups of people that can execute arbitrary messages based on pre-defin. A simple use-case of `x/group` is to create on-chain multisigs (with updateable members and thresholds), but `x/group` can also be used to create more complex DAOs.

The `x/group` module resolves around 3 concepts:

- A **group** is simply an aggregation of accounts with associated weights.
- A **group policy** is a group with a set of rules attached. This set of rules (called decision policy) defines how voting and arbitrary message execution happens (e.g. does a proposal pass on 50% yes? 2/3 yes? is there a way to veto? etc). Each group policy has its own an on-chain account, so can hold funds. Managing group membership separately from decision policies results in the least overhead and keeps membership consistent across different policies.
- Any member of a group can submit a **proposal** for a group policy account to decide upon. A proposal consists of a set of messages that will be executed if the proposal passes voting.

For more details about `x/group`, please refer to [the SDK documentation](https://docs.cosmos.network/master/modules/group/) and [ADR-042](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-042-group-module.md).

The folder structure of `x/group` contains an `internal` folder, which holds a custom ORM used only by `x/group` (and which will be replaced by the [new ORM](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-055-orm.md)) as well as a new implementation of `Dec` (for decimals) based on [`cockroachdb/apd`](https://github.com/cockroachdb/apd), which serves as a proof-of-concept for the [new `sdk.Dec`](https://github.com/cosmos/cosmos-sdk/issues/11783).

## `Msg`-based Gov Proposals

In an effort to [align](https://github.com/cosmos/cosmos-sdk/issues/9438) `x/gov` with `x/group`, the SDK v0.46 release introduces a new Protobuf package: `cosmos.gov.v1`.

The biggest change compared to the previous `cosmoss.gov.v1beta1` is in `MsgSubmitProposal`: instead of defining gov router proposal handlers, the v0.46 gov execution models is based on `sdk.Msg`s:

```diff
message MsgSubmitProposal {
-  google.protobuf.Any content                       = 1 [(cosmos_proto.accepts_interface) = "Content"];
+  repeated google.protobuf.Any messages             = 1 [(cosmos_proto.accepts_interface) = "sdk.Msg"];
  repeated cosmos.base.v1beta1.Coin initial_deposit = 2 [(gogoproto.nullable) = false];
  string                            proposer        = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
+  // metadata is any arbitrary metadata attached to the proposal.
+  string metadata = 4;
}
```

For example, instead of broadcasting a `v1beta1.MsgSubmitProposal` with content a `SoftwareUpgradeProposal`, the proposer would submit a `v1.MsgSubmitProposal` with a `cosmos.upgrade.v1beta1.MsgSoftwareUpgrade` message. When the proposal passes, the `sdk.Msg` will be executed by the `Msg` service router (instead of going through the gov proposal handlers).

A `metadata` field has also been added to `MsgSubmitProposal` and `MsgVote`, for users to provide optional justification for their action.

From a client perspective, the new gov v1 is purely additive. All `v1beta1` Protobuf defintions, queries and `Msg`s still work. Morever, users can also submit `v1beta1` legacy proposals using the `v1` `Msg` service, by including a `MsgExecLegacyContent` inside the `v1.MsgSubmitProposal`. It is recommended to switch to gov `v1` during v0.46, as the gov `v1beta1` backwards-compatibility might be removed in a future version.

As an app developer, some API changes from `v1beta1` to `v1` are to be expected in the `x/gov` Keeper, and are documented in the [`UPGRADING.md` guide](TODO).

## Antehandlers are Replaced by Middlewares

⚠️ Middlewares' design is still being discussed in [#11955](https://github.com/cosmos/cosmos-sdk/issues/11955). Please note that API-breaking changes still may happen during the Release Candidate phrase.

Antehandlers were defined as the logic that was run by BaseApp's `CheckTx` and `DeliverTx` ABCI methods before running the actual transaction body's `sdk.Msg`s. One limitation they had was that we couldn't write logic for _post-_`runMsgs` execution, which is required by transaction tips (see below), or in the future, [fee refunds](https://github.com/cosmos/cosmos-sdk/issues/2150).

[ADR-045](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-045-check-delivertx-middlewares.md) proposes a new design called "middlewares", which are similar to antehandlers, except that each middleware can optionally also execute some custom logic after `runMsgs` execution. In fact, `runMsgs` itself has been transformed into a middleware, making BaseApp more lightweight.

Each antehandler has been converted into its equivalent middleware in v0.46. Some new middlewares have also been created, mostly to move transaction-related logic out of BaseApp.

## Transaction Tips and `SIGN_MODE_DIRECT_AUX`

Transaction tips are a mechanism to pay for transaction fees using another denom than the native fee denom of the chain.

The transaction initiator signs a partial transaction (called `AuxSignerData`) without specifying fees, but uses a new `Tip` field. They send this `AuxSignerData` to a fee relayer who will choose the transaction fees and broadcast the final transaction, and the SDK provides a mechanism that will transfer the pre-defined `Tip` to the fee payer, to cover for fees. A market between tippers and feepayers could arise, based on exchange rates between the tip denom and the fee denom.

For this mechanism to work, the SDK introduces a new sign mode, `SIGN_MODE_DIRECT_AUX`, whereby the signer signs over the transaction body and their own signer info, but not over fees or other signers' info. This sign mode is not limited to transaction tips though, and can be used in any multi-signer transaction, where N-1 signers sign using `SIGN_MODE_DIRECT_AUX`, and only one signer, the fee payer, signs using `SIGN_MODE_DIRECT`, allowing for a better UX for the N-1 other signers.
