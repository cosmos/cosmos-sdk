<!--
order: 0
title: Fee grant
parent:
  title: "feegrant"
-->

## Abstract

This document specifies the feegrant module. For the full ADR, please see [Fee Grant ADR-029](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/docs/architecture/adr-029-fee-grant-module.md).

This module allows accounts to grant fee allowances and to use fees from their accounts. Grantees can execute any transaction without the need to maintain sufficient fees.

## Contents

1. **[Concepts](01_concepts.md)**
    - [FeeAllowanceGrant](01_concepts.md#feeallowancegrant)
    - [Fee Allowance types](01_concepts.md#fee-allowance-types)
    - [BasicFeeAllowance](01_concepts.md#basicfeeallowance)
    - [PeriodicFeeAllowance](01_concepts.md#periodicfeeallowance)
    - [FeeAccount flag](01_concepts.md#feeaccount-flag)
    - [Gas](01_concepts.md#gas)
2. **[State](02_state.md)**
    - [FeeAllowance](02_state.md#feeallowance)
3. **[Messages](03_messages.md)**
    - [Msg/GrantFeeAllowance](03_messages.md#msggrantfeeallowance)
    - [Msg/RevokeFeeAllowance](03_messages.md#msgrevokefeeallowance)
4. **[Events](04_events.md)**
    - [MsgGrantFeeAllowance](04_events.md#msggrantfeeallowance)
    - [MsgrevokeFeeAllowance](04_events.md#msgrevokefeeallowance)
    - [Exec fee allowance](04_events.md#exec-fee-allowance)
    
