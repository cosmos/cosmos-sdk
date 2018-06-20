## ABCI Full Node vs ABCI Light Client

A light client has all the security of a full node with minimal bandwidth requirements. It is used as the basis of Cosmos IBC. The minimal bandwidth requirements allows developers to build fully secure, efficient and usable mobile apps.

A full node of ABCI is different from its light client in the following ways:

- Full Node
  - Node discovery 
  - Verify and broadcast valid transactions in mempool
  - Verify and store new blocks
  - If this node is a validator node, it could contribute to protect the safety of network and reach consensus.
  - Resource consuming: huge computing resources for transaction verification and huge storage resources for saving blocks
- Light Client Node
  - Redirect requests to full nodes
  - Verify transaction according to its hash
  - Verify precommit info at specific height
  - Verify block at specific height
  - Verify the proof in abci query result
  - Only need limited computing and storage resources, available for mobiles and personal computers



## Design 

![light-client-architecture](/Users/suyu/Documents/bianjie/exchange/light-client-architecture.png)

Router will redirect the requests. It will directly send some of the requests to Tendermint Endpoint, like `query for validator set `or `broadcast_tx_commit`. Certifier is reponsible for tracking the voting power of validators. `Warpper engine` will collect all the related info about verifying certain information. 

