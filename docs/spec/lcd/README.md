# Cosmos-SDK lcd (lite client daemon) Design Specifications

## Abstract

This paper specifies the lcd (lite client daemon) module of Cosmos-SDK. This module enables consumers to query and verify transactions and blocks, as well as other abci application states like coin quantity, without deploying a fullnode which requires much computing resource and storage space.

All consumers can deploy their own lcd nodes on their own personal computers, even on their smart phones. Then without trusting any blockchain fullnodes or any single validator node, just trusting their own lcd node and the whole validator set, they can verify all blockchain state. For instance, a cosmos consumer wants to check how many Atom coins he/she has. He/she can send a coin query request to his/her own lcd node, then the lcd node send another query request to a fullnode to get coin quantity and related Merkle proof. Finally the lcd node verify the proof. If the proof is valid, then the coin quantity the consumer has is definitely right.

## Lcd rest-server interfaces

Cosmos-SDK lcd (lite client daemon) acts as a rest-server. It provides a set of APIs which cover key management, tendermint blockchain monitor and other cosmos modules related interfaces.

1.  **Key management**

    1. **url: /keys, Method: GET**
        ```
          Functionality: Get all keys
          Example parameters:
          Example return:
          [
            {
              "name": "monkey",
              "address": "cosmosaccaddr1fedh326uxqlxs8ph9ej7cf854gz7fd5zlym5pd",
              "pub_key": "cosmosaccpub1zcjduc3q8s8ha96ry4xc5xvjp9tr9w9p0e5lk5y0rpjs5epsfxs4wmf72x3shvus0t"
            },
            {
              "name": "test",
              "address": "cosmosaccaddr1thlqhjqw78zvcy0ua4ldj9gnazqzavyw4eske2",
              "pub_key": "cosmosaccpub1zcjduc3qyx6hlf825jcnj39adpkaxjer95q7yvy25yhfj3dmqy2ctev0rxmse9cuak"
            }
          ]
        ```
    2. **url: /keys, Method: POST**
        ```
          Functionality: Recover your key from seed and persist it with your password protection
          Example parameters:
          {
            "name":"test3",
            "password":"12345678",
            "seed":"electric opera balcony clever square coconut typical orbit wonder initial tragic year ride spread angle abandon"

          }
          Example return: BD607C37147656A507A5A521AA9446EB72B2C907
        ```
    3. **url: /keys/seed, Method: GET**
        ```
          Functionality: Create new seed
          Example parameters:
          Example return:
          crime carpet recycle erase simple prepare moral dentist fee cause pitch trigger when velvet animal abandon
        ```
    4. **url: /keys/{name}, Method: GET**
        ```
          Functionality: Get key information according to the specified key name
          Example parameters:
          Example return:
          {
            "name": "test",
            "address": "cosmosaccaddr1thlqhjqw78zvcy0ua4ldj9gnazqzavyw4eske2",
            "pub_key": "cosmosaccpub1zcjduc3qyx6hlf825jcnj39adpkaxjer95q7yvy25yhfj3dmqy2ctev0rxmse9cuak"
          }
        ```
    5. **url: /keys/{name}, Method: PUT**
        ```
          Functionality: Update key password
          Example parameters:
          {
            "old_password":"12345678",
            "new_password":"123456789"
          }
          Example return:
        ```
    6. **url: /keys/{name}, Method: DELETE**
        ```
          Functionality: Delete key from keystore
          Example parameters:
          {
            "password":"12345678"
          }
          Example return:
        ```

2.  **Blockchain Monitor**

    1. **url: /node_info, Method: GET**
        ```
          Functionality: Get lcd node status
          Example parameters:
          Example return:
          {
              "id": "992e24f5761b37de48536cecff0a0687937049a3",
              "listen_addr": "10.0.2.15:46656",
              "network": "test-chain-F0bln0",
              "version": "0.19.7-dev",
              "channels": "4020212223303800",
              "moniker": "lhy-ubuntu",
              "other": [
                  "amino_version=0.9.9",
                  "p2p_version=0.5.0",
                  "consensus_version=v1/0.2.2",
                  "rpc_version=0.7.0/3",
                  "tx_index=on",
                  "rpc_addr=tcp://0.0.0.0:46657"
              ]
          }
        ```
    2. **url: /syncing, Method: GET**
        ```
          Functionality: Check syncing status of the fullnode which is connecting with the lcd node
          Example parameters:
          Example return: false
        ```
    3. **url: /blocks/latest, Method: GET**
        ```
          Functionality: Get the lasted block and verify it
          Example parameters:
          Example return:
        ```
    4. **url: /blocks/{height}, Method: GET**
        ```
          Functionality: Get the block at specified height and verify it
          Example parameters:
          Example return:
        ```
    5. **url: /validatorsets/latest, Method: GET**
        ```
          Functionality: Get the lasted validatorsets and verify it
          Example parameters:
          Example return:
        ```
    6. **url: /validatorsets/{height}, Method: GET**
        ```
          Functionality: Get the validatorsets at specified block height and verify it
          Example parameters:
          Example return:
        ```

3.  **Transaction**

    1. **url: /txs/{hash}, Method: GET**
        ```
          Functionality: Get the transaction by its hash and verify it
          Example parameters:
          Example return:
        ```

4.  **Auth module**

    1. **url: /accounts/{address}, Method: GET**
        ```
          Functionality: Get the account information and verify the returned proof
          Example parameters:
          Example return:
        ```

5.  **Bank module**

    1. **url: /accounts/{address}/send, Method: POST**
        ```
          Functionality: transfer asset
          Example parameters:
          Example return:
        ```

6.  **Ibc module**

    1. **url: /ibc/{destchain}/{address}/send, Method: POST**
        ```
          Functionality: transfer asset across chain
          Example parameters:
          Example return:
        ```

7.  **Stake module**

    1. **url: /stake/{delegator}/bonding_status/{validator}, Method: GET**
        ```
          Functionality: get delegator information and verify returned proof
          Example parameters:
          Example return:
        ```
    2. **url: /stake/validators, Method: GET**
        ```
          Functionality: get all validators information in stake module and verify returned proof
          Example parameters:
          Example return:
        ```
    3. **url: /stake/delegations, Method: POST**
        ```
          Functionality: send a delegate transaction
          Example parameters:
          Example return:
        ```

8.  **Directly send transactions**

    1. **url: /broadcast_tx_commit, Method: POST**
        ```
          Functionality: Directly send a transaction and wait until on-chain
          Example parameters:
          Example return:
        ```
    2. **url: /broadcast_tx_sync, Method: POST**
        ```
          Functionality: Directly send a transaction and wait until checkTX is done
          Example parameters:
          Example return:
        ```
    3. **url: /broadcast_tx_async, Method: POST**
        ```
          Functionality: Directly send a transaction asynchronous without wait for anything
          Example parameters:
          Example return:
        ```

## Build Merkle Proof

## Verify Merkle Proof