<!--
order: 0
title: Quarantine Overview
parent:
  title: "quarantine"
-->

# Quarantine Module

## Abstract

This module allows management of quarantined accounts and funds.
It also injects restrictions into the `x/bank` module to enforce account quarantines.

## Contents

1. **[Concepts](01_concepts.md)**
   - [Quarantined Account](01_concepts.md#quarantined-account)
   - [Opt-In](01_concepts.md#opt-in)
   - [Quarantined Funds](01_concepts.md#quarantined-funds)
   - [Auto-Responses](01_concepts.md#auto-responses)
2. **[State](02_state.md)**
   - [Quarantined Accounts](02_state.md#quarantined-accounts)
   - [Auto-Responses](02_state.md#auto-responses)
   - [Quarantine Records](02_state.md#quarantine-records)
   - [Quarantine Records Suffix Index](02_state.md#quarantine-records-suffix-index)
3. **[Msg Service](03_messages.md)**
   - [Msg/OptIn](03_messages.md#msgoptin)
   - [Msg/OptOut](03_messages.md#msgoptout)
   - [Msg/Accept](03_messages.md#msgaccept)
   - [Msg/Decline](03_messages.md#msgdecline)
   - [Msg/UpdateAutoResponses](03_messages.md#msgupdateautoresponses)
4. **[Events](04_events.md)**
   - [EventOptIn](04_events.md#eventoptin)
   - [EventOptOut](04_events.md#eventoptout)
   - [EventFundsQuarantined](04_events.md#eventfundsquarantined)
   - [EventFundsReleased](04_events.md#eventfundsreleased)
5. **[gRPC Queries](05_queries.md)**
   - [Query/IsQuarantined](05_queries.md#queryisquarantined)
   - [Query/QuarantinedFunds](05_queries.md#queryquarantinedfunds)
   - [Query/AutoResponses](05_queries.md#queryautoresponses)
6. **[Client](06_client.md)**
   - [gRPC](06_client.md#grpc)
   - [CLI](06_client.md#cli)
   - [REST](06_client.md#rest)
