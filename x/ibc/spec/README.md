<!--
order: 0
title: IBC Overview
parent:
  title: "ibc"
-->

# `ibc`

## Abstract

This paper defines the implementation of the IBC protocol on the Cosmos SDK, the
changes made to the specification and where to find each specific ICS spec within
the module.

For the general specification please refer to the [Interchain Standards](https://github.com/cosmos/ics).

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transitions](02_state_transitions.md)**
4. **[Messages](03_messages.md)**
5. **[Callbacks](06_callbacks.md)**
6. **[Events](07_events.md)**

## Implementation Details

As stated above, the IBC implementation on the Cosmos SDK introduces some changes
to the general specification, in order to avoid code duplication and to take
advantage of the SDK architectural components such as the `AnteHandler` and
transaction routing through `Handlers`.

### Interchain Standards reference

The following list is a mapping from each Interchain Standard to their implementation
in the SDK's `x/ibc` module:

* [ICS 002 - Client Semantics](https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics): Implemented in [`x/ibc/02-client`](https://github.com/cosmos/x/ibc/02-client)
* [ICS 003 - Connection Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-003-connection-semantics): Implemented in [`x/ibc/03-connection`](https://github.com/cosmos/x/ibc/03-connection)
* [ICS 004 - Channel and Packet Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-004-channel-and-packet-semantics): Implemented in [`x/ibc/04-channel`](https://github.com/cosmos/x/ibc/04-channel)
* [ICS 005 - Port Allocation](https://github.com/cosmos/ics/blob/master/spec/ics-005-port-allocation): Implemented in [`x/ibc/05-port`](https://github.com/cosmos/x/ibc/05-port)
* [ICS 006 - Solo Machine Client](https://github.com/cosmos/ics/blob/master/spec/ics-006-solo-machine-client): Implemented in [`x/ibc/06-solomachine`](https://github.com/cosmos/x/ibc/06-solomachine)
* [ICS 007 - Tendermint Client](https://github.com/cosmos/ics/blob/master/spec/ics-007-tendermint-client): Implemented in [`x/ibc/07-tendermint`](https://github.com/cosmos/x/ibc/07-tendermint)
* [ICS 009 - Loopback Client](https://github.com/cosmos/ics/blob/master/spec/ics-009-loopback-client):  Implemented in [`x/ibc/09-localhost`](https://github.com/cosmos/x/ibc/09-localhost)
* [ICS 018- Relayer Algorithms](https://github.com/cosmos/ics/tree/master/spec/ics-018-relayer-algorithms): Implemented in it's own [relayer repository](https://github.com/cosmos/relayer)
* [ICS 020 - Fungible Token Transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer): Implemented in [`x/ibc-transfer`](https://github.com/cosmos/x/ibc-transfer)
* [ICS 023 - Vector Commitments](https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments): Implemented in [`x/ibc/23-commitment`](https://github.com/cosmos/x/ibc/23-commitment)
* [ICS 024 - Host Requirements](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements): Implemented in [`x/ibc/24-host`](https://github.com/cosmos/x/ibc/24-host)
* [ICS 025 - Handler Interface](https://github.com/cosmos/ics/tree/master/spec/ics-025-handler-interface): Handler interfaces are implemented at the top level in `x/ibc/handler.go`,
which call each ICS submodule's handlers (i.e `x/ibc/{XX-ICS}/handler.go`).
* [ICS 026 - Routing Module](https://github.com/cosmos/ics/blob/master/spec/ics-026-routing-module): Replaced by [ADR 15 - IBC Packet Receiver](../../../docs/architecture/adr-015-ibc-packet-receiver.md).

### Architecture Decision Records (ADR)

The following ADR provide the design and architecture decision of IBC-related components.

* [ADR 10 - Modular AnteHandler](../../../docs/architecture/adr-010-modular-antehandler.md): Introduces a decorator pattern for the [`AnteHandler`](../../../docs/basics/gas-fees.md#antehandler), making it modular.  
* [ADR 15 - IBC Packet Receiver](../../../docs/architecture/adr-015-ibc-packet-receiver.md): replaces the ICS26 routing module with [`AnteHandler`](../../../docs/basics/gas-fees.md#antehandler) logic within IBC. This is implemented using the `AnteDecorators` defined in [ADR10]((../../../docs/architecture/adr-010-modular-antehandler.md))
* [ADR 17 - Historical Header Module](../../../docs/architecture/adr-017-historical-header-module.md): Introduces the ability to introspect past
consensus states in order to verify their membership in the counterparty clients.
* [ADR 19 - Protobuf State Encoding](../../../docs/architecture/adr-019-protobuf-state-encoding.md): Migration from Amino to Protobuf for state encoding.
* [ADR 020 - Protocol Buffer Transaction Encoding](./../../docs/architecture/adr-020-protobuf-transaction-encoding.md): Client side migration to Protobuf.
* [ADR 021 - Protocol Buffer Query Encoding](./../../docs/architecture/adr-020-protobuf-query-encoding.md): Queries migration to Protobuf.

### SDK Modules

* [`x/capability`](https://github.com/cosmos/tree/master/x/capability): The capability module provides object-capability keys support through scoped keepers in order to authenticate usage of ports or channels. Check [ADR 3 - Dynamic Capability Store](../../../docs/architecture/adr-003-dynamic-capability-store.md) for more details.
* [`x/evidence`](https://github.com/cosmos/tree/master/x/evidence): The evidence module provides the interfaces and client logic to handle light client misbehaviour. Check [ADR 09 - Evidence Module](../../../docs/architecture/adr-009-evidence-module.md) for more details.

## IBC module architecture

> **NOTE for auditors**: If you're not familiar with the overall module structure from
the SDK modules, please check this [document](../../../docs/building-modules/structure.md) as
prerequisite reading.

For ease of auditing, every Interchain Standard has been developed in its own
package. The development team separated the IBC TAO (Transport, Authentication, Ordering) ICS specifications from the IBC application level
specification. The following tree describes the architecture of the directories that
the `ibc` (TAO) and `ibc-transfer` ([ICS20](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)) modules:

```shell
x/
├── ibc/
│  ├── 02-client/
│  ├── 03-connection/
│  ├── 04-channel/
│  ├── 05-port/
│  ├── 06-solo/
│  ├── 07-tendermint/
│  ├── 09-localhost/
│  ├── 23-commitment/
│  ├── 24-host/
│  ├── ante
│  │   └── ante.go
│  ├── client
│  │   ├── cli
│  │   │   └── cli.go
│  │   └── rest
│  │       └── rest.go
│  ├── keeper
│  │   ├── keeper.go
│  │   └── querier.go
│  ├── types
│  │   ├── errors.go
│  │   └── keys.go
│  ├── alias.go
│  ├── handler.go
│  └── module.go
└── ibc-transfer/
```
