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

### POST /txs

url: /txs

Query Parameters:

- `?return={sync|async|block}`:
  * `return=sync`: Waits for the transaction to pass `CheckTx`
  * `return=async`: Returns the request immediately after it is received by the server
  * `return=block`: waits for for the transaction to be committed in a block

POST Body:

```js
{
  "transaction": "string",
  "return": "string",
}
```

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

### GET /keys

url: `/keys`

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



### POST /keys

url: `/keys`

Functionality: Create a new key.

POST Body:

```js
{
  "name": "string",
  "password": "string",
  "seed": "string",
}
```

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



### GET /keys/{name}

url: `/keys/{name}`

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

### PUT /keys/{name}

url: `/keys/{name}`

Functionality: Change the encryption password for the specified key.

PUT Body:

```js
{
  "old_password": "string",
  "new_password": "string",
}
```

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

### DELETE /keys/{name}

url: /keys/{name}

Functionality: Delete the specified key.

DELETE Body:

```js
{
  "password": "string",
}
```

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

### POST /keys/{name}/recover

url: `/keys/{name}/recover`

Functionality: Recover your key from seed and persist it encrypted with the password.

POST Body:

```js
{
  "password": "string",
  "seed": "string",
}
```

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

url: `/auth/accounts/{address}`

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

### GET /bank/balance/{account}

url: `/bank/balance/{account}`

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

### POST /bank/transfers

url: `/bank/transfers`

Functionality: Create a transfer in the bank module.

POST Body:

```js
{
  "amount": [
    {
      "denom": "string",
      "amount": 64,
    }
  ],
  "name": "string",
  "password": "string",
  "chain_id": "string",
  "account_number": 64,
  "sequence": 64,
  "gas": 64,
}
```

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

### GET /stake/delegators/{delegatorAddr}

url: `/stake/delegators/{delegatorAddr}`

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

### GET /stake/delegators/{delegatorAddr}/validators

url: `/stake/delegators/{delegatorAddr}/validators`

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

### GET /stake/delegators/{delegatorAddr}/validators/{validatorAddr}

url: `/stake/delegators/{delegatorAddr}/validators/{validatorAddr}`

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

### GET /stake/delegators/{delegatorAddr}/txs

url: `/stake/delegators/{delegatorAddr}/txs`

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

### POST /stake/delegators/{delegatorAddr}/delegations

url: `/stake/delegators/{delegatorAddr}/delegations`

Functionality: Submit or edit a delegation.

> NOTE: Should this be a PUT instead of a POST? the code indicates that this is an edit operation

POST Body:

```js
{
  "name": "string",
  "password": "string",
  "chain_id": "string",
  "account_number": 64,
  "sequence": 64,
  "gas": 64,
  "delegations": [
    {
      "delegator_addr": "string",
      "validator_addr": "string",
      "delegation": {
        "denom": "string",
        "amount": 1234
      }
    }
  ],
  "begin_unbondings": [
    {
      "delegator_addr": "string",
      "validator_addr": "string",
      "shares": "string",
    }
  ],
  "complete_unbondings": [
    {
      "delegator_addr": "string",
      "validator_addr": "string",
    }
  ],
  "begin_redelegates" [
    {
      "delegator_addr": "string",
      "validator_src_addr": "string",
      "validator_dst_addr": "string",
      "shares": "string",
    }
  ]
  "complete_redelegates": [
    {
      "delegator_addr": "string",
      "validator_src_addr": "string",
      "validator_dst_addr": "string",
    }
  ]
}

```

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

### GET /stake/delegators/{delegatorAddr}/delegations/{validatorAddr}

url: `/stake/delegators/{delegatorAddr}/delegations/{validatorAddr}`

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

### GET /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}

url: `/stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}`

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

### GET /stake/validators

url: `/stake/validators`

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

### GET /stake/validators/{validatorAddr}

url: `/stake/validators/{validatorAddr}`

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

### GET /gov/proposals

url: `/gov/proposals`

Functionality: Query all submited proposals

Response on Success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "proposals":[
         "TODO"
      ]
    }
}
```

Response on Failure:

```json
{
    "rest api":"2.0",
    "code":500,
    "error":"Could not create the transaction.",
    "result":{}
}
```

### POST /gov/proposals

url: `/gov/proposals`

Functionality: Submit a proposal

POST Body:

```js
{
	"base_req": {
     // Name of key to use
    "name": "string",
    // Password for that key
    "password": "string",
    "chain_id": "string",
    "account_number": 64,
    "sequence": 64,
    "gas": 64,
  },
  // Title of the proposal
  "title": "string",
  // Description of the proposal
  "description": "string",
  // PlainTextProposal supported now. SoftwareUpgradeProposal and other types may be supported soon
  "proposal_type": "string",
  // A cosmosaccaddr address
  "proposer": "string",
  "initial_deposit": [
      {
	      "denom": "string",
        "amount": 64,
      }
  ]
}
```

Returns on success:

```js
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
      "TODO": "TODO",
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

### GET /gov/proposals/{proposal-id}/votes

url: `/gov/proposals/{proposal-id}/votes`

Functionality: Query all votes from a specific proposal

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
     "votes":[
         "TODO"
     ]
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



### POST /gov/proposals/{proposal-id}/votes

url: `/gov/proposals/{proposal-id}/votes`

Functionality: Vote for a specific proposal

POST Body:

```js
{
	"base_req": {
     // Name of key to use
    "name": "string",
    // Password for that key
    "password": "string",
    "chain_id": "string",
    "account_number": 64,
    "sequence": 64,
    "gas": 64,
  },
  // A cosmosaccaddr address
  "voter": "string",
  // Value of the vote option `Yes`, `No` `Abstain`, `NoWithVeto`
  "option": "string",
}
```

Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{
      "TODO": "TODO",
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


## ICS23 - SlashingAPI

The SlashingAPI exposes all functionality needed for to slash validators and delegators in PoS.

### GET /slashing/validator/{validatorAddr}/signing-info

url: `/slashing/validator/{validatorAddr}/signing-info`

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

### POST /slashing/validators/{validatorAddr}/unrevokes

url: `/slashing/validators/{validatorAddr}/unrevoke`

Functionality: Query the information from a single validator.

POST Body:

```js
{
  // Name of key to use
  "name": "string",
  // Password for that key
  "password": "string",
  "chain_id": "string",
  "account_number": 64,
  "sequence": 64,
  "gas": 64,
}
```

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
