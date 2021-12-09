<!--
order: 3
-->

# Msg Service

## Msg/CreateGroup

A new group can be created with the `MsgCreateGroupRequest`, which has an admin address, a list of members and some optional metadata bytes.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L53-L64

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/UpdateGroupMembers

Group members can be updated with the `UpdateGroupMembersRequest`.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L73-L85

In the list of `MemberUpdates`, an existing member can be removed by setting its weight to 0.

It's expecting to fail if the signer is not the admin of the group.

## Msg/UpdateGroupAdmin

The `UpdateGroupAdminRequest` can be used to update a group admin.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L90-L101

It's expecting to fail if the signer is not the admin of the group.

## Msg/UpdateGroupMetadata

The `UpdateGroupMetadataRequest` can be used to update a group metadata.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L106-L117

It's expecting to fail if:
- new metadata length is greater than some `MaxMetadataLength`.
- the signer is not the admin of the group.

## Msg/CreateGroupAccount

A new group account can be created with the `MsgCreateGroupAccountRequest`, which has an admin address, a group id, a decision policy and some optional metadata bytes.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L126-L141

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/UpdateGroupAccountAdmin

The `UpdateGroupAccountAdminRequest` can be used to update a group account admin.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L150-L161

It's expecting to fail if the signer is not the admin of the group account.

## Msg/UpdateGroupAccountDecisionPolicy

The `UpdateGroupAccountDecisionPolicyRequest` can be used to update a decision policy.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L166-L178

It's expecting to fail if the signer is not the admin of the group account.

## Msg/UpdateGroupAccountMetadata

The `UpdateGroupAccountMetadataRequest` can be used to update a group account metadata.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L183-L194

It's expecting to fail if:
- new metadata length is greater than some `MaxMetadataLength`.
- the signer is not the admin of the group.

## Msg/CreateProposal

A new group account can be created with the `MsgCreateProposalRequest`, which has a group account address, a list of proposers addresses, a list of messages to execute if the proposal is accepted and some optional metadata bytes.
An optional `Exec` value can be provided to try to execute the proposal immediately after proposal creation. Proposers signatures are considered as yes votes in this case.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L217-L238

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/Vote

A new vote can be created with the `MsgVoteRequest`, given a proposal id, a voter address, a choice (yes, no, veto or abstain) and some optional metadata bytes.
An optional `Exec` value can be provided to try to execute the proposal immediately after voting.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L247-L265

It's expecting to fail if metadata length is greater than some `MaxMetadataLength`.

## Msg/Exec

A proposal can be executed with the `MsgExecRequest`.

+++ https://github.com/regen-network/regen-ledger/blob/8cebfb2d0dd000c42ae4d2da583629fdb96966c0/proto/regen/group/v1alpha1/tx.proto#L270-L278

The messages that are part of this proposal won't be executed if:
- the group has been modified before tally.
- the group account has been modified before tally.
- the proposal has not been accepted.
- the proposal status is not closed.
- the proposal has already been successfully executed.