<!--
order: 0
title: Sanction Overview
parent:
  title: "sanction"
-->

# Sanction Module

## Abstract

This module allows management of a list of sanctioned accounts that are prevented from sending or spending any funds.
It injects a restriction into the `x/bank` module to enforce these sanctions.

## Contents

1. **[Concepts](01_concepts.md)**
   - [Sanctioned Account](01_concepts.md#sanctioned-account)
   - [Immediate Temporary Sanctions](01_concepts.md#immediate-temporary-sanctions)
   - [Unsanctioning](01_concepts.md#unsanctioning)
   - [Immediate Temporary Unsanctions](01_concepts.md#immediate-temporary-unsanctions)
   - [Unsanctionable Addresses](01_concepts.md#unsanctionable-addresses)
   - [Params](01_concepts.md#params)
   - [Complex Interactions](01_concepts.md#complex-interactions)
2. **[State](02_state.md)**
   - [Params](02_state.md#params)
   - [Sanctioned Accounts](02_state.md#sanctioned-accounts)
   - [Temporary Entries](02_state.md#temporary-entries)
   - [Temporary Index](02_state.md#temporary-index)
3. **[Msg Service](03_messages.md)**
   - [Msg/Sanction](03_messages.md#msgsanction)
   - [Msg/Unsanction](03_messages.md#msgunsanction)
   - [Msg/UpdateParams](03_messages.md#msgupdateparams)
4. **[Events](04_events.md)**
   - [EventAddressSanctioned](04_events.md#eventaddresssanctioned)
   - [EventAddressUnsanctioned](04_events.md#eventaddressunsanctioned)
   - [EventTempAddressSanctioned](04_events.md#eventtempaddresssanctioned)
   - [EventTempAddressUnsanctioned](04_events.md#eventtempaddressunsanctioned)
   - [EventParamsUpdated](04_events.md#eventparamsupdated)
5. **[gRPC Queries](05_queries.md)**
   - [Query/IsSanctioned](05_queries.md#queryissanctioned)
   - [Query/SanctionedAddresses](05_queries.md#querysanctionedaddresses)
   - [Query/TemporaryEntries](05_queries.md#querytemporaryentries)
   - [Query/Params](05_queries.md#queryparams)
6. **[Client](06_client.md)**
   - [gRPC](06_client.md#grpc)
   - [CLI](06_client.md#cli)
   - [REST](06_client.md#rest)