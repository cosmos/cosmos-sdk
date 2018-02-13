# IBC Protocol Specification

_v0.4.0 / Feb. 13, 2018_

**Ethan Frey**

## Abstract

This paper specifies the IBC (inter blockchain communication) protocol, which was first described in the Cosmos white paper [[1](./footnotes.md#1)] in June 2016. The IBC protocol uses authenticated message passing to simultaneously solve two problems: transferring value (and state) between two distinct chains, as well as sharding one chain securely. IBC follows the message-passing paradigm and assumes the participating chains are independent.

Each chain maintains a local partial order, while inter-chain messages track any cross-chain causality relations. Once two chains have registered a trust relationship, cryptographically provable packets can be securely sent between the chains, using Tendermint's instant finality for quick and efficient transmission.

We currently use this protocol for secure value transfer in the Cosmos Hub, but the protocol can support arbitrary application logic. Details of how Cosmos Hub uses IBC to securely route and transfer value are provided in a separate paper, along with a framework for expressing global invariants. Designing secure communication logic for other types of applications is still an area of research.

The protocol makes no assumptions of block times or network delays in the transmission of the packets between chains and requires cryptographic proofs for every message, and thus is highly robust in a heterogeneous environment with Byzantine actors. This paper explains the requirements and structure of the Cosmos IBC protocol. It aims to provide enough detail to fully understand and analyze the security of the protocol.

## Contents

1.  **[Overview](overview.md)**
    1.  Definitions
    1.  Threat Models
1.  **[Proofs](proofs.md)**
    1.  Establishing a Root of Trust
    1.  Following Block Headers
1.  **[Messaging Queue](queues.md)**
    1.  Merkle Proofs for Queues
    1.  Naming Queues
    1.  Message Contents
    1.  Sending a Packet
    1.  Receipts
    1.  Relay Process
1.  **[Optimizations](optimizations.md)**
    1.  Cleanup
    1.  Timeout
    1.  Handling Byzantine Failures
1.  **[Conclusion](conclusion.md)**

**[Appendix A: Encoding Libraries](appendix-a.md)**

**[Appendix B: IBC Queue Format](appendix-b.md)**

**[Appendix C: Merkle Proof Format](appendix-c.md)**

**[Appendix D: Universal IBC Packets](appendix-d.md)**

**[Appendix E: Tendermint Header Proofs](appendix-e.md)**

