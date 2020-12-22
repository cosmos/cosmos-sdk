<!--
order: 0
title: Nameservice Overview
parent:
  title: "nameservice"
-->

# `x/nameservice`

## Table of Contents

<!-- TOC -->

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Keepers](03_keepers.md)**
4. **[Events](04_events.md)**
5. **[Messages](05_messages.md)**
6. **[Params](06_params.md)**

## Abstract

`x/nameservice` is an implementation of a Cosmos SDK module, per [ADR 049](/docs/architecture/adr-049-nameservice.md) <!-- placeholder mock ADR number and link for an undetermined ADR doc --> that handles the core logic that lets users buy names and set a value for the name resolution.


Each `Nameservice` type requires a `Handler` that is registered with the nameservice module keeper for successful routing and execution.

Each corresponding handler must fulfill the `Handler` interface contract. The `Handler` for a given `Nameservice` type can perform state transitions.
