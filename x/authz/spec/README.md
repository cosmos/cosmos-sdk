<!--
order: 0
title: Authz Overview
parent:
  title: "authz"
-->

# `authz`

## Contents

## Abstract
`x/authz` is an implementation of a Cosmos SDK module, per [ADR 30](../../../architecture/adr-030-authz-module.md), that allows 
granting arbitrary privileges from one account (the granter) to another account (the grantee). Authorizations must be granted for a particular Msg service method one by one using an implementation of the `Authorization` interface.

1. **[Concept](01_concepts.md)**
    - [Authorization](01_concepts.md#Authorization)
    - [Built-in Authorizations](01_concepts.md#Built-in-Authorization)
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
    - [Msg/GrantAuthorization](03_messages.md#MsgGrantAuthorization)
    - [Msg/RevokeAuthorization](03_messages.md#MsgRevokeAuthorization)
    - [Msg/ExecAuthorized](03_messages.md#MsgExecAuthorized)
4. **[Events](04_events.md)**
    - [Keeper](04_events.md#Keeper)
    
