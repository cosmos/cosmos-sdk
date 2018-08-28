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
  "return": "async"
}
```

- Returns on success:

```json
{
    
	"code":0,
	"hash":"0D33F2F03A5234F38706E43004489E061AC40A2E",
	"data":"",
	"log":""
}
```

## ICS1 - KeyAPI

This API exposes all functionality needed for key creation, signing and management.

### GET /keys

- **URL**: `/keys`
- **Functionality**: Gets a list of all the keys.
- Returns on success:

```json
[
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
]
```

### POST /keys

- **URL**: `/keys`
- **Functionality**: Create a new key.
- POST Body:

```json
{
    "name": "string",
    "password": "string",
    "seed": "string"
}
```

Returns on success:

```json
{
    "name":"test2",
    "type":"local",
    "address":"cosmosaccaddr17pjnvcae9pktplx0jz5r0vn8ugsqmfuzwsvzzj",
    "pub_key":"cosmosaccpub1addwnpepq0gxp200rysljv9v645llj3y3d3t4vjrdcnl8dnk540fnmu25h6tgustsyt",
    "seed":"muffin novel usual evoke camp canal decade asthma creek lend record media adapt fresh brisk govern plate debris come mother behave coil process next"
}
```

### GET /keys/{name}

- **URL** : `/keys/{name}`
- **Functionality**: Get the information for the specified key.
- Returns on success:

```json
{
  "name": "test1",
  "type": "local",
  "address": "cosmosaccaddr1dla9p5dycmwndmqfehlymwvvjjtfkfw4he576q",
  "pub_key": "cosmosaccpub1addwnpepqgdykrxehg3k9vttjref3dhndtnsy5r3k5f30kekv8elqdluj5ktgl86nwd"
}
```

### PUT /keys/{name}

- **URL** : `/keys/{name}`
- **Functionality**: Change the encryption password for the specified key.
- PUT Body:

```json
{
  "old_password": "string",
  "new_password": "string"
}
```

- Returns on success:

```string
success
```

### DELETE /keys/{name}

- **URL**: `/keys/{name}`
- **Functionality**: Delete the specified key.
- DELETE Body:

```json
{
  "password": "string"
}
```

- Returns on success:

```string
success
```

### POST /keys/{name}/recover

- **URL**: `/keys/{name}/recover`
- **Functionality**: Recover your key from seed and persist it encrypted with the password.
- POST Body:

```json
{
  "password": "string",
  "seed": "string"
}
```

- Returns on success:

```json
{
    "name":"test2",
    "type":"local",
    "address":"cosmosaccaddr17pjnvcae9pktplx0jz5r0vn8ugsqmfuzwsvzzj",
    "pub_key":"cosmosaccpub1addwnpepq0gxp200rysljv9v645llj3y3d3t4vjrdcnl8dnk540fnmu25h6tgustsyt"
}
```

### GET /auth/accounts/{address}

- **URL**: `/auth/accounts/{address}`
- **Functionality**: Query the information of an account .
- Returns on success:

```json
{
    "type": "auth/Account",
    "value": {
        "address": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
        "coins": [
            {
                "denom": "monikerToken",
                "amount": "990"
            },
            {
                "denom": "steak",
                "amount": "49"
            }
        ],
        "public_key": {
            "type": "tendermint/PubKeySecp256k1",
            "value": "Aje2CWOpo0mcrfZy0Q+zSabeHjvT7oEuXuKljLU9agE/"
        },
        "account_number": "0",
        "sequence": "1"
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
[
    {
        "denom":"monikerToken",
        "amount":"990"
    },
    {
        "denom":"steak",
        "amount":"49"
    }
]
```

### POST /bank/transfers

- **URL**: `/bank/transfers`
- **Functionality**: Create a transfer in the bank module.
- POST Body, generate is false:

```json
{
  "account_number": "0",
  "amount": [
    {
      "amount": "10",
      "denom": "monikerToken"
    }
  ],
  "chain_id": "test-chain-KfHYRV",
  "ensure_account_sequence": false,
  "fee": "",
  "gas": "10000",
  "generate": false,
  "name": "moniker",
  "password": "12345678",
  "sequence": "1",
  "to_address": "cosmosaccaddr17pjnvcae9pktplx0jz5r0vn8ugsqmfuzwsvzzj"
}
```

- Returns on success:

```json
{
  "check_tx": {
    "log": "Msg 0: ",
    "gasWanted": "10000",
    "gasUsed": "1242"
  },
  "deliver_tx": {
    "log": "Msg 0: ",
    "gasWanted": "10000",
    "gasUsed": "3288",
    "tags": [
      {
        "key": "c2VuZGVy",
        "value": "Y29zbW9zYWNjYWRkcjF6dmw4cDJ3NnY5MGswZ2VlcnFzbnozZTZqeGEycnpnM3d2dXBtbg=="
      },
      {
        "key": "cmVjaXBpZW50",
        "value": "Y29zbW9zYWNjYWRkcjE3cGpudmNhZTlwa3RwbHgwano1cjB2bjh1Z3NxbWZ1endzdnp6ag=="
      }
    ]
  },
  "hash": "3474141FF827BDB85F39EF94D3B6B93E3D3C2A7A",
  "height": "1574"
}
```

- POST Body, generate is true:

```json
{
    "account_number": "0",
    "amount": [
        {
            "amount": "10",
            "denom": "monikerToken"
        }
    ],
    "chain_id": "test-chain-KfHYRV",
    "ensure_account_sequence": false,
    "fee": "1steak",
    "gas": "10000",
    "generate": true,
    "name": "moniker",
    "password": "12345678",
    "sequence": "2",
    "to_address": "cosmosaccaddr17pjnvcae9pktplx0jz5r0vn8ugsqmfuzwsvzzj"
}
```

- Returns on success:

```json
{
    "account_number": "0",
    "chain_id": "test-chain-KfHYRV",
    "fee": {
        "amount": [
            {
                "amount": "1",
                "denom": "steak"
            }
        ],
        "gas": "10000"
    },
    "memo": "",
    "msgs": [
        {
            "inputs": [
                {
                    "address": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
                    "coins": [
                        {
                            "amount": "10",
                            "denom": "monikerToken"
                        }
                    ]
                }
            ],
            "outputs": [
                {
                    "address": "cosmosaccaddr17pjnvcae9pktplx0jz5r0vn8ugsqmfuzwsvzzj",
                    "coins": [
                        {
                            "amount": "10",
                            "denom": "monikerToken"
                        }
                    ]
                }
            ]
        }
    ],
    "sequence": "2"
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
    "delegations": [
        {
            "delegator_addr": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
            "validator_addr": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
            "shares": "100.0000000000",
            "height": "0"
        }
    ],
    "unbonding_delegations": null,
    "redelegations": null
}
```

### GET /stake/delegators/{delegatorAddr}/validators

- **URL**: `/stake/delegators/{delegatorAddr}/validators`
- **Functionality**: Query all validators that a delegator is bonded to.
- Returns on success:

```json
[
    {
        "operator": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
        "pub_key": "cosmosvalpub1zcjduepq6lql7kxcsrss2wmaxmsv5afkprdqshhupafqc7h37h0wmlhvkn5qrcky3x",
        "jailed": false,
        "status": 2,
        "tokens": "1000000000000",
        "delegator_shares": "1000000000000",
        "description": {
            "moniker": "moniker",
            "identity": "",
            "website": "",
            "details": ""
        },
        "bond_height": "0",
        "bond_intra_tx_counter": 0,
        "proposer_reward_pool": null,
        "commission": "0",
        "commission_max": "0",
        "commission_change_rate": "0",
        "commission_change_today": "0",
        "prev_bonded_shares": "0"
    }
]
```

### GET /stake/delegators/{delegatorAddr}/validators/{validatorAddr}

- **URL**: `/stake/delegators/{delegatorAddr}/validators/{validatorAddr}`
- **Functionality**: Query a validator that a delegator is bonded to
- Returns on success:

```json
{
    "operator": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
    "pub_key": "cosmosvalpub1zcjduepq6lql7kxcsrss2wmaxmsv5afkprdqshhupafqc7h37h0wmlhvkn5qrcky3x",
    "jailed": false,
    "status": 2,
    "tokens": "1000000000000",
    "delegator_shares": "1000000000000",
    "description": {
        "moniker": "moniker",
        "identity": "",
        "website": "",
        "details": ""
    },
    "bond_height": "0",
    "bond_intra_tx_counter": 0,
    "proposer_reward_pool": null,
    "commission": "0",
    "commission_max": "0",
    "commission_change_rate": "0",
    "commission_change_today": "0",
    "prev_bonded_shares": "0"
}
```

### GET /stake/delegators/{delegatorAddr}/txs

- **URL**: `/stake/delegators/{delegatorAddr}/txs`
- **Functionality**: Get all staking txs (i.e msgs) from a delegator.
- Returns on success:

```json
[
  {
    "hash": "string",
    "height": 0,
    "result": {
      "code": 0,
      "data": "string",
      "gas_used": 0,
      "gas_wanted": 0,
      "info": "string",
      "log": "string",
      "tags": [
        [
          {
            "key": "string",
            "value": 0
          }
        ]
      ]
    },
    "tx": "string"
  }
]
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
      "shares": "string"
    }
  ],
  "complete_unbondings": [
    {
      "delegator_addr": "string",
      "validator_addr": "string"
    }
  ],
  "begin_redelegates": [
    {
      "delegator_addr": "string",
      "validator_src_addr": "string",
      "validator_dst_addr": "string",
      "shares": "string"
    }
  ],
  "complete_redelegates": [
    {
      "delegator_addr": "string",
      "validator_src_addr": "string",
      "validator_dst_addr": "string"
    }
  ]
}

