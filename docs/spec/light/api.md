# Cosmos Hub (Gaia) LCD API

This document describes the API that is exposed by the specific LCD implementation of the Cosmos
Hub (Gaia). Those APIs are exposed by a REST server and can easily be accessed over HTTP/WS(websocket)
connections.

The complete API is comprised of the sub-APIs of different modules. The modules in the Cosmos Hub
(Gaia) API are:

* ICS0 (TendermintAPI) - not yet implemented
* ICS1 (KeyAPI)
* ICS20 (TokenAPI)
* ICS21 (StakingAPI) - not yet implemented
* ICS22 (GovernanceAPI) - not yet implemented

## ICS0 - TendermintAPI

Exposes the same functionality as the Tendermint RPC from a full node. It aims to have a very
similar API.

## ICS1 - KeyAPI

This API exposes all functionality needed for key creation, signing and management.

### /keys - GET

url: /keys

Functionality: Get all keys

Parameters: null

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
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
"rest api": "2.0",
"code":500,
"error":"no keys available",
"result":{}
}
```


### /keys - POST

url: /keys, Method: POST

Functionality: Recover your key from seed and persist it with your password protection

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| name      | string | null    | true     | name of keys     |
| password  | string | null    | true     | password of keys |
| seed      | string | null    | true     | seed of keys     |

Parameters: null

* The above command returns JSON structured like this if success:

```
{
    "error": "",
    "code":200,
    "result": {
    	"address":BD607C37147656A507A5A521AA9446EB72B2C907
    },
    "rest api": "2.0"
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


### /keys/seed - GET

url: /keys/seed, Method: GET

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


### /keys/{name} - GET

url: /keys/{name}, Method: GET

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


### /keys/{name} - PUT

url: /keys/{name}, Method: PUT

Functionality: Update key password

Parameters:

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


### /keys/{name} - DELETE

url: /keys/{name}, Method: DELETE

Functionality: Delete key from keystore

Parameters:

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


## ICS20 - TokenAPI

The TokenAPI exposes all functionality needed to query account balances and send transactions.

### /balance/{account} - GET

url: /balance/{account}, Method: GET

Functionality:

Parameters:

* The above command returns JSON structured like this if success:

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

* The above command returns JSON structured like this if fails:

```
{
    "error": "Invalid account",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```


### /create_transfer - POST

url: /create_transfer, Method: **POST**

Functionality: transfer asset

Parameters:

| Parameter | Type   | Default | Required | Description                 |
| --------- | ------ | ------- | -------- | --------------------------- |
| from_address  | string | null    | true     | address from                |
| from_chain_id  | string | null    | true     | chain from                |
| to_address  | string | null    | true     | address to send to     |
| to_chain_id  | string | null    | true     | chain to send to     |
| amount  | int    | null    | true     | amount of the token         |
| denomonation  | string | null    | true     | denomonation of the token   |


* The above command returns JSON structured like this if success:

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

* The above command returns JSON structured like this if fails:

```
{
    "error": "Insufficient Funds",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```


### /signed_transfer - POST

url: /signed_transfer, Method: POST

Functionality: transfer asset

Parameters:

| Parameter       | Type   | Default | Required | Description            |
| ------------    | ------ | ------- | -------- | ----------------------------------------------- |
| signed_transfer | []byte | null    | true     | bytes of a valid transaction and it's signature |

* The above command returns JSON structured like this if success:

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

* The above command returns JSON structured like this if fails:

```
{
    "error": "Invalid Signature",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
## StakingAPI

This API exposes all functionality needed for staking info query.

### /stake/validators - GET

url: /stake/validators, Method: GET

Functionality: Get all validators' detailed information

Parameters: null

* The above command returns JSON structured like this if success:

```
{
    "error": "cannot verify the latest block",
    "code": 500,
    "result": [
        {
            "owner": "cosmosvaladdr1fedh326uxqlxs8ph9ej7cf854gz7fd5ze4wr05",
            "pub_key": "cosmosvalpub1zcjduc3qpp0k3kaxnk8pn5syrcltx5ndx6gml7kuxz62zvh87ga42f3tak0sglfqx5",
            "revoked": false,
            "pool_shares": {
                "status": 2,
                "amount": "100"
            },
            "delegator_shares": "0",
            "description": {
                "moniker": "monkey",
                "identity": "",
                "website": "",
                "details": ""
            },
            "bond_height": 0,
            "bond_intra_tx_counter": 0,
            "proposer_reward_pool": null,
            "commission": "0",
            "commission_max": "0",
            "commission_change_rate": "0",
            "commission_change_today": "0",
            "prev_bonded_shares": "0"
        }
    ],
    "rest api": "2.0"
}
```

* The above command returns JSON structured like this if fails:

```
{
    "error": "Encountered internal error",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```

### /stake/{delegator}/bonding_status/{validator} - GET

url: /stake/{delegator}/bonding_status/{validator}, Method: GET

Functionality: Get and verify the delegation informantion

Parameters:
```
delegator: cosmosaccaddr1fedh326uxqlxs8ph9ej7cf854gz7fd5zlym5pd
validator: cosmosvaladdr1fedh326uxqlxs8ph9ej7cf854gz7fd5ze4wr05
```
* The above command returns JSON structured like this if success:

```
{
    "error": "",
    "code":200,
    "result": {
    },
    "rest api": "2.0"
}
```

* The above command returns JSON structured like this if fails:

```
{
    "error": "invalid bech32 prefix. Expected cosmosaccaddr, Got cosmosvaladdr",
    "code":500,
    "result": {},
    "rest api": "2.0"
}
```
