<!--
order: 7
-->

# Client

# Auth

## CLI

A user can query and interact with the `auth` module using the CLI.

### Query

The `query` commands allow users to query `auth` state.

```bash
simd query auth --help
```

#### account

The `account` command allow users to query for an account by it's address.

```bash
simd query auth account [address] [flags]
```

Example:

```bash
simd query auth account cosmos1...
```

Example Output:

```bash
'@type': /cosmos.auth.v1beta1.BaseAccount
account_number: "0"
address: cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2
pub_key:
  '@type': /cosmos.crypto.secp256k1.PubKey
  key: ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD
sequence: "1"
```

#### accounts

The `accounts` command allow users to query all the available accounts.

```bash
simd query auth accounts [flags]
```

Example:

```bash
simd query auth accounts
```

Example Output:

```bash
accounts:
- '@type': /cosmos.auth.v1beta1.BaseAccount
  account_number: "0"
  address: cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2
  pub_key:
    '@type': /cosmos.crypto.secp256k1.PubKey
    key: ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD
  sequence: "1"
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "8"
    address: cosmos1yl6hdjhmkf37639730gffanpzndzdpmhwlkfhr
    pub_key: null
    sequence: "0"
  name: transfer
  permissions:
  - minter
  - burner
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "4"
    address: cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh
    pub_key: null
    sequence: "0"
  name: bonded_tokens_pool
  permissions:
  - burner
  - staking
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "5"
    address: cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r
    pub_key: null
    sequence: "0"
  name: not_bonded_tokens_pool
  permissions:
  - burner
  - staking
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "6"
    address: cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn
    pub_key: null
    sequence: "0"
  name: gov
  permissions:
  - burner
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "3"
    address: cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl
    pub_key: null
    sequence: "0"
  name: distribution
  permissions: []
- '@type': /cosmos.auth.v1beta1.BaseAccount
  account_number: "1"
  address: cosmos147k3r7v2tvwqhcmaxcfql7j8rmkrlsemxshd3j
  pub_key: null
  sequence: "0"
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "7"
    address: cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q
    pub_key: null
    sequence: "0"
  name: mint
  permissions:
  - minter
- '@type': /cosmos.auth.v1beta1.ModuleAccount
  base_account:
    account_number: "2"
    address: cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta
    pub_key: null
    sequence: "0"
  name: fee_collector
  permissions: []
pagination:
  next_key: null
  total: "0"
```

#### params

The `params` command allow users to query the current auth parameters.

```bash
simd query auth params [flags]
```

Example:

```bash
simd query auth params
```

Example Output:

```bash
max_memo_characters: "256"
sig_verify_cost_ed25519: "590"
sig_verify_cost_secp256k1: "1000"
tx_sig_limit: "7"
tx_size_cost_per_byte: "10"
```

## gRPC

A user can query the `auth` module using gRPC endpoints.

### Account

The `account` endpoint allow users to query for an account by it's address.

```bash
cosmos.auth.v1beta1.Query/Account
```

Example:

```bash
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/Account
```

Example Output:

```bash
{
  "account":{
    "@type":"/cosmos.auth.v1beta1.BaseAccount",
    "address":"cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2",
    "pubKey":{
      "@type":"/cosmos.crypto.secp256k1.PubKey",
      "key":"ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD"
    },
    "sequence":"1"
  }
}
```

### Accounts

The `accounts` endpoint allow users to query all the available accounts.

```bash
cosmos.auth.v1beta1.Query/Accounts
```

Example:

```bash
grpcurl -plaintext \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/Accounts
```

Example Output:

```bash
{
   "accounts":[
      {
         "@type":"/cosmos.auth.v1beta1.BaseAccount",
         "address":"cosmos1zwg6tpl8aw4rawv8sgag9086lpw5hv33u5ctr2",
         "pubKey":{
            "@type":"/cosmos.crypto.secp256k1.PubKey",
            "key":"ApDrE38zZdd7wLmFS9YmqO684y5DG6fjZ4rVeihF/AQD"
         },
         "sequence":"1"
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1yl6hdjhmkf37639730gffanpzndzdpmhwlkfhr",
            "accountNumber":"8"
         },
         "name":"transfer",
         "permissions":[
            "minter",
            "burner"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
            "accountNumber":"4"
         },
         "name":"bonded_tokens_pool",
         "permissions":[
            "burner",
            "staking"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r",
            "accountNumber":"5"
         },
         "name":"not_bonded_tokens_pool",
         "permissions":[
            "burner",
            "staking"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
            "accountNumber":"6"
         },
         "name":"gov",
         "permissions":[
            "burner"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd88lyufl",
            "accountNumber":"3"
         },
         "name":"distribution"
      },
      {
         "@type":"/cosmos.auth.v1beta1.BaseAccount",
         "accountNumber":"1",
         "address":"cosmos147k3r7v2tvwqhcmaxcfql7j8rmkrlsemxshd3j"
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q",
            "accountNumber":"7"
         },
         "name":"mint",
         "permissions":[
            "minter"
         ]
      },
      {
         "@type":"/cosmos.auth.v1beta1.ModuleAccount",
         "baseAccount":{
            "address":"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
            "accountNumber":"2"
         },
         "name":"fee_collector"
      }
   ],
   "pagination":{
      "total":"9"
   }
}
```

### Params

The `params` endpoint allow users to query the current auth parameters.

```bash
cosmos.auth.v1beta1.Query/Params
```

Example:

```bash
grpcurl -plaintext \
    localhost:9090 \
    cosmos.auth.v1beta1.Query/Params
```

Example Output:

```bash
{
  "params": {
    "maxMemoCharacters": "256",
    "txSigLimit": "7",
    "txSizeCostPerByte": "10",
    "sigVerifyCostEd25519": "590",
    "sigVerifyCostSecp256k1": "1000"
  }
}
```

## REST

A user can query the `auth` module using REST endpoints.

### Account

The `account` endpoint allow users to query for an account by it's address.

```bash
/cosmos/auth/v1beta1/account?address={address}
```

### Accounts

The `accounts` endpoint allow users to query all the available accounts.

```bash
/cosmos/auth/v1beta1/accounts
```

### Params

The `params` endpoint allow users to query the current auth parameters.

```bash
/cosmos/auth/v1beta1/params
```

# Vesting

## CLI

A user can query and interact with the `vesting` module using the CLI.

### Transactions

The `tx` commands allow users to interact with the `vesting` module.

```bash
simd tx vesting --help
```

#### create-periodic-vesting-account

The `create-periodic-vesting-account` command creates a new vesting account funded with an allocation of tokens, where a sequence of coins and period length in seconds. Periods are sequential, in that the duration of of a period only starts at the end of the previous period. The duration of the first period starts upon account creation.

```bash
simd tx vesting create-periodic-vesting-account [to_address] [periods_json_file] [flags]
```

Example:

```bash
simd tx vesting create-periodic-vesting-account cosmos1.. periods.json
```

#### create-vesting-account

The `create-vesting-account` command creates a new vesting account funded with an allocation of tokens. The account can either be a delayed or continuous vesting account, which is determined by the '--delayed' flag. All vesting accounts created will have their start time set by the committed block's time unless specified explicitly using the `--start-time`flag. The `end_time` must be provided as a UNIX epoch timestamp.

```bash
simd tx vesting create-vesting-account [to_address] [amount] [end_time] [flags]
```

Example:

```bash
simd tx vesting create-vesting-account cosmos1.. 100stake 2592000
```
