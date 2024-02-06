<!--
order: 6
-->

# Client

## CLI

A user can query and interact with the `bank` module using the CLI.

### Query

The `query` commands allow users to query `bank` state.

```sh
simd query bank --help
```

#### balances

The `balances` command allows users to query account balances by address.

```sh
simd query bank balances [address] [flags]
```

Example:

```sh
simd query bank balances cosmos1..
```

Example Output:

```yml
balances:
- amount: "1000000000"
  denom: stake
pagination:
  next_key: null
  total: "0"
```

#### denom-metadata

The `denom-metadata` command allows users to query metadata for coin denominations. A user can query metadata for a single denomination using the `--denom` flag or all denominations without it.

```sh
simd query bank denom-metadata [flags]
```

Example:

```sh
simd query bank denom-metadata --denom stake
```

Example Output:

```yml
metadata:
  base: stake
  denom_units:
  - aliases:
    - STAKE
    denom: stake
  description: native staking token of simulation app
  display: stake
  name: SimApp Token
  symbol: STK
```

#### total

The `total` command allows users to query the total supply of coins. A user can query the total supply for a single coin using the `--denom` flag or all coins without it.

```sh
simd query bank total [flags]
```

Example:

```sh
simd query bank total --denom stake
```

Example Output:

```yml
amount: "10000000000"
denom: stake
```

### Transactions

The `tx` commands allow users to interact with the `bank` module.

```sh
simd tx bank --help
```

#### send

The `send` command allows users to send funds from one account to another.

```sh
simd tx bank send [from_key_or_address] [to_address] [amount] [flags]
```

Example:

```sh
simd tx bank send cosmos1.. cosmos1.. 100stake
```

## gRPC

A user can query the `bank` module using gRPC endpoints.

### Balance

The `Balance` endpoint allows users to query account balance by address for a given denomination.

```sh
cosmos.bank.v1beta1.Query/Balance
```

Example:

```sh
grpcurl -plaintext \
    -d '{"address":"cosmos1..","denom":"stake"}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/Balance
```

Example Output:

```json
{
  "balance": {
    "denom": "stake",
    "amount": "1000000000"
  }
}
```

### AllBalances

The `AllBalances` endpoint allows users to query account balance by address for all denominations.

```sh
cosmos.bank.v1beta1.Query/AllBalances
```

Example:

```sh
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/AllBalances
```

Example Output:

```json
{
  "balances": [
    {
      "denom": "stake",
      "amount": "1000000000"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### DenomMetadata

The `DenomMetadata` endpoint allows users to query metadata for a single coin denomination.

```sh
cosmos.bank.v1beta1.Query/DenomMetadata
```

Example:

```sh
grpcurl -plaintext \
    -d '{"denom":"stake"}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/DenomMetadata
```

Example Output:

```json
{
  "metadata": {
    "description": "native staking token of simulation app",
    "denomUnits": [
      {
        "denom": "stake",
        "aliases": [
          "STAKE"
        ]
      }
    ],
    "base": "stake",
    "display": "stake",
    "name": "SimApp Token",
    "symbol": "STK"
  }
}
```

### DenomsMetadata

The `DenomsMetadata` endpoint allows users to query metadata for all coin denominations.

```sh
cosmos.bank.v1beta1.Query/DenomsMetadata
```

Example:

```sh
grpcurl -plaintext \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/DenomsMetadata
```

Example Output:

```json
{
  "metadatas": [
    {
      "description": "native staking token of simulation app",
      "denomUnits": [
        {
          "denom": "stake",
          "aliases": [
            "STAKE"
          ]
        }
      ],
      "base": "stake",
      "display": "stake",
      "name": "SimApp Token",
      "symbol": "STK"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### DenomOwners

The `DenomOwners` endpoint allows users to query metadata for a single coin denomination.

```sh
cosmos.bank.v1beta1.Query/DenomOwners
```

Example:

```sh
grpcurl -plaintext \
    -d '{"denom":"stake"}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/DenomOwners
```

Example Output:

```json
{
  "denomOwners": [
    {
      "address": "cosmos1..",
      "balance": {
        "denom": "stake",
        "amount": "5000000000"
      }
    },
    {
      "address": "cosmos1..",
      "balance": {
        "denom": "stake",
        "amount": "5000000000"
      }
    },
  ],
  "pagination": {
    "total": "2"
  }
}
```

### TotalSupply

The `TotalSupply` endpoint allows users to query the total supply of all coins.

```sh
cosmos.bank.v1beta1.Query/TotalSupply
```

Example:

```sh
grpcurl -plaintext \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/TotalSupply
```

Example Output:

```json
{
  "supply": [
    {
      "denom": "stake",
      "amount": "10000000000"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### SupplyOf

The `SupplyOf` endpoint allows users to query the total supply of a single coin.

```sh
cosmos.bank.v1beta1.Query/SupplyOf
```

Example:

```sh
grpcurl -plaintext \
    -d '{"denom":"stake"}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/SupplyOf
```

Example Output:

```json
{
  "amount": {
    "denom": "stake",
    "amount": "10000000000"
  }
}
```

### Params

The `Params` endpoint allows users to query the parameters of the `bank` module.

```sh
cosmos.bank.v1beta1.Query/Params
```

Example:

```sh
grpcurl -plaintext \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/Params
```

Example Output:

```json
{
  "params": {
    "defaultSendEnabled": true
  }
}
```
