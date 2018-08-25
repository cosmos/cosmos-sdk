# Cosmos Hub (Gaia-Lite) LCD API

This document describes the API that is exposed by the specific Light Client Daemon (LCD) implementation of the Cosmos Hub (Gaia). Those APIs are exposed by a REST server and can easily be accessed over HTTP/WS (websocket)
connections.

The complete API is comprised of the sub-APIs of different modules. The modules in the Cosmos Hub (Gaia-Lite) API are:

- ICS0 ([TendermintAPI](api.md#ics0---tendermintapi))
- ICS1 ([KeyAPI](api.md#ics1---keyapi))
- ICS20 ([TokenAPI](api.md#ics20---tokenapi))
- ICS21 ([StakingAPI](api.md#ics21---stakingapi))
- ICS22 ([GovernanceAPI](api.md#ics22---governanceapi))
- ICS23 ([SlashingAPI](api.md#ics23---slashingapi))

Error messages my change and should be only used for display purposes. Error messages should not be
used for determining the error type.

## ICS0 - TendermintAPI

Exposes the same functionality as the Tendermint RPC from a full node. It aims to have a very similar API.

### POST /txs

- **URL**: `/txs`
- Query Parameters:
  - `?return={sync|async|block}`:
    - `return=sync`: Waits for the transaction to pass `CheckTx`
    - `return=async`: Returns the request immediately after it is received by the server
    - `return=block`: waits for for the transaction to be committed in a block
- POST Body:

```json
{
  "transaction": "string",
  "return": "string",
}
```

- Returns on success:

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

## ICS1 - KeyAPI

This API exposes all functionality needed for key creation, signing and management.

### GET /keys

- **URL**: `/keys`
- **Functionality**: Gets a list of all the keys.
- Returns on success:

```json
{
    "rest api":"1.0",
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

### POST /keys

- **URL**: `/keys`
- **Functionality**: Create a new key.
- POST Body:

```json
{
  "name": "string",
  "password": "string",
  "seed": "string",
}
```

Returns on success:

```json
{
    "rest api":"1.0",
    "code":200,
    "error":"",
    "result":{
        "seed":"crime carpet recycle erase simple prepare moral dentist fee cause pitch trigger when velvet animal abandon"
    }
}
```

### GET /keys/{name}

- **URL** : `/keys/{name}`
- **Functionality**: Get the information for the specified key.
- Returns on success:

```json
{
    "rest api":"1.0",
    "code":200,
    "error":"",
    "result":{
        "name":"test",
            "address":"cosmosaccaddr1thlqhjqw78zvcy0ua4ldj9gnazqzavyw4eske2",
            "pub_key":"cosmosaccpub1zcjduc3qyx6hlf825jcnj39adpkaxjer95q7yvy25yhfj3dmqy2ctev0rxmse9cuak"
    }
}
```

### PUT /keys/{name}

- **URL** : `/keys/{name}`
- **Functionality**: Change the encryption password for the specified key.
- PUT Body:

```json
{
  "old_password": "string",
  "new_password": "string",
}
```

- Returns on success:

```json
{
    "rest api":"2.0",
    "code":200,
    "error":"",
    "result":{}
}
```

### DELETE /keys/{name}

- **URL**: `/keys/{name}`
- **Functionality**: Delete the specified key.
- DELETE Body:

```json
{
  "password": "string",
}
```

- Returns on success:

```json
{
    "rest api":"1.0",
    "code":200,
    "error":"",
    "result":{}
}
```

### POST /keys/{name}/recover

- **URL**: `/keys/{name}/recover`
- **Functionality**: Recover your key from seed and persist it encrypted with the password.
- POST Body:

```json
{
  "password": "string",
  "seed": "string",
}
```

- Returns on success:

```json
{
    "rest api":"1.0",
    "code":200,
    "error":"",
    "result":{
        "address":"BD607C37147656A507A5A521AA9446EB72B2C907"
    }
}
```

### GET /auth/accounts/{address}

- **URL**: `/auth/accounts/{address}`
- **Functionality**: Query the information of an account .
- Returns on success:

```json
{
    "rest api":"1.0",
    "code":200,
    "error":"",
    "result":{
        "address": "82A57F8575BDFA22F5164C75361A21D2B0E11089",
        "public_key": "PubKeyEd25519{A0EEEED3C9CE1A6988DEBFE347635834A1C0EBA0B4BB1125896A7072D22E650D}",
        "coins": [
            {"atom": 300},
            {"photon": 15}
        ],
        "account_number": 1,
        "sequence": 7
    }
}
}
```

## ICS20 - TokenAPI

The TokenAPI exposes all functionality needed to query account balances and send transactions.

### GET /bank/balance/{account}

- **URL**: `/bank/balance/{account}`
- **Functionality**: Query the specified account's balance.
- Returns on success:

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

### POST /bank/transfers

- **URL**: `/bank/transfers`
- **Functionality**: Create a transfer in the bank module.
- POST Body:

```json
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

- Returns on success:

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

## ICS21 - StakingAPI

The StakingAPI exposes all functionality needed for validation and delegation in Proof-of-Stake.

### GET /stake/delegators/{delegatorAddr}

- **URL**: `/stake/delegators/{delegatorAddr}`
- **Functionality**: Get all delegations (delegation, undelegation) from a delegator.
- Returns on success:

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

### GET /stake/delegators/{delegatorAddr}/validators

- **URL**: `/stake/delegators/{delegatorAddr}/validators`
- **Functionality**: Query all validators that a delegator is bonded to.
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{}
}
```

### GET /stake/delegators/{delegatorAddr}/validators/{validatorAddr}

- **URL**: `/stake/delegators/{delegatorAddr}/validators/{validatorAddr}`
- **Functionality**: Query a validator that a delegator is bonded to
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{}
}
```

### GET /stake/delegators/{delegatorAddr}/txs

- **URL**: `/stake/delegators/{delegatorAddr}/txs`
- **Functionality**: Get all staking txs (i.e msgs) from a delegator.
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

### POST /stake/delegators/{delegatorAddr}/delegations

- **URL**: `/stake/delegators/{delegatorAddr}/delegations`
- **Functionality**: Submit or edit a delegation.
  <!--NOTE Should this be a PUT instead of a POST? the code indicates that this is an edit operation-->
- POST Body:

```json
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
  "begin_redelegates": [
    {
      "delegator_addr": "string",
      "validator_src_addr": "string",
      "validator_dst_addr": "string",
      "shares": "string",
    }
  ],
  "complete_redelegates": [
    {
      "delegator_addr": "string",
      "validator_src_addr": "string",
      "validator_dst_addr": "string",
    }
  ]
}

```

- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

### GET /stake/delegators/{delegatorAddr}/delegations/{validatorAddr}

- **URL**: `/stake/delegators/{delegatorAddr}/delegations/{validatorAddr}`
- **Functionality**: Query the current delegation status between a delegator and a validator.
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

### GET /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}

- **URL**: `/stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}`
- **Functionality**: Query all unbonding delegations between a delegator and a validator.
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

### GET /stake/validators

- **URL**: `/stake/validators`
- **Functionality**: Get all validator candidates.
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

### GET /stake/validators/{validatorAddr}

- **URL**: `/stake/validators/{validatorAddr}`
- **Functionality**: Query the information from a single validator.
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

### GET /stake/parameters

- **URL**: `/stake/parameters`
- **Functionality**: Get the current value of staking parameters.
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
      "inflation_rate_change": 1300000000,
      "inflation_max": 2000000000,
      "inflation_min": 700000000,
      "goal_bonded": 6700000000,
      "unbonding_time": "72h0m0s",
      "max_validators": 100,
      "bond_denom": "atom"
    }
}
```

### GET /stake/pool

- **URL**: `/stake/pool`
- **Functionality**: Get the current value of the dynamic parameters of the current state (*i.e* `Pool`).
- Returns on success:

```json
{
    "rest api":"2.1",
    "code":200,
    "error":"",
    "result":{
      "loose_tokens": 0,
      "bonded_tokens": 0,
      "inflation_last_time": "1970-01-01 01:00:00 +0100 CET",
      "inflation": 700000000,
      "date_last_commission_reset": 0,
      "prev_bonded_shares": 0,
    }
}
```

## ICS22 - GovernanceAPI

The GovernanceAPI exposes all functionality needed for casting votes on plain text, software upgrades and parameter change proposals.

### GET /gov/proposals

- **URL**: `/gov/proposals`
- **Functionality**: Query all submited proposals
- Response on Success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
     "proposals":[
         "TODO"
      ]
    }
}
```

### POST /gov/proposals

- **URL**: `/gov/proposals`
- **Functionality**: Submit a proposal
- POST Body:

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
    	"gas": 64
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

- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
      "TODO": "TODO",
    }
}
```

### GET /gov/proposals/{proposal-id}

- **URL**: `/gov/proposals/{proposal-id}`
- **Functionality**: Query a proposal
- Response on Success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
        "proposal_id": 1,
        "title": "Example title",
        "description": "a larger description with the details of the proposal",
        "proposal_type": "Text",
        "proposal_status": "DepositPeriod",
        "tally_result": {
            "yes": 0,
            "abstain": 0,
            "no": 0,
            "no_with_veto": 0
        },
        "submit_block": 5238512,
        "total_deposit": {"atom": 50},
    	"voting_start_block": -1
    }
}
```

### POST /gov/proposals/{proposal-id}/deposits

- **URL**: `/gov/proposals/{proposal-id}/deposits`
- **Functionality**: Submit or rise a deposit to a proposal in order to make it active
- POST Body:

```json
{
    "base_req": {
    	"name": "string",
    	"password": "string",
    	"chain_id": "string",
        "account_number": 0,
    	"sequence": 0,
    	"gas": 0
  },
  "depositer": "string",
  "amount": 0,
}
```

- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
      "TODO": "TODO",
    }
}
```

### GET /gov/proposals/{proposal-id}/deposits/{address}

- **URL**: `/gov/proposals/{proposal-id}/deposits/{address}`
- **Functionality**: Query a validator's deposit to submit a proposal
- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
        "amount": {"atom": 150},
        "depositer": "cosmosaccaddr1fedh326uxqlxs8ph9ej7cf854gz7fd5zlym5pd",
        "proposal-id": 16
    }
}
```

### GET /gov/proposals/{proposal-id}/tally

- **URL**: `/gov/proposals/{proposal-id}/tally`
- **Functionality**: Get the tally of a given proposal.
- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result": {
        "yes": 0,
        "abstain": 0,
        "no": 0,
        "no_with_veto": 0
    }
}
```



### GET /gov/proposals/{proposal-id}/votes

- **URL**: `/gov/proposals/{proposal-id}/votes`
- **Functionality**: Query all votes from a specific proposal
- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result": [
        {
            "proposal-id": 1,
        	"voter": "cosmosaccaddr1fedh326uxqlxs8ph9ej7cf854gz7fd5zlym5pd",
        	"option": "no_with_veto"
    	},
        {
            "proposal-id": 1,
        	"voter": "cosmosaccaddr1849m9wncrqp6v4tkss6a3j8uzvuv0cp7f75lrq",
        	"option": "yes"
    	},
    ]
}
```



### POST /gov/proposals/{proposal-id}/votes

- **URL**: `/gov/proposals/{proposal-id}/votes`
- **Functionality**: Vote for a specific proposal
- POST Body:

```js
{
	"base_req": {
    	"name": "string",
    	"password": "string",
    	"chain_id": "string",
    	"account_number": 0,
    	"sequence": 0,
    	"gas": 0
  	},
    // A cosmosaccaddr address
  	"voter": "string",
  	// Value of the vote option `Yes`, `No` `Abstain`, `NoWithVeto`
  	"option": "string",
}
```

- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
      "TODO": "TODO",
    }
}
```

### GET /gov/proposals/{proposal-id}/votes/{address}

- **URL** : `/gov/proposals/{proposal-id}/votes/{address}`
- **Functionality**: Get the current `Option` submited by an address
- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
        "proposal-id": 1,
        "voter": "cosmosaccaddr1fedh326uxqlxs8ph9ej7cf854gz7fd5zlym5pd",
        "option": "no_with_veto"
    }
}
```

## ICS23 - SlashingAPI

The SlashingAPI exposes all functionalities needed to slash (*i.e* penalize) validators and delegators in Proof-of-Stake. The penalization is a fine of the staking coin and jail time, defined by governance parameters. During the jail period, the penalized validator is "jailed".

### GET /slashing/validator/{validatorAddr}/signing-info

- **URL**: `/slashing/validator/{validatorAddr}/signing-info`
- **Functionality**: Query the information from a single validator.
- Returns on success:

```json
{
    "rest api":"2.3",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```

### POST /slashing/validators/{validatorAddr}/unjail

- **URL**: `/slashing/validators/{validatorAddr}/unjail`
- **Functionality**: Submit a message to unjail a validator after it has been penalized.
- POST Body:

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

- Returns on success:

```json
{
    "rest api":"2.3",
    "code":200,
    "error":"",
    "result":{
     "transaction":"TODO"
    }
}
```
