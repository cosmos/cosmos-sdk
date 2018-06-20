# Specifications

This specification describes how to implement the LCD. It explains how to follow the header chain,
update the root of trust and verify proofs against the latest state root. It furthermore describes
how to integrate modules, such as ICS0 (TendermintAPI), ICS1 (KeyAPI), ICS20 (TokenAPI), 
ICS21 (StakingAPI) and ICS22 (GovernanceAPI).


## Introduction

This document specifies the LCD (light client daemon) module of Cosmos-SDK. The initial design is 
mentioned in [Cosmos Whitepaper](https://cosmos.network/resources/whitepaper#inter-blockchain-communication-ibc). 
LCD is a server, which exposes a REST api to query and interact with a fullnode. It verifies all 
data that a fullnode returns against a recent state root and follows the header chain to stay up to 
date on the latest state root. This module allows application developers to easily write their 
develop their own LCD for their own application. 

From a client implementor's perspective, LCD will decrease the difficulty and latency of verifying 
the authenticity of any type of transaction. A typical architecture for using LCD is the following:

![architecture](https://github.com/irisnet/cosmos-sdk/raw/suyu/lcd/docs/spec/lcd/pics/architecture.png)