```

- Returns on success:

```json
{
  "check_tx": {
    "log": "Msg 0: ",
    "gasWanted": "10000",
    "gasUsed": "1242"
  },
  "deliver_tx": {
    "log": "Msg 0: ",
    "gasWanted": "10000",
    "gasUsed": "3288",
    "tags": [
      {
        "key": "c2VuZGVy",
        "value": "Y29zbW9zYWNjYWRkcjF6dmw4cDJ3NnY5MGswZ2VlcnFzbnozZTZqeGEycnpnM3d2dXBtbg=="
      },
      {
        "key": "cmVjaXBpZW50",
        "value": "Y29zbW9zYWNjYWRkcjE3cGpudmNhZTlwa3RwbHgwano1cjB2bjh1Z3NxbWZ1endzdnp6ag=="
      }
    ]
  },
  "hash": "3474141FF827BDB85F39EF94D3B6B93E3D3C2A7A",
  "height": "1574"
}
```

### GET /stake/delegators/{delegatorAddr}/delegations/{validatorAddr}

- **URL**: `/stake/delegators/{delegatorAddr}/delegations/{validatorAddr}`
- **Functionality**: Query the current delegation status between a delegator and a validator.
- Returns on success:

```json
{
    "delegator_addr": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
    "validator_addr": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
    "shares": "100.0000000000",
    "height": "0"
}
```

### GET /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}

- **URL**: `/stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}`
- **Functionality**: Query all unbonding delegations between a delegator and a validator.
- Returns on success:

```json
[
  {
    "balance": {
      "amount": "string",
      "denom": "string"
    },
    "creation_height": 0,
    "delegator_addr": "string",
    "initial_balance": {
      "amount": "string",
      "denom": "string"
    },
    "min_time": 0,
    "validator_addr": "string"
  }
]
```

### GET /stake/validators

- **URL**: `/stake/validators`
- **Functionality**: Get all validator candidates.
- Returns on success:

```json
[
    {
        "operator": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
        "pub_key": "cosmosvalpub1zcjduepq6lql7kxcsrss2wmaxmsv5afkprdqshhupafqc7h37h0wmlhvkn5qrcky3x",
        "jailed": false,
        "status": 2,
        "tokens": "1000000000000",
        "delegator_shares": "1000000000000",
        "description": {
            "moniker": "moniker",
            "identity": "",
            "website": "",
            "details": ""
        },
        "bond_height": "0",
        "bond_intra_tx_counter": 0,
        "proposer_reward_pool": null,
        "commission": "0",
        "commission_max": "0",
        "commission_change_rate": "0",
        "commission_change_today": "0",
        "prev_bonded_shares": "0"
    }
]
```

### GET /stake/validators/{validatorAddr}

- **URL**: `/stake/validators/{validatorAddr}`
- **Functionality**: Query the information from a single validator.
- Returns on success:

```json
{
    "operator": "cosmosaccaddr1zvl8p2w6v90k0geerqsnz3e6jxa2rzg3wvupmn",
    "pub_key": "cosmosvalpub1zcjduepq6lql7kxcsrss2wmaxmsv5afkprdqshhupafqc7h37h0wmlhvkn5qrcky3x",
    "jailed": false,
    "status": 2,
    "tokens": "1000000000000",
    "delegator_shares": "1000000000000",
    "description": {
        "moniker": "moniker",
        "identity": "",
        "website": "",
        "details": ""
    },
    "bond_height": "0",
    "bond_intra_tx_counter": 0,
    "proposer_reward_pool": null,
    "commission": "0",
    "commission_max": "0",
    "commission_change_rate": "0",
    "commission_change_today": "0",
    "prev_bonded_shares": "0"
}
```

### GET /stake/parameters

- **URL**: `/stake/parameters`
- **Functionality**: Get the current value of staking parameters.
- Returns on success:

```json
{
    "inflation_rate_change": "1300000000",
    "inflation_max": "2000000000",
    "inflation_min": "700000000",
    "goal_bonded": "6700000000",
    "unbonding_time": "259200000000000",
    "max_validators": 100,
    "bond_denom": "steak"
}
```

### GET /stake/pool

- **URL**: `/stake/pool`
- **Functionality**: Get the current value of the dynamic parameters of the current state (*i.e* `Pool`).
- Returns on success:

```json
{
    "loose_tokens": "500035934654",
    "bonded_tokens": "1000000000000",
    "inflation_last_time": "2018-08-28T06:40:39.617950067Z",
    "inflation": "700002217",
    "date_last_commission_reset": "0",
    "prev_bonded_shares": "0"
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

```json
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
            "amount": 64
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
      "TODO": "TODO"
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
  "amount": 0
}
```

- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
      "TODO": "TODO"
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
    	}
    ]
}
```



### POST /gov/proposals/{proposal-id}/votes

- **URL**: `/gov/proposals/{proposal-id}/votes`
- **Functionality**: Vote for a specific proposal
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
    // A cosmosaccaddr address
  	"voter": "string",
  	// Value of the vote option `Yes`, `No` `Abstain`, `NoWithVeto`
  	"option": "string"
}
```

- Returns on success:

```json
{
    "rest api":"2.2",
    "code":200,
    "error":"",
    "result":{
      "TODO": "TODO"
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

```json
{
  // Name of key to use
  "name": "string",
  // Password for that key
  "password": "string",
  "chain_id": "string",
  "account_number": 64,
  "sequence": 64,
  "gas": 64
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
