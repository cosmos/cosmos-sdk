# Cosmos-SDK lcd (lite client daemon) Design Specifications

## Abstract

This paper specifies the lcd (lite client daemon) module of Cosmos-SDK. This module enables consumers to query and verify transactions and blocks, as well as other abci application states like coin quantity, without deploying a fullnode which requires much computing resource and storage space.

All consumers can deploy their own lcd nodes on their own personal computers, even on their smart phones. Then without trusting any blockchain fullnodes or any single validator node, just trusting their own lcd node and the whole validator set, they can verify all blockchain state. For instance, a cosmos consumer wants to check how many Atom coins he/she has. He/she can send a coin query request to his/her own lcd node, then the lcd node send another query request to a fullnode to get coin quantity and related Merkle proof. Finally the lcd node verify the proof. If the proof is valid, then the coin quantity the consumer has is definitely right.

## Contents

Cosmos-SDK lcd (lite client daemon) acts as a REST-SERVER. It provides a set of APIs which covers key management, tendermint blockchain monitor and other cosmos modules related interfaces.

### key management:
Url:/keys, Method: GET

Url:/keys, Method: POST

Url:/keys/seed, Method: GET

Url:/keys/{name}, Method: GET

Url:/keys/{name}, Method: PUT

Url:/keys/{name}, Method: DELETE

### blockchain:
Url:/node_info, Method: GET

Url:/syncing, Method: GET

Url:/blocks/latest, Method: GET

Url:/blocks/{height}, Method: GET

Url:/validatorsets/latest, Method: GET

Url:/validatorsets/{height}, Method: GET

### transaction:
Url:/txs/{hash}, Method: GET

### auth module:query account info
Url:/accounts/{address}, Method: GET

### bank module:transfer coin
Url:/accounts/{address}/send, Method: POST

### ibc module:
Url:/ibc/{destchain}/{address}/send, Method: POST

### stake module:
Url:/stake/{delegator}/bonding_status/{validator}, Method: GET

Url:/stake/validators, Method: GET

Url:/stake/delegations, Method: POST