<!--
order: 0
title: Fee grant
parent:
  title: "feegrant"
-->

# Fee grant

## Abstract

This document specifies the fee grant module. For the full ADR, please see [Fee Grant ADR-029](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/docs/architecture/adr-029-fee-grant-module.md).

This module allows accounts to grant fee allowances and to use fees from their accounts. Grantees can execute any transaction without the need to maintain sufficient fees.

## Contents

1. **[Concepts](01_concepts.md)**
    * [Grant](01_concepts.md#grant)
    * [Fee Allowance types](01_concepts.md#fee-allowance-types)
    * [BasicAllowance](01_concepts.md#basicallowance)
    * [PeriodicAllowance](01_concepts.md#periodicallowance)
    * [FeeAccount flag](01_concepts.md#feeaccount-flag)
    * [Granted Fee Deductions](01_concepts.md#granted-fee-deductions)
    * [Gas](01_concepts.md#gas)
2. **[State](02_state.md)**
    * [FeeAllowance](02_state.md#feeallowance)
3. **[Messages](03_messages.md)**
    * [Msg/GrantAllowance](03_messages.md#msggrantallowance)
    * [Msg/RevokeAllowance](03_messages.md#msgrevokeallowance)
4. **[Events](04_events.md)**
    * [MsgGrantAllowance](04_events.md#msggrantallowance)
    * [MsgRevokeAllowance](04_events.md#msgrevokeallowance)
    * [Exec fee allowance](04_events.md#exec-fee-allowance)
5. **[Client](05_client.md)**
    * [CLI](05_client.md#cli)
    * [gRPC](05_client.md#grpc)
