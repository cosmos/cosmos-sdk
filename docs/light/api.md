# Cosmos Hub (Gaia) LCD API

This document describes the API that is exposed by the specific LCD implementation of the Cosmos
Hub (Gaia). Those APIs are exposed by a REST server and can easily be accessed over HTTP/WS(websocket)
connections.

The complete API is comprised of the sub-APIs of different modules. The modules in the Cosmos Hub
(Gaia) API are:

* ICS0 (TendermintAPI)
* ICS1 (KeyAPI)
* ICS20 (TokenAPI)
* ICS21 (StakingAPI) - not yet implemented
* ICS22 (GovernanceAPI) - not yet implemented

Error messages my change and should be only used for display purposes. Error messages should not be
used for determining the error type.

## ICS0 - TendermintAPI - not yet implemented

Exposes the same functionality as the Tendermint RPC from a full node. It aims to have a very
similar API.

### /broadcast_tx_sync - POST

url: /broadcast_tx_sync

Functionality: Submit a signed transaction synchronously. This returns a response from CheckTx.

Parameters:

| Parameter   | Type   | Default | Required | Description     |
| ----------- | ------ | ------- | -------- | --------------- |
| transaction | string | null    | true     | signed tx bytes |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
	"code":0,
	"hash":"0D33F2F03A5234F38706E43004489E061AC40A2E",
	"data":"",
	"log":""
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not submit the transaction synchronously.",
    "result":{}
}
```

### /broadcast_tx_async - POST

url: /broadcast_tx_async

Functionality: Submit a signed transaction asynchronously. This does not return a response from CheckTx.

Parameters:

| Parameter   | Type   | Default | Required | Description     |
| ----------- | ------ | ------- | -------- | --------------- |
| transaction | string | null    | true     | signed tx bytes |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result": {
		"code":0,
		"hash":"E39AAB7A537ABAA237831742DCE1117F187C3C52",
		"data":"",
		"log":""
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not submit the transaction asynchronously.",
    "result":{}
}
```

### /broadcast_tx_commit - POST

url: /broadcast_tx_commit

Functionality: Submit a signed transaction and waits for it to be committed in a block.

Parameters:

| Parameter   | Type   | Default | Required | Description     |
| ----------- | ------ | ------- | -------- | --------------- |
| transaction | string | null    | true     | signed tx bytes |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
        "height":26682,
        "hash":"75CA0F856A4DA078FC4911580360E70CEFB2EBEE",
        "deliver_tx":{
            "log":"",
            "data":"",
            "code":0
        },
        "check_tx":{
        "log":"",
        "data":"",
        "code":0
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not commit the transaction.",
    "result":{}
}
```

## ICS1 - KeyAPI

This API exposes all functionality needed for key creation, signing and management.

### /keys - GET

url: /keys

Functionality: Gets a list of all the keys.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
        "keys":[
           {
                "name":"monkey",
                "address":"cosmosaccaddr1fedh326uxqlxs8ph9ej7cf854gz7fd5zlym5pd",
                "pub_key":"cosmosaccpub1zcjduc3q8s8ha96ry4xc5xvjp9tr9w9p0e5lk5y0rpjs5epsfxs4wmf72x3shvus0t"
            },
            {
                "name":"test",
                "address":"cosmosaccaddr1thlqhjqw78zvcy0ua4ldj9gnazqzavyw4eske2",
                "pub_key":"cosmosaccpub1zcjduc3qyx6hlf825jcnj39adpkaxjer95q7yvy25yhfj3dmqy2ctev0rxmse9cuak"
            }
        ],
        "block_height":5241
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not retrieve the keys.",
    "result":{}
}
```

### /keys/recover - POST

url: /keys/recover

Functionality: Recover your key from seed and persist it encrypted with the password.

