# Overview

These documents describe the Light Client Daemon, commonly referred to as LCD. The LCD is split into
two separate components. 
The first component is generic for any Tendermint based application. It 
handles the security and connectivity aspects of following the header chain and verifying proofs
from full nodes against locally trusted state roots. Furthermore it exposes exactly the same API as
any Tendermint Core node.
The second component is specific for the Cosmos Hub (Gaiad). It works through the query endpoint on
Tendermint Core and exposes the application specific functionality, which can be arbitrary. All 
queries against the application state have to go through the query endpoint. The advantage of the
query endpoint is that it can verify the proofs that the application returns.

An application developer that wants to build an third party application for the Cosmos Hub (or any
other zone) should build it against it's canonical API. That API is a combination of multiple parts.
All zones have to expose ICS0 (TendermintAPI). Beyond that any zone is free to choose any
combination of module APIs, depending on which modules the state machine uses. The Cosmos Hub will 
initially support ICS0 (TendermintAPI), ICS1 (KeyAPI), ICS20 (TokenAPI), ICS21 (StakingAPI) and 
ICS22 (GovernanceAPI).

All applications are expected to only run against the LCD. The LCD is the only piece of software 
that offers stability guarantees around the zone API.


## What is a Light Client?

A light client has all the security of a full node with minimal bandwidth requirements. The minimal 
bandwidth requirements allows developers to build fully secure, efficient and usable mobile apps,
websites or any other application that does not want to download and follow the full blockchain.


## High-Level Architecture

An application developer that would like to build a third party integration can ship his application
with the LCD for the Cosmos Hub (or any other zone) and only needs to initialise it. Afterwards his
application can interact with the zone as if it was running against a full node.

==============                         ============== 
=     LCD    = ----------------------> = Full Node  =
==============   HTTP/WS Connection    ==============


## Contents

1. [**Specification**](/specification.md)
2. [**API**](/api.md)
3. [**Getting Started**](/getting_started.md)
