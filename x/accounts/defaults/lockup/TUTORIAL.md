# Using lockup account on Cosmos sdk

* [Setup](#setup)
* [Init](#init)
* [Execution](#execution)
    * [Delegate](#delegate)
    * [Undelegate](#undelegate)
    * [Withdraw reward](#withdraw-reward)
    * [Withdraw unlocked token](#withdraw-unlocked-token)
    * [Send coins](#send-coins)
* [Query](#query)
    * [Query account info](#query-account-info)
    * [Query periodic lockup account locking periods](#query-periodic-lockup-account-locking-periods)

## Setup 

To create a lockup account we need 2 wallets (newly created or use any of the existing wallet that you have) one for the granter and one for the owner of the lockup account. 

```bash
simd keys add granter --keyring-backend test --home ./.testnets/node0/simd/
simd keys add owner --keyring-backend test --home ./.testnets/node0/simd/
```

## Init

Normally the granter must have enough token to create to the lockup account during the lockup account init process. The owner wallet should be associated with the individual that the granter wants to create the lockup account for.

Now, the granter can craft the lockup account init messages. This message depend on what type of lockup account the granter want to create.

</br>

For continous, delayed, permanent locking account:

```go
{
    "owner": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
    "end_time": 1495793860
    "start_time": 1465793854
}
```

:::infor
*note: `start_time` is only needed for continous locking account init process. For the other two, you dont have to set it in. Error will returned if `start_time` is not provided when creating continous locking account*
:::

</br>
 
For periodic locking account:

```go
    {
      "owner": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
      "locking_periods": [
          {
            "length": 84600
            "amount": {
                "denom": "stake",
                "amount": 2000
            }
          },
          {
            "length": 84600
            "amount": {
                "denom": "stake",
                "amount": 1500
            }
          }
      ]
      "start_time": 1465793854
    }
```

Periodic locking account locking duration is the combines of all the period length from `locking_periods`.
    
The `owner` field takes a string while `start_time` and `end_time` takes a timestamp as value. `locking_periods` are an array of `period`s which consist of 2 field: `length` for the duration of that period and the `amount` that will be release after such duration.

To initialize the account, we have to run the accounts init command passing the account type and the json string for the init message.

```bash
initcontents=$(cat init.json)
simd tx accounts init <lockup_type> $initcontents  --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from granter
```

Whereas the available `lockup_type` options are: 

* continuous-locking-account

* delayed-locking-account

* periodic-locking-account

* permanent-locking-account

If success, we'll check the tx result for the lockup account address. You can send token to it like a normal account. 

## Execution

To execute a message, we can use the command below:

```bash
msgcontents=$(cat msg.json)
simd tx accounts execute <account_address> <execute-msg-type-url> $msgcontents  --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from owner
```

Whereas `execute-msg-type-url` and `msgcontents`  corresponds to lockup account available executions, which are:

### Delegate

The execute message type url for this execution is `cosmos.accounts.defaults.lockup.MsgDelegate`.

Example of json file:

```go
{
    "sender": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
    "validator_address": "cosmosvaloper1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
    "amount": {
        "amount": 100
        "denom": "stake"
    }
}
``` 

:::warning
The `sender` field are the address of the owner of the lockup account. If the sender is not the owner an error will be returned.
:::

### Undelegate

The execute message type url for this execution is `cosmos.accounts.defaults.lockup.MsgUndelegate`.

Example of json file:

```go
{
    "sender": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
    "validator_address": "cosmosvaloper1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
    "amount": {
        "amount": 100
        "denom": "stake"
    }
}
``` 

:::warning
The `sender` field are the address of the owner of the lockup account. If the sender is not the owner an error will be returned.
:::

### Withdraw reward 

The execute message type url for this execution is `cosmos.accounts.defaults.lockup.MsgWithdrawReward`.

Example of json file:

```go
{
    "sender": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
    "validator_address": "cosmosvaloper1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
}
```

:::warning
The `sender` field are the address of the owner of the lockup account. If the sender is not the owner an error will be returned.
:::

### Withdraw unlocked token

The execute message type url for this execution is `cosmos.accounts.defaults.lockup.MsgWithdraw`.

Example of json file:

```go
{
    // lockup account owner address
    "withdrawer": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx46",
    // withdraw to an account of choice
    "to_address": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx47",
    "denoms": ["stake"]
}
``` 

:::warning
The `withdrawer` field are the address of the owner of the lockup account. If the sender is not the owner an error will be returned.
:::

### Send coins

The execute message type url for this execution is `cosmos.accounts.defaults.lockup.MsgSend`.

Example of json file:

```go
{
    "sender": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx45",
    "to_address": "cosmos1vaqh39cdex9sgr69ef0tdln5cn0hdyd3s0lx46",
    "amount": {
        "amount": 100
        "denom": "stake"
    }
}
``` 

:::warning
The `sender` field are the address of the owner of the lockup account. If the sender is not the owner an error will be returned.
:::

## Query

To query a lockup account state, we can use the command below:

```bash
querycontents=$(cat query.json)
simd tx accounts query <account_address> <query-request-type-url> $querycontents  --fees 5stake --chain-id $CHAINID --keyring-backend test --home ./.testnets/node0/simd/ --from owner
```

### Query account info

The query request type url for this query is `cosmos.accounts.defaults.lockup.QueryLockupAccountInfoRequest`. And query json file can be an empty object since `QueryLockupAccountInfoRequest` does not required an input.

Account informations including:

* original locked amount

* delegated amount that are locked

* delegated amount that are free

* start and end time

* owner address

* current locked and unlocked amount

### Query periodic lockup account locking periods

:::info
Note, can only be queried from a periodic lockup account
:::

The query request type url for this query is `cosmos.accounts.defaults.lockup.QueryLockingPeriodsRequest`. And query json file can be an empty object since `QueryLockingPeriodsRequest` does not required an input.

Locking periods including:

* List of period with its duration and amount


