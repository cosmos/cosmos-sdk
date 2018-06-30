## LCD Rest Interfaces

Cosmos-SDK LCD acts as a rest-server. It provides a set of APIs which cover key management, tendermint blockchain monitor and other cosmos modules related interfaces.

1. **Key Management**

1.1  url: /keys, Method: GET
Parameters: null
Functionality: Get all keys
* The above command returns JSON structured like this if success:
```
{
  "jsonrpc": "2.0",
  "code":200,
  "error": "",
  "result": {
        "keys": [
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
	],
    "block_height": 5241
    }   
}
```
* The above command returns JSON structured like this if fails:
```
{
"jsonrpc": "2.0",
"code":500,
"error":"no keys available",
"result":{}
}
```
1.2  url: /keys, Method: POST

Parameters: null

Functionality: Recover your key from seed and persist it with your password protection

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| name      | string | null    | true     | name of keys     |
| password  | string | null    | true     | password of keys |
| seed      | string | null    | true     | seed of keys     |

* The above command returns JSON structured like this if success:

```
{
    "error": "",
    "code":200,
    "result": {
    	"address":BD607C37147656A507A5A521AA9446EB72B2C907
    },
    "jsonrpc": "2.0"
}
        
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "invalid inputs",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
        
```
1.3 url: /keys/seed, Method: **GET**
Functionality: Create new seed
Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
    	"seed":crime carpet recycle erase simple prepare moral dentist fee cause pitch trigger when velvet animal abandon
    },
    "rest api": "2.0"
}

```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot generate new seed",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
        
```
1.4  url: /keys/{name}, Method: GET
Functionality: Get key information according to the specified key name
Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
    	"name": "test",
          "address": "cosmosaccaddr1thlqhjqw78zvcy0ua4ldj9gnazqzavyw4eske2",
          "pub_key": "cosmosaccpub1zcjduc3qyx6hlf825jcnj39adpkaxjer95q7yvy25yhfj3dmqy2ctev0rxmse9cuak"
    },
    "rest api": "2.0"
}

```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot find corresponding name",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
1.5 url: /keys/{name}, Method: **PUT**

Functionality: Update key password

| Parameter       | Type   | Default | Required | Description     |
| --------------- | ------ | ------- | -------- | --------------- |
| old_password    | string | null    | true     | password before |
| new_password    | string | null    | true     | password before |
| repeat_password | string | null    | true     | password before |
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "updated":name
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot update the corresponding key",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
1.6 url: /keys/{name}, Method: **DELETE**

Functionality: Delete key from keystore

Parameters: null

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| password  | string | null    | true     | password of keys |
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "deleted":name
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot delete the corresponding key",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

2. **Blockchain Monitor**

2.1  url: /node_info, Method: **GET**
Functionality: Get LCD node status
Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
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
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot get the node info",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```


2.2  url: /syncing, Method: **GET**
Functionality: Check syncing status of the fullnode which is connecting with the LCD node
Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "syncing":false
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot get the syncing info",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

2.3 url: /blocks/latest, Method: **GET**
Functionality: Get the lasted block and verify it
Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "authentic":true
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot verify the latest block",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

2.4  url: /blocks/{height}, Method: **GET**
Functionality: Get the block at specified height and verify it

Parameters: null

* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "authentic":true
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot verify the block at this height",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
2.5 url: /validatorsets/latest, Method: **GET**
Functionality: Get the lasted validatorsets and verify it

Parameters: null

* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "authentic":true
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot verify the latest block",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
2.6  url: /validatorsets/{height}, Method: **GET**
Functionality: Get the validatorsets at specified block height and verify it

Parameters: null

* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "authentic":true
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot verify the validatorset at this height",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
3. **Transaction**

3.1  url: /txs/{hash}, Method: ** GET**
 Functionality: Get the transaction by its hash and verify it

Parameters: null

* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "authentic":true
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot verify the tx hash",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

3.2  url: /broadcast_tx_commit, Method: **POST**
Functionality: Directly send a transaction and wait until on-chain

Parameters: null

* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "status":sent
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot send the tx to the blockchain",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
3.3  url: /broadcast_tx_sync, Method: **POST**
Functionality: Directly send a transaction and wait until checkTX is done

Parameters: null

* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "status":sent
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot send the tx to the blockchain",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
3.4 url: /broadcast_tx_async, Method: **POST**
  Functionality: Directly send a transaction asynchronous without wait for anything

  Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "status":sent
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot send the tx to the blockchain",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

4. Token API

4.1  url: /balance/{account}, Method: **GET**

- The above command returns JSON structured like this if success:

```
{
    "error": "",
    "code":200,
    "result": {
     {
         "atom": 1000,
         "photon": 500,
         "ether": 20
     }
    },
    "rest api": "2.0"
}
```

- The above command returns JSON structured like this if fails:

```
{
    "error": "Invalid account",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

4.2  url: /create_transfer, Method: **POST**

Functionality: transfer asset


| Parameter | Type   | Default | Required | Description                 |
| --------- | ------ | ------- | -------- | --------------------------- |
| from  | string | null    | true     | address from                |
| to  | string | null    | true     | address want to send to     |
| amount  | int    | null    | true     | amount of the token         |
| denomonation  | string | null    | true     | denomonation of the token   |


- The above command returns JSON structured like this if success:

```
{
    "error": "",
    "code":200,
    "result": {
     ""transaction": "[]bytes of a valid transaction bytes to be signed for that zone"
    },
    "rest api": "2.0"
}
```

- The above command returns JSON structured like this if fails:

```
{
    "error": "Insufficient Funds",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

4.3  url: /signed_transfer, Method: **POST**

Functionality: transfer asset


| Parameter       | Type   | Default | Required | Description            |
| ------------    | ------ | ------- | -------- | ----------------------------------------------- |
| signed_transfer | []byte | null    | true     | bytes of a valid transaction and it's signature |


- The above command returns JSON structured like this if success:

```
{
    "error": "",
    "code":200,
    "result": {
     "tx_hash": ""
    },
    "rest api": "2.0"
}
```

- The above command returns JSON structured like this if fails:

```
{
    "error": "Invalid Signature",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```


5. **Stake module**

5.1 url: /stake/{delegator}/bonding_status/{validator}, Method: **GET**
Functionality: get delegator information and verify returned proof
Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "validator":address1
     "delegator":address2
     "liability": 1
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot send asset to the account",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

5.2 url: /stake/validators, Method: **GET**
Functionality: get all validators information in stake module and verify returned proof
Parameters: null
* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
      "validatorset": [
          {
            "name": "monkey",
            "address": "cosmosaccaddr1fedh326uxqlxs8ph9ej7cf854gz7fd5zlym5pd",
            "pub_key": "cosmosaccpub1zcjduc3q8s8ha96ry4xc5xvjp9tr9w9p0e5lk5y0rpjs5epsfxs4wmf72x3shvus0t",
            power:1000
          },
   		 {
            "name": "test",
            "address": "cosmosaccaddr1thlqhjqw78zvcy0ua4ldj9gnazqzavyw4eske2",
            "pub_key": "cosmosaccpub1zcjduc3qyx6hlf825jcnj39adpkaxjer95q7yvy25yhfj3dmqy2ctev0rxmse9cuak",
             power:1000
         }
	],
    "block_height": 5241
    }   
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot send asset to the account",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

5.3. url: /stake/delegations, Method: **POST**

Functionality: send a delegate transaction

| Parameter    | Type   | Default | Required | Description               |
| ------------ | ------ | ------- | -------- | ------------------------- |
| from         | string | null    | true     | address from              |
| validator    | string | null    | true     | address want to send to   |
| amount       | int    | null    | true     | amount of the token       |
| denomonation | string | null    | true     | denomonation of the token |
| password     | string | null    | true     | passsword of from address |

* The above command returns JSON structured like this if success:
```
{
    "error": "",
    "code":200,
    "result": {
     "sent": success
    },
    "rest api": "2.0"
}
```
* The above command returns JSON structured like this if fails:
```
{
    "error": "cannot get all the delegation",
    "code":500,
    "result": {},
    "rest api": "2.0"
}

```

## Configuration

To start a rest server, we need to specify the following parameters:

| Parameter | Type   | Default | Required | Description                          |
| --------- | ------ | ------- | -------- | ------------------------------------ |
| Chain-id  | string | null    | true     | chain id of the full node to connect |
| node      | URL | null    | true     | address of the full node to connect  |
| laddr      | URL | null    | true     | address to run the rest server on  |
| trust-store      | DIRECTORY | null    | true     | directory for save checkpoints and validator sets |

