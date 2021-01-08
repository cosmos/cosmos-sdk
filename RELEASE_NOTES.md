# Cosmos SDK v0.40.0 "Stargate" Release Notes

This release introduces several new important updates to the Cosmos SDK. The release notes below provide an overview of
the larger high-level changes introduced in the v0.40 (aka Stargate) release series.

That being said, this release does contain many more minor and module-level changes besides those mentioned below. For a
comprehsive list of all breaking changes and improvements since the v0.39 release series, please see the
[changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/CHANGELOG.md).

## Protocol Buffer Migration

Stargate introduces [protocol buffers](https://developers.google.com/protocol-buffers) as the new standard serialization
format for blockchain state & wire communication within the Cosmos SDK. Protobuf definitions are organized into packages
that mirror Cosmos SDK modules in the new [./proto](https://github.com/cosmos/cosmos-sdk/tree/master/proto) directory
of the SDK repo.

For an overview of the SDK's usage of protocol buffers, please look at the following ADRs and meta-issues which tracked
the corresponding work:
- [ADR019 Protobuf State Encoding](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-019-protobuf-state-encoding.md) / [Full Proto Encoding (#5444)](https://github.com/cosmos/cosmos-sdk/issues/5444)
- [ADR020 Protobuf Transaction Encoding](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-020-protobuf-transaction-encoding.md) / [Proto Any Tx Migration (#6213)](https://github.com/cosmos/cosmos-sdk/issues/6213) 
- [ADR021 Protobuf Query Encoding](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-021-protobuf-query-encoding.md) / [Query Protobuf Migration (#5921)](https://github.com/cosmos/cosmos-sdk/issues/5921)
- [ARD031 Protobuf Msg Services](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-031-msg-service.md)
- [ADR023 Protobuf Naming Conventions](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-023-protobuf-naming.md)
- [ADR027 Deterministic Protobuf Serialization](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-027-deterministic-protobuf-serialization.md)

As a high level summary these represent the following major changes to the SDK:
- New protocol buffer based encoding for all blockchain state (direct queries to tendermint now return protobuf binary
  encoded data, as opposed to Amino encoded data)
- New transaction signing path implemented according to ADR020 above
- Two new querier APIs (see [#5921](https://github.com/cosmos/cosmos-sdk/issues/5921) for details)
  - Support for new gRPC based querier services
  - gRPC Gateway for REST querying corresponding to the new gRPC querier services

For details on how to upgrade Cosmos SDK based apps and modules to Stargate, please see
[App and Modules Migration](https://docs.cosmos.network/master/migrations/app_and_modules.html) in the Cosmos SDK docs.

**Note:** Existing Amino REST endpoints are all preserved, though they are planned to be deprecated in a future release.

## Inter Blockchain Communication (IBC)

The `x/ibc` module is now available and ready for use. High level IBC documentation is available at [docs.cosmos.network](https://docs.cosmos.network/master/ibc/overview.html). For more details check the the module documentation in the [`x/ibc/core/spec`](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc/core/spec) directory, or the ICS specs below:
* [ICS 002 - Client Semantics](https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics) subpackage
* [ICS 003 - Connection Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-003-connection-semantics) subpackage
* [ICS 004 - Channel and Packet Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-004-channel-and-packet-semantics) subpackage
* [ICS 005 - Port Allocation](https://github.com/cosmos/ics/blob/master/spec/ics-005-port-allocation) subpackage
* [ICS 006 - Solo Machine Client](https://github.com/cosmos/ics/tree/master/spec/ics-006-solo-machine-client) subpackage
* [ICS 007 - Tendermint Client](https://github.com/cosmos/ics/blob/master/spec/ics-007-tendermint-client) subpackage
* [ICS 009 - Loopback Client](https://github.com/cosmos/ics/tree/master/spec/ics-009-loopback-client) subpackage
* [ICS 020 - Fungible Token Transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer) subpackage
* [ICS 023 - Vector Commitments](https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments) subpackage
* [ICS 024 - Host State Machine Requirements](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements) subpackage

## Single application binary [#6571](https://github.com/cosmos/cosmos-sdk/pull/6571)

Cosmos SDK now compiles to a single application binary, as opposed to seperate binaries for running a node and one for
the CLI & REST server.

We've now included a barebones application `simapp` / `simd` for testing and demonstrating how an SDK application should
be constructed.

Details of the CLI refactor can be found [here](https://github.com/cosmos/cosmos-sdk/issues/6571).

## Test Network Testing Framework [#6489](https://github.com/cosmos/cosmos-sdk/pull/6489)

Introduction of the testutil package. This package allows the creation of an entirely in-process testing cluster with
fully operational Tendermint nodes constructed with SimApp. Each node has an RPC & API exposed. In addition, the network
exposes a Local client that can be used to directly interface with Tendermint's RPC. The test network is entirely
configurable.

## Tendermint 0.34.1 [#6365](https://github.com/cosmos/cosmos-sdk/issues/6365)

Update to the latest version of tendermint which adds support for the following (in addition many other improvements):
- ABCI update to give application control over block pruning
- Support for arbitrary initial block height
- Support for State Sync
- Evidence handling for new types of evidence submitted by Tendermint from light clients

A more detailed list of Tendermint updates can be found [here](https://github.com/tendermint/tendermint/blob/master/CHANGELOG.md#v0340-rc4).