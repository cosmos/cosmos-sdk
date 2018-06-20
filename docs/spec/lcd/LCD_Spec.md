# Cosmos-SDK LCD(light client daemon) Specifications

## Intro

This document specifies the LCD (light client daemon) module of Cosmos-SDK. The initial design is mentioned in [Cosmos Whitepaper](https://cosmos.network/resources/whitepaper#inter-blockchain-communication-ibc). LCD is a server, which exposes a REST api to query and interact with a fullnode. It verifies all data that a fullnode returns against a recent state root and follows the header chain to stay up to date on the latest state root. This module allows application developers to easily write their develop their own LCD for their own application. 

From a client implementor's perspective, LCD will decrease the difficulty and latency of verifying the authenticity of any type of transaction. A typical architecture for using LCD is the following:

![architecture](https://github.com/irisnet/cosmos-sdk/raw/suyu/lcd/docs/spec/lcd/pics/architecture.png)

## Contents

1. **Design overview**
2. **API Document**


