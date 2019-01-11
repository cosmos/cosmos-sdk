## 1 Overview

([Back to table of contents](README.md#contents))

### 1.1 Summary

The IBC protocol creates a mechanism by which two replicated fault-tolerant state machines may pass messages to each other. These messages provide a base layer for the creation of communicating blockchain architecture that overcomes challenges in the scalability and extensibility of computing blockchain environments.

The IBC protocol assumes that multiple applications are running on their own blockchain with their own state and own logic. Communication is achieved over an ordered message queue primitive, allowing the creation of complex inter-chain processes without trusted third parties.

The message packets are not signed by one psuedonymous account, or even multiple, as in multi-signature sidechain implementations. Rather, IBC assigns authorization of the packets to the source blockchain's consensus algorithm, performing light-client style verification on the destination chain. The Byzantine-fault-tolerant properties of the underlying blockchains are preserved: a user transferring assets between two chains using IBC must trust only the consensus algorithms of both chains.

In this paper, we define a process of posting block headers and Merkle tree proofs to enable secure verification of individual packets. We then describe how to combine these packets into a messaging queue to guarantee ordered delivery. We then explain how to handle packet receipts (response/error) on the source chain, which enables the creation of asynchronous RPC-like protocols on top of IBC. Finally, we detail some optimizations and how to handle Byzantine blockchains.

### 1.2 Definitions

*Blockchain* - A replicated fault-tolerant state machine with a distributed consensus algorithm. The smallest unit produced through consensus is a block, which may contain many transactions, each applying some arbitrary mutation to the state.

*Module* - We assume that the state machine of each blockchain is comprised of multiple components that have limited rights to execute some particular set of state transfers (these are modules in the Cosmos SDK or smart contracts in Ethereum).

*Finality* - The guarantee that a given block will not be reverted within some predefined conditions of a consensus algorithm. All proof-of-work systems offer probabilistic finality, which means that the difficulty of reverting a block increases as the block is embedded more deeply in the chain. Many proof-of-stake systems offer much weaker guarantees, based only on the honesty of the block producers. BFT algorithms such as Tendermint guarantee complete finality upon production of a block (unless over two thirds of the validators collude to break consensus, in which case the offenders can be identified and punished - further discussion of that scenario is outside the scope of this document).

*Attributable* - Knowledge of the pseudonymous identity which made a statement, whom we can punish with some deduction of value (slashing) if the statement is false. Synonymous with accountability.

*Unbonding period* - Proof-of-stake algorithms need to lock the stake (prevent transfers) for some time to provide a lower bound for the length of a long-range attack [[3](./references.md#3)]. Complete finality is associated with a subset of the proof-of-stake class of consensus algorithms. We assume the proof-of-stake algorithms utilized by the two blockchains have some unbonding period P.

### 1.3 Threat Models

*False statements* - Any information we receive may be false.

*Network partitions and delays* - We assume an asynchronous, adversarial network with unbounded latency. Network messages may be modified, reordered, duplicated, or selectively dropped. Actors may be arbitrarily partitioned by a powerful adversary. The IBC protocol favors correctness over liveness where applicable.

*Byzantine actors* - An entire blockchain may not act according to protocol. This must be detectable and provable, allowing the communicating blockchain to revoke trust and take necessary action. Application-level protocols designed on top of IBC should consider and mitigate this risk in a manner suitable to their application.
