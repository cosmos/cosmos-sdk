# Cosmos Hub (Gaia) LCD API

This document describes the API that is exposed by the specific LCD implementation of the Cosmos
Hub (Gaia). Those APIs are exposed by a REST server and can easily be accessed over HTTP/WS(websocket)
connections.

The complete API is comprised of the sub-APIs of different modules. The modules in the Cosmos Hub
(Gaia) API are:

- ICS0 (TendermintAPI)
- ICS1 (KeyAPI)
- ICS20 (TokenAPI)
- ICS21 (StakingAPI)
- ICS22 (GovernanceAPI) - not yet implemented

Error messages my change and should be only used for display purposes. Error messages should not be
used for determining the error type.

## ICS0 - TendermintAPI

Exposes the same functionality as the Tendermint RPC from a full node. It aims to have a very
similar API.

### /txs - POST

url: /txs

Functionality: Submit a signed transaction. Can be of type:

- `return=sync`: returns a response from CheckTx
- `return=async`: does not return a response from CheckTx
- `return=block`: waits for for the transaction to be committed in a block

Parameters:

| Parameter   | Type   | Default | Required | Description            |
| ----------- | ------ | ------- | -------- | ---------------------- |
| transaction | string | null    | true     | signed tx bytes        |
| return      | string | null    | true     | broadcast return value |

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
    "error":"Could not submit the transaction.",
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
        "account":[
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



### /keys - POST

url: /keys

Functionality: Create a new key.

Parameter:

| Parameter | Type   | Default | Required | Description     |
| --------- | ------ | ------- | -------- | --------------- |
| name      | string | null    | true     | name of key     |
| password  | string | null    | true     | password of key |

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

| Parameter    | Type   | Default | Required | Description  |
| ------------ | ------ | ------- | -------- | ------------ |
| old_password | string | null    | true     | old password |
| new_password | string | null    | true     | new password |

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

| Parameter | Type   | Default | Required | Description     |
| --------- | ------ | ------- | -------- | --------------- |
| password  | string | null    | true     | password of key |

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

### /keys/{name}/recover - POST

url: /keys/{name}/recover

Functionality: Recover your key from seed and persist it encrypted with the password.

Parameter:

| Parameter | Type   | Default | Required | Description     |
| --------- | ------ | ------- | -------- | --------------- |
| password  | string | null    | true     | password of key |
| seed      | string | null    | true     | seed of key     |

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



### /auth/accounts/{address} - GET

url: /auth/accounts/{address}

Functionality: Query the information of an account .

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result": {
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
        "address": "82A57F8575BDFA22F5164C75361A21D2B0E11089",
        "public_key": "PubKeyEd25519{A0EEEED3C9CE1A6988DEBFE347635834A1C0EBA0B4BB1125896A7072D22E650D}",
        "coins": [
            "atom": 300,
            "photon": 15
        ],
        "account_number": 1,
        "sequence": 7
    }
}
}
```

Returns on error:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not find the account",
    "result":{}
}
```



## ICS20 - TokenAPI

The TokenAPI exposes all functionality needed to query account balances and send transactions.



### /bank/balance/{account} - GET

url: /bank/balance/{account}

Functionality: Query the specified account's balance.

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

### /bank/transfer - POST

url: /bank/transfer

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
    "rest api":"2.1",
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

### /stake/delegators/{delegatorAddr}/validators - GET

url: /stake/delegators/{delegatorAddr}/validators

Functionality: Query all validators that a delegator is bonded to.

Returns on success:

```json
{
    "rest api":"2.1",
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
    "error":"TODO",
    "result":{}
}
```

### /stake/delegators/{delegatorAddr}/validators/{validatorAddr} - GET

url: /stake/delegators/{delegatorAddr}/validators/{validatorAddr}

Functionality: Query a validator that a delegator is bonded to

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
    "error":"TODO",
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
    "error":"TODO",
    "result":{}
}
```



## ICS22 - GovernanceAPI

The GovernanceAPI exposes all functionality needed for casting votes on plain text, software upgrades and parameter change proposals.

::: tip Note
🚧 We are actively working on documentation for the governance module rest endpoints .
:::



## ICS23 - SlashingAPI

The SlashingAPI exposes all functionality needed for to slash validators and delegators in PoS.

### /slashing/validator/{validatorAddr}/signing-info - GET

url: /slashing/validator/{validatorAddr}/signing-info

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

### /slashing/validators/{validatorAddr}/unrevoke - POST

url: /slashing/validators/{validatorAddr}/unrevoke

Functionality: Query the information from a single validator.

Parameter:

| Parameter      | Type   | Default | Required | Description     |
| -------------- | ------ | ------- | -------- | --------------- |
| name           | string | null    | true     | name of the key |
| password       | string | null    | true     | password of key |
| chain_id       | string | null    | false    |                 |
| account_number | int64  |         |          |                 |
| sequence       | int64  |         |          |                 |
| gas            | int64  |         |          |                 |

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
    "error":"TODO",
    "result":{}
}
```
