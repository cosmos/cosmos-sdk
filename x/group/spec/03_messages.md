<!--
order: 3
-->

# Msg Service

## Msg/CreateGroup

A new group can be created with the `MsgCreateGroup`, which has an admin address, a list of members and some optional metadata.

The metadata has a maximum length that is chosen by the app developer, and
passed into the group keeper as a config.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L66-L78

It's expected to fail if

* metadata length is greater than `MaxMetadataLen`
  config
* members are not correctly set (e.g. wrong address format, duplicates, or with 0 weight).

## Msg/UpdateGroupMembers

Group members can be updated with the `UpdateGroupMembers`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L87-L100

In the list of `MemberUpdates`, an existing member can be removed by setting its weight to 0.

It's expected to fail if:

* the signer is not the admin of the group.
* for any one of the associated group policies, if its decision policy's `Validate()` method fails against the updated group.

## Msg/UpdateGroupAdmin

The `UpdateGroupAdmin` can be used to update a group admin.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L105-L117

It's expected to fail if the signer is not the admin of the group.

## Msg/UpdateGroupMetadata

The `UpdateGroupMetadata` can be used to update a group metadata.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L122-L134

It's expected to fail if:

* new metadata length is greater than `MaxMetadataLen` config.
* the signer is not the admin of the group.

## Msg/CreateGroupPolicy

A new group policy can be created with the `MsgCreateGroupPolicy`, which has an admin address, a group id, a decision policy and some optional metadata.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L143-L160

It's expected to fail if:

* the signer is not the admin of the group.
* metadata length is greater than `MaxMetadataLen` config.
* the decision policy's `Validate()` method doesn't pass against the group.

## Msg/CreateGroupWithPolicy

A new group with policy can be created with the `MsgCreateGroupWithPolicy`, which has an admin address, a list of members, a decision policy, a `group_policy_as_admin` field to optionally set group and group policy admin with group policy address and some optional metadata for group and group policy.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L183-L206

It's expected to fail for the same reasons as `Msg/CreateGroup` and `Msg/CreateGroupPolicy`.

## Msg/UpdateGroupPolicyAdmin

The `UpdateGroupPolicyAdmin` can be used to update a group policy admin.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L169-L181

It's expected to fail if the signer is not the admin of the group policy.

## Msg/UpdateGroupPolicyDecisionPolicy

The `UpdateGroupPolicyDecisionPolicy` can be used to update a decision policy.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L219-L235

It's expected to fail if:

* the signer is not the admin of the group policy.
* the new decision policy's `Validate()` method doesn't pass against the group.

## Msg/UpdateGroupPolicyMetadata

The `UpdateGroupPolicyMetadata` can be used to update a group policy metadata.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L240-L252

It's expected to fail if:

* new metadata length is greater than `MaxMetadataLen` config.
* the signer is not the admin of the group.

## Msg/SubmitProposal

A new proposal can be created with the `MsgSubmitProposal`, which has a group policy account address, a list of proposers addresses, a list of messages to execute if the proposal is accepted and some optional metadata.
An optional `Exec` value can be provided to try to execute the proposal immediately after proposal creation. Proposers signatures are considered as yes votes in this case.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L275-L298

It's expected to fail if:

* metadata length is greater than `MaxMetadataLen` config.
* if any of the proposers is not a group member.

## Msg/WithdrawProposal

A proposal can be withdrawn using `MsgWithdrawProposal` which has an `address` (can be either a proposer or the group policy admin) and a `proposal_id` (which has to be withdrawn).

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L307-L316

It's expected to fail if:

* the signer is neither the group policy admin nor proposer of the proposal.
* the proposal is already closed or aborted.

## Msg/Vote

A new vote can be created with the `MsgVote`, given a proposal id, a voter address, a choice (yes, no, veto or abstain) and some optional metadata.
An optional `Exec` value can be provided to try to execute the proposal immediately after voting.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L321-L339

It's expected to fail if:

* metadata length is greater than `MaxMetadataLen` config.
* the proposal is not in voting period anymore.

## Msg/Exec

A proposal can be executed with the `MsgExec`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L341-L353

The messages that are part of this proposal won't be executed if:

* the proposal has not been accepted by the group policy.
* the proposal has already been successfully executed.

## Msg/LeaveGroup

The `MsgLeaveGroup` allows group member to leave a group.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/group/v1/tx.proto#L362-L370

It's expected to fail if:

* the group member is not part of the group.
* for any one of the associated group policies, if its decision policy's `Validate()` method fails against the updated group.
