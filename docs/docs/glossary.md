---
sidebar_position: 3

---

# Glossary

## ABCI (Application Blockchain Interface)

The interface between the Tendermint consensus engine and the application state machine, allowing them to communicate and perform state transitions. ABCI is a critical component of the Cosmos SDK, enabling developers to build applications using any programming language that can communicate via ABCI.

## ATOM

The native staking token of the Cosmos Hub, used for securing the network, participating in governance, and paying fees for transactions.

## CometBFT

A Byzantine Fault Tolerant (BFT) consensus engine that powers the Cosmos SDK. CometBFT is responsible for handling the consensus and networking layers of a blockchain.

## Cosmos Hub

The first blockchain built with the Cosmos SDK, functioning as a hub for connecting other blockchains in the Cosmos ecosystem through IBC.

## Cosmos SDK

A framework for building blockchain applications, focusing on modularity, scalability, and interoperability.

## CosmWasm

A smart contract engine for the Cosmos SDK that enables developers to write and deploy smart contracts in WebAssembly (Wasm). CosmWasm is designed to be secure, efficient, and easy to use, allowing developers to build complex applications on top of the Cosmos SDK.

## Delegator

A participant in a Proof of Stake network who delegates their tokens to a validator. Delegators share in the rewards and risks associated with the validator's performance in the consensus process.

## Gas

A measure of computational effort required to execute a transaction or smart contract on a blockchain. In the Cosmos ecosystem, gas is used to meter transactions and allocate resources fairly among users. Users must pay a gas fee, usually in the native token, to have their transactions processed by the network.

## Governance

The decision-making process in the Cosmos ecosystem, which allows token holders to propose and vote on network upgrades, parameter changes, and other critical decisions.

## IBC (Inter-Blockchain Communication)

A protocol for secure and reliable communication between heterogeneous blockchains built on the Cosmos SDK. IBC enables the transfer of tokens and data across multiple blockchains.

## Interoperability

The ability of different blockchains and distributed systems to communicate and interact with each other, enabling the seamless transfer of information, tokens, and other digital assets. In the context of Cosmos, the Inter-Blockchain Communication (IBC) protocol is a core technology that enables interoperability between blockchains built with the Cosmos SDK and other compatible blockchains. Interoperability allows for increased collaboration, innovation, and value creation across different blockchain ecosystems.

## Light Clients

Lightweight blockchain clients that verify and process only a small subset of the blockchain data, allowing users to interact with the network without downloading the entire blockchain. ABCI++ aims to enhance the security and performance of light clients by enabling them to efficiently verify state transitions and proofs.

## Module

A self-contained, reusable piece of code that can be used to build blockchain functionality within a Cosmos SDK application. Modules can be developed by the community and shared for others to use.

## Slashing

The process of penalizing validators or delegators by reducing their staked tokens if they behave maliciously or fail to meet the network's performance requirements.

## Staking

The process of locking up tokens as collateral to secure the network, participate in consensus, and earn rewards in a Proof of Stake (PoS) blockchain like Cosmos.

## State Sync

A feature that allows new nodes to quickly synchronize with the current state of the blockchain without downloading and processing all previous blocks. State Sync is particularly useful for nodes that have been offline for an extended period or are joining the network for the first time. ABCI++ aims to improve the efficiency and security of State Sync.

## Validator

A network participant responsible for proposing new blocks, validating transactions, and securing the Cosmos SDK-based blockchain through staking tokens. Validators play a crucial role in maintaining the security and integrity of the network.
