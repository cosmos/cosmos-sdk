## 1 Overview

([Back to table of contents](specification.md#contents))

The IBC protocol creates a mechanism by which multiple sovereign replicated fault tolerant state machines my pass messages to each other. These messages provide a base layer for the creation of communicating blockchain architecture that overcomes challenges in the scalability and extensibility of computing blockchain environments.

The IBC protocol assumes that multiple applications are running on their own blockchain with their own state and own logic. Communication is achieved over an extremely secure message queue protocol, allowing the creation of complex inter-chain processes without trusted parties. This architecture can be seen as a parallel to microservices in the blockchain space, and the IBC protocol can be seen as an analog to the AMQP messaging protocol[[2](./footnotes.md#2)], used by StormMQ, RabbitMQ, etc.

The message packets are not signed by one psuedonymous account, or even multiple. Rather, IBC effectively assigns authorization of the packets to the blockchain's consensus algorithm itself. Not only are blockchains highly secure, they are auditable and have an extremely high creation cost in comparison to cryptographic key pairs. This prevents Sybil attacks and allows out-of-protocol accountability, since any byzantine behavior is provable and can be published to damage the reputation/value of the other blockchain. By using registered blockchains as "actors" in the system, we can achieve extremely high security through a combination of cryptography and incentives.

In this paper, we define a process of posting block headers and merkle proofs to enable secure verification of individual packets. We then describe how to combine these packets into a messaging queue to guarantee reliable, in-order delivery of message. We then explain how to securely handle receipts (response/error), which enables the creation of asynchronous RPC-like protocols. Finally, we detail some optimizations and how to handle byzantine blockchains.

### 1.1   Definitions

_Blockchain_ - an immutable ledger created through distributed consensus, coupled with a deterministic state machine to process the transactions on the ledger. The smallest unit produced through consensus is a block, which may contain many transactions.

_Module_ - we assume that the state machine of the blockchain is comprised of multiple components (modules or smart contracts) that have limited rights, and they can only interact over pre-defined interfaces rather than directly mutating internal state.

_Finality_ - a guarantee that a given block will not be reverted within some predefined conditions. All proof of work systems offer probabilistic finality, which means the probability of that a block will be reverted approaches 0. A "better", alternative chain could exist, but the cost of creation increases rapidly over time. Many "proof of stake" systems offer much weaker guarantees, based only on the honesty of the miners. However, BFT algorithms such as Tendermint guarantee complete finality upon production of a block, unless over two thirds of the validators collude to break consensus. This collusion is provable and can be punished.

_Knowledge_ - what is certain to be true.

_Provable_ - the existence of irrefutable mathematical (often cryptographic) proof of the truth of a given statement. These can be expressed as: given knowledge **A** and a statement **s**, then **B** must be true. This is a form of deductive proof and they can be chained together without losing validity.

_Attributable_ - provable knowledge of who made a statement. If a statement is provably false, then it is known which actor lied. Attributable statements allow us to build incentives against lying, which help enforce finality. This is also referred to as accountability.

_Root of Trust_ - any proof depends on some prior assumptions, however simple they are. We refer to the first assumption we make as the root of trust, and all our knowledge of the system is derived from this root through a provable chain of information. We seek to make this root of trust as simple and a verifiable as possible, since if the original assignment of trust is false, all conclusions drawn will also be false.

_Unbonding Period_ - Proof of Stake algorithms need to freeze the stake for some time to provide a lower bound for the length of a long-range attack [[3](./footnotes.md#3)]. Since complete finality is associated with a subset of the Proof of Stake class of consensus algorithms, I will assume all implementations that support IBC have some unbonding period P, such that if my last knowledge of the blockchain is older than P, I can no longer trust any message without a new root of trust.

The IBC protocol requires each actor to be a blockchain with complete finality. All transitions must be provable and attributable to (at least) one actor. That implies the smallest unit of trust is the consensus algorithm of a blockchain.

### 1.2   Threat Models

_False statements_ - any information we receive may be false, all actors must have enough knowledge be able to prove its correctness without external dependencies. All statements should be attributable.

_Network partitions and delays_ - we assume an asynchronous, adversarial network. Any message may or may not reach the destination. They may be modified or selectively dropped. Messages may reach the destination out of order and may arrive multiple times. There is no upper limit to the time it takes for a message to be received. Actors may be arbitrarily partitioned by a powerful adversary. The protocol favors correctness over liveness. That is, it only acts upon information that is provably correct.

_Byzantine actors_ - it is possible that an entire blockchain is not acting according to protocol. This must be detectable and provable, allowing the communicating blockchain to revoke trust and take necessary action. Furthermore, we should design application-level protocols on top of IBC to minimize risk exposure in the face of Byzantine actors.
