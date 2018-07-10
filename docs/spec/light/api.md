# Cosmos Hub (Gaia) LCD API

This document describes the API that is exposed by the specific LCD implementation of the Cosmos
Hub (Gaia). Those APIs are exposed by a REST server and can easily be accessed over HTTP/WS(websocket)
connections.

The complete API is comprised of the sub-APIs of different modules. The modules in the Cosmos Hub
(Gaia) API are:
    * ICS19 (KeyAPI)
    * ICS20 (TokenAPI)
    * StakingAPI

## ICS19 - KeyAPI

This API exposes all functionality needed for key creation, signing and management.


### /ICS19/keys - GET

url: /ICS19/keys, Method: GET

Functionality: Get all keys

Parameters: null

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": [
    {
      "name": "moniker",
      "type": "local",
      "address": "cosmosaccaddr1t48m77vw08fqygkz96l3neqdzrnuvh6ansk7ks",
      "pub_key": "cosmosaccpub1addwnpepqvshl3s856ys9pwpnsfwmtk9psyn2ngzflmlvsvang08hpdmatj5cd0xsrr"
    },
    {
      "name": "string",
      "type": "local",
      "address": "cosmosaccaddr1hwq3hvnn57lqg2srgut68yjpt6f6r4arp0y52a",
      "pub_key": "cosmosaccpub1addwnpepq0p5rknmqctehv8sh8ppuw385n686rh6f553lt7dsn50fnw9n7g6xjndnhm"
    }
  ]
}
```

* The above command returns JSON structured like this if fails:

```
{
    "rest api": "2.0",
    "code":500,
    "error message":"no keys available"
}
```


### /ICS19/keys - POST

url: /ICS19/keys, Method: POST

Functionality: Recover your key from seed and persist it with your password protection

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| name      | string | null    | true     | name of keys     |
| password  | string | null    | true     | password of keys |

Parameters: null

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": {
    "name": "test",
    "type": "local",
    "address": "cosmosaccaddr1y4wkh64fhy4uv2myxjg27k7fk72x796q4j6uke",
    "pub_key": "cosmosaccpub1addwnpepqv3p2erx3v6gvp29xm7x48z4nlvulxzpjvgx2yv9ha4v45pjf66f58utf2v",
    "seed": "gun discover trust slam gap fall oven record until found mule sweet armed fine object save disorder churn expire devote twenty winner knee orphan"
  }
}

```

* The above command returns JSON structured like this if fails:

```
{
  "rest api": "2.0",
  "code": 409,
  "error message": "Account with name test already exists."
}

```


### /ICS19/keys/seed - GET

url: /ICS19/keys/seed, Method: GET

Functionality: Create new seed

Parameters: null

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": "useless tiny alone tilt trap wool know sense rapid balance force kite fork scissors face clap cherry mean task hurdle ahead artist engine magic"
}

```

* The above command returns JSON structured like this if fails:

```
{
  "rest api": "2.0",
  "code": 500,
  "error message": "Internal error."
}

```


### /ICS19/keys/get/{name} - GET

url: /ICS19/keys/get/{name}, Method: GET

Functionality: Get key information according to the specified key name

Parameters: null

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": {
    "name": "test",
    "type": "local",
    "address": "cosmosaccaddr1y4wkh64fhy4uv2myxjg27k7fk72x796q4j6uke",
    "pub_key": "cosmosaccpub1addwnpepqv3p2erx3v6gvp29xm7x48z4nlvulxzpjvgx2yv9ha4v45pjf66f58utf2v"
  }
}

```

* The above command returns JSON structured like this if fails:

```
{
  "rest api": "2.0",
  "code": 404,
  "error message": "Key testt not found"
}
```


### /ICS19/keys/{name} - PUT

url: /ICS19/keys/{name}, Method: PUT

Functionality: Update key password

Parameters:

* Path parameters:

| Parameter       | Type   | Default | Required | Description     |
| --------------- | ------ | ------- | -------- | --------------- |
| name            | string | null    | true     | account name    |

* Body parameters:

| Parameter       | Type   | Default | Required | Description     |
| --------------- | ------ | ------- | -------- | --------------- |
| old_password    | string | null    | true     | password before |
| new_password    | string | null    | true     | password before |


* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": "success"
}
```

* The above command returns JSON structured like this if fails:

```
{
  "rest api": "2.0",
  "code": 401,
  "error message": "Ciphertext decryption failed"
}
```


### /ICS19/keys/{name} - DELETE

url: /ICS19/keys/{name}, Method: DELETE

Functionality: Delete key from keystore

Parameters:

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| password  | string | null    | true     | password of keys |

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": "success"
}
```

* The above command returns JSON structured like this if fails:

```
{
  "rest api": "2.0",
  "code": 401,
  "error message": "Ciphertext decryption failed"
}
```


## ICS20 - TokenAPI

The TokenAPI exposes all functionality needed to query account balances and send transactions.

### /ICS20/balance/{account} - GET

url: /ICS20/balance/{account}, Method: GET

Functionality:

