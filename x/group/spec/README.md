<!--
order: 0
title: Group Overview
parent:
  title: "group"
-->

# Group Module

## Abstract

The following documents specify the group module.

This module allows the creation and management of on-chain multisig accounts and enables voting for message execution based on configurable decision policies.

## Contents

1. **[Concepts](01_concepts.md)**
    - [Group](01_concepts.md#group)
    - [Group Account](01_concepts.md#group-account)
    - [Decision Policy](01_concepts.md#decision-policy)
    - [Proposal](01_concepts.md#proposal)
    - [Voting](01_concepts.md#voting)
    - [Executing Proposals](01_concepts.md#executing-proposals)
2. **[State](02_state.md)**
    - [Group Table](02_state.md#group-table)
    - [Group Member Table](02_state.md#group-member-table)
    - [Group Account Table](02_state.md#group-account-table)
    - [Proposal](02_state.md#proposal-table)
    - [Vote Table](02_state.md#vote-table)
3. **[Msg Service](03_messages.md)**
    - [Msg/CreateGroup](03_messages.md#msgcreategroup)
    - [Msg/UpdateGroupMembers](03_messages.md#msgupdategroupmembers)
    - [Msg/UpdateGroupAdmin](03_messages.md#msgupdategroupadmin)
    - [Msg/UpdateGroupMetadata](03_messages.md#msgupdategroupmetadata)
    - [Msg/CreateGroupAccount](03_messages.md#msgcreategroupaccount)
    - [Msg/UpdateGroupAccountAdmin](03_messages.md#msgupdategroupaccountadmin)
    - [Msg/UpdateGroupAccountDecisionPolicy](03_messages.md#msgupdategroupaccountdecisionpolicy)
    - [Msg/UpdateGroupAccountMetadata](03_messages.md#msgupdategroupaccountmetadata)
    - [Msg/CreateProposal](03_messages.md#msgcreateproposal)
    - [Msg/Vote](03_messages.md#msgvote)
    - [Msg/Exec](03_messages.md#msgexec)
4. **[Events](04_events.md)**
    - [EventCreateGroup](04_events.md#eventcreategroup)
    - [EventUpdateGroup](04_events.md#eventupdategroup)
    - [EventCreateGroupAccount](04_events.md#eventcreategroupaccount)
    - [EventUpdateGroupAccount](04_events.md#eventupdategroupaccount)
    - [EventCreateProposal](04_events.md#eventcreateproposal)
    - [EventVote](04_events.md#eventvote)
    - [EventExec](04_events.md#eventexec)
5. **[Client](05_client.md)**
    - [CLI](05_client.md#cli)
    - [gRPC](05_client.md#grpc)
    - [REST](05_client.md#rest)
