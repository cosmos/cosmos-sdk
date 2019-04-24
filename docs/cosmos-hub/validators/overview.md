# Validators Overview

## Introduction

The [Cosmos Hub](/introduction/cosmos-hub.md) is based on [Tendermint](/introduction/tendermint.md), which relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes which contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own Atoms and have Atoms ["delegated"](/staking/delegators), or staked, to them by token holders. The Cosmos Hub will have 100 validators, but over time this will increase to 300 validators according to a predefined schedule. The validators are determined by who has the most stake delegated to them — the top 100 validator candidates with the most stake will become Cosmos validators.

Validators and their delegators will earn Atoms as block provisions and tokens as transaction fees through execution of the Tendermint consensus protocol. Initially, transaction fees will be paid in Atoms but in the future, any token in the Cosmos ecosystem will be valid as fee tender if it is whitelisted by governance. Note that validators can set commission on the fees their delegators receive as additional incentive.

If validators double sign, are frequently offline or do not participate in governance, their staked Atoms (including Atoms of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Hardware

There currently exists no appropriate cloud solution for validator key management. This may change in 2018 when cloud SGX becomes more widely available. For this reason, validators must set up a physical operation secured with restricted access. A good starting place, for example, would be co-locating in secure data centers.

Validators should expect to equip their datacenter location with redundant power, connectivity, and storage backups. Expect to have several redundant networking boxes for fiber, firewall and switching and then small servers with redundant hard drive and failover. Hardware can be on the low end of datacenter gear to start out with.

We anticipate that network requirements will be low initially. The current testnet requires minimal resources. Then bandwidth, CPU and memory requirements will rise as the network grows. Large hard drives are recommended for storing years of blockchain history.

## Set Up a Website

Set up a dedicated validator's website and signal your intention to become a validator on our [forum](https://forum.cosmos.network/t/validator-candidates-websites/127/3). This is important since delegators will want to have information about the entity they are delegating their Atoms to.

## Seek Legal Advice

Seek legal advice if you intend to run a Validator.

## Community

Discuss the finer details of being a validator on our community chat and forum:

* [Validator Chat](https://riot.im/app/#/room/#cosmos_validators:matrix.org)
* [Validator Forum](https://forum.cosmos.network/c/validating)
