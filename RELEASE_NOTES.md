# Cosmos SDK v0.46.0 Release Candidate 1 Release Notes

This release introduces several new important updates to the Cosmos SDK. The release notes below provide an overview of the larger high-level changes introduced in the v0.46 release series.

That being said, this release does contain many more minor and module-level changes besides those mentioned below. For a comprehsive list of all breaking changes and improvements since the v0.45 release series, please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md).

## One new module: `x/group`

The previous v0.43 series focused on simplifying [keys and fee management](https://github.com/cosmos/cosmos-sdk/issues/7074) for SDK users, by adding `x/feegrant` and `x/authz`. v0.46 finishes this work by introducing `x/group`.

`x/group` provides functionality to define on-chain groups of people that can execute arbitrary messages based on pre-defin. A simple use-case of `x/group` is to create on-chain multisigs (with updateable members and thresholds), but `x/group` can also be used to create more complex DAOs.

The `x/group` module resolves around 3 concepts:

- A **group** is simply an aggregation of accounts with associated weights.
- A **group policy** is a group with a set of rules attached. This set of rules (called decision policy) defines how voting and arbitrary message execution happens (e.g. does a proposal pass on 50% yes? 2/3 yes? is there a way to veto? etc). Each group policy has its own an on-chain account, so can hold funds. Managing group membership separately from decision policies results in the least overhead and keeps membership consistent across different policies.
- Any member of a group can submit a **proposal** for a group policy account to decide upon. A proposal consists of a set of messages that will be executed if the proposal passes voting.

For more details about `x/group`, please refer to [the SDK documentation](https://docs.cosmos.network/master/modules/group/) and [ADR-042](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-042-group-module.md).

The folder structure of `x/group` contains an `internal` folder, which holds a custom ORM used only by `x/group` (and which will be replaced by the [new ORM](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-055-orm.md)) as well as a new implementation of `Dec` (for decimals) based on [`cockroachdb/apd`](https://github.com/cockroachdb/apd), which serves as a proof-of-concept for the [new `sdk.Dec`](https://github.com/cosmos/cosmos-sdk/issues/11783).

## `Msg`-based Gov Proposals

In an effort to [align](https://github.com/cosmos/cosmos-sdk/issues/9438) `x/gov` with `x/group`, the SDK v0.46 release introduces a new Protobuf package: `cosmos.gov.v1`. Some API changes from `v1beta1` to `v1` are expected in the `x/gov` Keeper, and are documented in the [`UPGRADING.md` guide](TODO).

The biggest change is in `MsgSubmitProposal`: instead of defining gov router proposal handlers, the v0.46 gov execution models is based on `sdk.Msg`s:

```diff
message MsgSubmitProposal {
-  google.protobuf.Any content                       = 1 [(cosmos_proto.accepts_interface) = "Content"];
+  repeated google.protobuf.Any messages             = 1;
  repeated cosmos.base.v1beta1.Coin initial_deposit = 2 [(gogoproto.nullable) = false];
  string                            proposer        = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
+  // metadata is any arbitrary metadata attached to the proposal.
+  string metadata = 4;
}
```

For example, instead of broadcasting a `v1beta1.MsgSubmitProposal` with content a `SoftwareUpgradeProposal`, the proposer would submit a `v1.MsgSubmitProposal` with a `cosmos.upgrade.v1beta1.MsgSoftwareUpgrade` message. When the proposal passes, the `sdk.Msg` will be executed by the `Msg` service router (instead of going through the gov proposal handlers).

From a client perspective, the new gov v1 is purely additive. All `v1beta1` Protobuf defintions, queries and `Msg`s still work. Morever, users can also submit `v1beta1` legacy proposals using the `v1` `Msg` service, by including a `MsgExecLegacyContent` inside the `v1.MsgSubmitProposal`. It is recommended to switch to gov `v1` during v0.46, as the gov `v1beta1` backwards-compatibility might be removed in a future version.
