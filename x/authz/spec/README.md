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
    - [Gas](01_concepts.md#gas)
2. **[State](02_state.md)**
3. **[Protobuf Services](03_services.md)**
    - [Grant](03_services.md#Grant)
    - [Revoke](03_services.md#Revoke)
    - [Exec](03_services.md#Exec)
4. **[Events](04_events.md)**
    - [Keeper](04_events.md#Keeper)
