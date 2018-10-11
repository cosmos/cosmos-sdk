# Cosmos Inter-Blockchain Communication (IBC) Protocol

## Abstract

This paper specifies the Cosmos Inter-Blockchain Communication (IBC) protocol. The IBC protocol defines a set of semantics for authenticated, strictly-ordered message passing between two blockchains with independent consensus algorithms.  

The core IBC protocol is payload-agnostic. On top of IBC, developers can implement the semantics of a particular application, enabling users to transfer valuable assets between different blockchains while preserving the contractual guarantees of the asset in question - such as scarcity and fungibility for a currency or global uniqueness for a digital kitty-cat. 

IBC requires two blockchains with cheaply verifiable rapid finality and Merkle tree substate proofs. The protocol makes no assumptions of block confirmation times or maximum network latency of packet transmissions, and the two consensus algorithms remain completely independent. Each chain maintains a local partial order and inter-chain message sequencing ensures cross-chain linearity. Once the two chains have registered a trust relationship, cryptographically verifiable packets can be sent between them.

IBC was first outlined in the [Cosmos Whitepaper](https://github.com/cosmos/cosmos/blob/master/WHITEPAPER.md#inter-blockchain-communication-ibc), and later described in more detail by the [IBC specification paper](https://github.com/cosmos/ibc/blob/master/CosmosIBCSpecification.pdf). This document supersedes both. It explains the requirements and structure of the protocol and provides sufficient detail for both analysis and implementation.

## Contents

1.  **[Overview](overview.md)**
    1.  [Summary](overview.md#11-summary)
    1.  [Definitions](overview.md#12-definitions)
    1.  [Threat Models](overview.md#13-threat-models)
1.  **[Connections](connections.md)**
    1.  [Definitions](connections.md#21-definitions)
    1.  [Requirements](connections.md#22-requirements)
    1.  [Connection lifecycle](connections.md#23-connection-lifecycle)
        1.  [Opening a connection](connections.md#231-opening-a-connection)
        1.  [Following block headers](connections.md#232-following-block-headers)
        1.  [Closing a connection](connections.md#233-closing-a-connection)
1.  **[Channels & Packets](channels-and-packets.md)**
    1.  [Background](channels-and-packets.md#31-background)
    1.  [Definitions](channels-and-packets.md#32-definitions)
        1. [Packet](channels-and-packets.md#321-packet)
        1. [Receipt](channels-and-packets.md#322-receipt)
        1. [Queue](channels-and-packets.md#323-queue)
        1. [Channel](channels-and-packets.md#324-channel)
    1.  [Requirements](channels-and-packets.md#33-requirements)
    1.  [Sending a packet](channels-and-packets.md#34-sending-a-packet)
    1.  [Receiving a packet](channels-and-packets.md#35-receiving-a-packet)
    1.  [Packet relayer](channels-and-packets.md#36-packet-relayer)
1.  **[Optimizations](optimizations.md)**
    1.  [Timeouts](optimizations.md#41-timeouts)
    1.  [Cleanup](optimizations.md#42-cleanup)
1.  **[Conclusion](conclusion.md)**
1.  **[References](references.md)**
1.  **[Appendices](appendices.md)**
    1. [Appendix A: Encoding Libraries](appendices.md#appendix-a-encoding-libraries)
    1. [Appendix B: IBC Queue Format](appendices.md#appendix-b-ibc-queue-format)
    1. [Appendix C: Merkle Proof Format](appendices.md#appendix-c-merkle-proof-formats)
    1. [Appendix D: Byzantine Recovery Strategies](appendices.md#appendix-d-byzantine-recovery-strategies)
    1. [Appendix E: Tendermint Header Proofs](appendices.md#appendix-e-tendermint-header-proofs)
