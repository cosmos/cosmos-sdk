<!--
order: 3
-->

# Msg Service

## Msg/CreateGroup

A new group can be created with the `MsgCreateGroup`, which has an admin address, a list of members and some optional metadata bytes.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L54-L65

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/UpdateGroupMembers

Group members can be updated with the `UpdateGroupMembers`.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L74-L86

In the list of `MemberUpdates`, an existing member can be removed by setting its weight to 0.

It's expecting to fail if the signer is not the admin of the group.

## Msg/UpdateGroupAdmin

The `UpdateGroupAdmin` can be used to update a group admin.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L91-L102

It's expecting to fail if the signer is not the admin of the group.

## Msg/UpdateGroupMetadata

The `UpdateGroupMetadata` can be used to update a group metadata.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L107-L118

It's expecting to fail if:
- new metadata length is greater than some `MaxMetadataLength`.
- the signer is not the admin of the group.

## Msg/CreateGroupAccount

A new group account can be created with the `MsgCreateGroupAccount`, which has an admin address, a group id, a decision policy and some optional metadata bytes.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L121-L142

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/UpdateGroupAccountAdmin

The `UpdateGroupAccountAdmin` can be used to update a group account admin.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L151-L162

It's expecting to fail if the signer is not the admin of the group account.

## Msg/UpdateGroupAccountDecisionPolicy

The `UpdateGroupAccountDecisionPolicy` can be used to update a decision policy.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L167-L179

It's expecting to fail if the signer is not the admin of the group account.

## Msg/UpdateGroupAccountMetadata

The `UpdateGroupAccountMetadata` can be used to update a group account metadata.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L184-L195

It's expecting to fail if:
- new metadata length is greater than some `MaxMetadataLength`.
- the signer is not the admin of the group.

## Msg/CreateProposal

A new group account can be created with the `MsgCreateProposal`, which has a group account address, a list of proposers addresses, a list of messages to execute if the proposal is accepted and some optional metadata bytes.
An optional `Exec` value can be provided to try to execute the proposal immediately after proposal creation. Proposers signatures are considered as yes votes in this case.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L218-L239

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/Vote

A new vote can be created with the `MsgVote`, given a proposal id, a voter address, a choice (yes, no, veto or abstain) and some optional metadata bytes.
An optional `Exec` value can be provided to try to execute the proposal immediately after voting.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L248-L265

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/Exec

A proposal can be executed with the `MsgExec`.

+++ https://github.com/cosmos/cosmos-sdk/blob/6f58963e7f6ce820e9b33f02f06f7b96f6d2e347/proto/cosmos/group/v1beta1/tx.proto#L270-L278

The messages that are part of this proposal won't be executed if:
- the group has been modified before tally.
- the group account has been modified before tally.
- the proposal has not been accepted.
- the proposal status is not closed.
- the proposal has already been successfully executed.