Parameter:

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| name      | string | null    | true     | name of key      |
| password  | string | null    | true     | password of key  |
| seed      | string | null    | true     | seed of key      |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
        "address":"BD607C37147656A507A5A521AA9446EB72B2C907"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not recover the key.",
    "result":{}
}
```

### /keys/create - POST

url: /keys/create

Functionality: Create a new key.

Parameter:

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| name      | string | null    | true     | name of key      |
| password  | string | null    | true     | password of key  |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
        "seed":"crime carpet recycle erase simple prepare moral dentist fee cause pitch trigger when velvet animal abandon"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create new key.",
    "result":{}
}
```

### /keys/{name} - GET

url: /keys/{name}

Functionality: Get the information for the specified key.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
        "name":"test",
            "address":"cosmosaccaddr1thlqhjqw78zvcy0ua4ldj9gnazqzavyw4eske2",
            "pub_key":"cosmosaccpub1zcjduc3qyx6hlf825jcnj39adpkaxjer95q7yvy25yhfj3dmqy2ctev0rxmse9cuak"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not find information on the specified key.",
    "result":{}
}
```

### /keys/{name} - PUT

url: /keys/{name}

Functionality: Change the encryption password for the specified key.

Parameters:

| Parameter       | Type   | Default | Required | Description     |
| --------------- | ------ | ------- | -------- | --------------- |
| old_password    | string | null    | true     | old password    |
| new_password    | string | null    | true     | new password    |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{}
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not update the specified key.",
    "result":{}
}
```

### /keys/{name} - DELETE

url: /keys/{name}

Functionality: Delete the specified key.

Parameters:

| Parameter | Type   | Default | Required | Description      |
| --------- | ------ | ------- | -------- | ---------------- |
| password  | string | null    | true     | password of key  |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{}
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not delete the specified key.",
    "result":{}
}
```

## ICS20 - TokenAPI

The TokenAPI exposes all functionality needed to query account balances and send transactions.

### /bank/balance/{account} - GET

url: /bank/balance/{account}

Functionality: Query the specified account.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result": {
        "atom":1000,
        "photon":500,
        "ether":20
    }
}
```

Returns on error:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not find any balance for the specified account.",
    "result":{}
}
```

### /bank/create_transfer - POST

url: /bank/create_transfer

Functionality: Create a transfer in the bank module.

Parameters:

| Parameter    | Type   | Default | Required | Description               |
| ------------ | ------ | ------- | -------- | ------------------------- |
| sender       | string | null    | true     | Address of sender         |
| receiver     | string | null    | true     | address of receiver       |
| chain_id     | string | null    | true     | chain id                  |
| amount       | int    | null    | true     | amount of the token       |
| denomonation | string | null    | true     | denomonation of the token |

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO:<JSON sign bytes for the transaction>"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```

## ICS21 - StakingAPI

The StakingAPI exposes all functionality needed for validation and delegation in Proof-of-Stake.

### /stake/delegators/{delegatorAddr} - GET

url: /stake/delegators/{delegatorAddr}

Functionality: Get all delegations (delegation, undelegation) from a delegator.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result": {
        "atom":1000,
        "photon":500,
        "ether":20
    }
}
```

Returns on error:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not find any balance for the specified account.",
    "result":{}
}
```

### /stake/delegators/{delegatorAddr}/txs - GET

url: /stake/delegators/{delegatorAddr}/txs

Functionality: Get all staking txs (i.e msgs) from a delegator.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```

### /stake/delegators/{delegatorAddr}/delegations - POST

url: /stake/delegators/{delegatorAddr}/delegations

Functionality: Submit a delegation.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```

### /stake/delegators/{delegatorAddr}/delegations/{validatorAddr} - GET

url: /stake/delegators/{delegatorAddr}/delegations/{validatorAddr}

Functionality: Query the current delegation status between a delegator and a validator.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```

### /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} - GET

url: /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}

Functionality: Query all unbonding delegations between a delegator and a validator.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```

### /stake/validators - GET

url: /stake/validators

Functionality: Get all validator candidates.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```

### /stake/validators/{validatorAddr} - GET

url: /stake/validators/{validatorAddr}

Functionality: Query the information from a single validator.

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

Returns on failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```
