<!--
order: 6
-->

# Client

## CLI

A user can query and interact with the `feegrant` module using the CLI.

### Query

The `query` commands allow users to query `feegrant` state.

```sh
simd query feegrant --help
```

#### grant

The `grant` command allows users to query a grant for a given granter-grantee pair.

```sh
simd query feegrant grant [granter] [grantee] [flags]
```

Example:

```sh
simd query feegrant grant cosmos1.. cosmos1..
```

Example Output:

```yml
allowance:
  '@type': /cosmos.feegrant.v1beta1.BasicAllowance
  expiration: null
  spend_limit:
  - amount: "100"
    denom: stake
grantee: cosmos1..
granter: cosmos1..
```

#### grants

The `grants` command allows users to query all grants for a given grantee.

```sh
simd query feegrant grants [grantee] [flags]
```

Example:

```sh
simd query feegrant grants cosmos1..
```

Example Output:

```yml
allowances:
- allowance:
    '@type': /cosmos.feegrant.v1beta1.BasicAllowance
    expiration: null
    spend_limit:
    - amount: "100"
      denom: stake
  grantee: cosmos1..
  granter: cosmos1..
pagination:
  next_key: null
  total: "0"
```

### Transactions

The `tx` commands allow users to interact with the `feegrant` module.

```sh
simd tx feegrant --help
```

#### grant

The `grant` command allows users to grant fee allowances to another account. The fee allowance can have an expiration date, a total spend limit, and/or a periodic spend limit.

```sh
simd tx feegrant grant [granter] [grantee] [flags]
```

Example (one-time spend limit):

```sh
simd tx feegrant grant cosmos1.. cosmos1.. --spend-limit 100stake
```

Example (periodic spend limit):

```sh
simd tx feegrant grant cosmos1.. cosmos1.. --period 3600 --period-limit 10stake
```

#### revoke

The `revoke` command allows users to revoke a granted fee allowance.

```sh
simd tx feegrant revoke [granter] [grantee] [flags]
```

Example:

```sh
simd tx feegrant revoke cosmos1.. cosmos1..
```

## gRPC

A user can query the `feegrant` module using gRPC endpoints.

### Allowance

The `Allowance` endpoint allows users to query a granted fee allowance.

```sh
cosmos.feegrant.v1beta1.Query/Allowance
```

Example:

```sh
grpcurl -plaintext \
    -d '{"grantee":"cosmos1..","granter":"cosmos1.."}' \
    localhost:9090 \
    cosmos.feegrant.v1beta1.Query/Allowance
```

Example Output:

```json
{
  "allowance": {
    "granter": "cosmos1..",
    "grantee": "cosmos1..",
    "allowance": {"@type":"/cosmos.feegrant.v1beta1.BasicAllowance","spendLimit":[{"denom":"stake","amount":"100"}]}
  }
}
```

### Allowances

The `Allowances` endpoint allows users to query all granted fee allowances for a given grantee.

```sh
cosmos.feegrant.v1beta1.Query/Allowances
```

Example:

```sh
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.feegrant.v1beta1.Query/Allowances
```

Example Output:

```json
{
  "allowances": [
    {
      "granter": "cosmos1..",
      "grantee": "cosmos1..",
      "allowance": {"@type":"/cosmos.feegrant.v1beta1.BasicAllowance","spendLimit":[{"denom":"stake","amount":"100"}]}
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```