Parameters:

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": {
    "address": "5D4FBF798E79D20222C22EBF19E40D10E7C65F5D",
    "coins": [
      {
        "denom": "monikerToken",
        "amount": "1000"
      },
      {
        "denom": "steak",
        "amount": "50"
      }
    ],
    "public_key": null,
    "account_number": 0,
    "sequence": 0
  }
}
```

* The above command returns JSON structured like this if fails:

```
{
  "rest api": "2.0",
  "code": 409,
  "error message": "decoding bech32 failed: checksum failed. Expected 4rjz6l, got nsk7ks."
}
```


### /ICS20/create_transfer - POST

url: /ICS20/create_transfer, Method: **POST**

Functionality: transfer asset

Parameters:

| Parameter | Type   | Default | Required | Description                 |
| --------- | ------ | ------- | -------- | --------------------------- |
| from_address  | string | null    | true     | address from                |
| to_address  | string | null    | true     | address to send to     |
| amount  | int    | null    | true     | amount of the token         |
| denomination  | string | null    | true     | denomination of the token   |
| account_number  | int | null    | false     | account number, user can query the valid account number by accessing previous API: /balance/{account} |
| sequence  | int    | null    | false     | sequence number, once a transaction is send from this account, the sequence number increases one. User can get the valid sequence number by accessing previous API: /balance/{account} |
| ensure_account_sequence  | bool | false    | false     | if true, lcd will query full node and calculate correct value for account_number and sequence  |
| gas  | int | null    | false     | transaction fee   |


* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": "eyJhY2NvdW50X251bWJlciI6IjAiLCJjaGFpbl9pZCI6InRlc3QtY2hhaW4tUnRBUzBLIiwiZmVlIjp7ImFtb3VudCI6W3siYW1vdW50IjoiMCIsImRlbm9tIjoiIn1dLCJnYXMiOiIwIn0sIm1lbW8iOiIiLCJtc2dzIjpbeyJpbnB1dHMiOlt7ImFkZHJlc3MiOiJjb3Ntb3NhY2NhZGRyMXQ0OG03N3Z3MDhmcXlna3o5NmwzbmVxZHpybnV2aDZhbnNrN2tzIiwiY29pbnMiOlt7ImFtb3VudCI6IjAiLCJkZW5vbSI6InN0cmluZyJ9XX1dLCJvdXRwdXRzIjpbeyJhZGRyZXNzIjoiY29zbW9zYWNjYWRkcjF0NDhtNzd2dzA4ZnF5Z2t6OTZsM25lcWR6cm51dmg2YW5zazdrcyIsImNvaW5zIjpbeyJhbW91bnQiOiIwIiwiZGVub20iOiJzdHJpbmcifV19XX1dLCJzZXF1ZW5jZSI6IjAifQ=="
}
```

The result value is the base64 encoding string of transaction bytes. Firstly user should decode the base64 string, then sign the transaction bytes.

* The above command returns JSON structured like this if fails:

```
{
  "rest api": "2.0",
  "code": 400,
  "error message": "decoding bech32 failed: checksum failed. Expected xfzga5, got nsk7ks."
}
```


### /ICS20/signed_transfer - POST

url: /ICS20/signed_transfer, Method: POST

Functionality: transfer asset

Parameters:

| Parameter       | Type   | Default | Required | Description            |
| ------------    | ------ | ------- | -------- | ----------------------------------------------- |
| transaction_data | []byte | null    | true     | bytes of a valid transaction |
| signature_list | [][]byte | null    | true     | bytes of signature list |
| public_key_list | [][]byte | null    | true     | bytes of public key list |

* The above command returns JSON structured like this if success:

```
{
  "rest api": "2.0",
  "code": 0,
  "result": {
    "check_tx": {
      "log": "Msg 0: ",
      "gasUsed": "3315",
      "tags": [
        {
          "key": "c2VuZGVy",
          "value": "MjU1RDZCRUFBOUI5MkJDNjJCNjQzNDkwQUY1QkM5Qjc5NDZGMTc0MA=="
        },
        {
          "key": "cmVjaXBpZW50",
          "value": "QkI4MTFCQjI3M0E3QkUwNDJBMDM0NzE3QTM5MjQxNUU5M0ExRDdBMw=="
        }
      ],
      "fee": {
        "key": ""
      }
    },
    "deliver_tx": {
      "log": "Msg 0: ",
      "gasUsed": "3315",
      "tags": [
        {
          "key": "c2VuZGVy",
          "value": "MjU1RDZCRUFBOUI5MkJDNjJCNjQzNDkwQUY1QkM5Qjc5NDZGMTc0MA=="
        },
        {
          "key": "cmVjaXBpZW50",
          "value": "QkI4MTFCQjI3M0E3QkUwNDJBMDM0NzE3QTM5MjQxNUU5M0ExRDdBMw=="
        }
      ],
      "fee": {}
    },
    "hash": "8F2958D1B1A0C6D03D8E0A983CACF984AE4918C1",
    "height": 138
  }
}
```

* The above command returns JSON structured like this if fails:

```
{
    "rest api": "2.0",
    "error message": "Invalid Signature",
    "code":500
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
