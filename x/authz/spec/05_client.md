<!--
order: 5
-->

# Client

## CLI

A user can query and interact with the `authz` module using the CLI.

### Query

The `query` commands allow users to query `authz` state.

```bash
simd query authz --help
```

#### grants

The `grants` command allows users to query grants for a granter-grantee pair. If the message type URL is set, it selects grants only for that message type.

```bash
simd query authz grants [granter-addr] [grantee-addr] [msg-type-url]? [flags]
```

Example:

```bash
simd query authz grants cosmos1.. cosmos1.. /cosmos.bank.v1beta1.MsgSend
```

Example Output:

```bash
grants:
- authorization:
    '@type': /cosmos.bank.v1beta1.SendAuthorization
    spend_limit:
    - amount: "100"
      denom: stake
  expiration: "2022-01-01T00:00:00Z"
pagination: null
```

### Transactions

The `tx` commands allow users to interact with the `authz` module.

```bash
simd tx authz --help
```

#### exec

The `exec` command allows a grantee to execute a transaction on behalf of granter.

```bash
  simd tx authz exec [tx-json-file] --from [grantee] [flags]
```

Example:

```bash
simd tx authz exec tx.json --from=cosmos1..
```

#### grant

The `grant` command allows a granter to grant an authorization to a grantee.

```bash
simd tx authz grant <grantee> <authorization_type="send"|"generic"|"delegate"|"unbond"|"redelegate"> --from <granter> [flags]
```

Example:

```bash
simd tx authz grant cosmos1.. send --spend-limit=100stake --from=cosmos1..
```

#### revoke

The `revoke` command allows a granter to revoke an authorization from a grantee.

```bash
simd tx authz revoke [grantee] [msg-type-url] --from=[granter] [flags]
```

Example:

```bash
simd tx authz revoke cosmos1.. /cosmos.bank.v1beta1.MsgSend --from=cosmos1..
```

## gRPC

A user can query the `authz` module using gRPC endpoints.

### Grants

The `Grants` endpoint allows users to query grants for a granter-grantee pair. If the message type URL is set, it selects grants only for that message type.

```bash
cosmos.authz.v1beta1.Query/Grants
```

Example:

```bash
grpcurl -plaintext \
    -d '{"granter":"cosmos1..","grantee":"cosmos1..","msg_type_url":"/cosmos.bank.v1beta1.MsgSend"}' \
    localhost:9090 \
    cosmos.authz.v1beta1.Query/Grants
```

Example Output:

```bash
{
  "grants": [
    {
      "authorization": {
        "@type": "/cosmos.bank.v1beta1.SendAuthorization",
        "spendLimit": [
          {
            "denom":"stake",
            "amount":"100"
          }
        ]
      },
      "expiration": "2022-01-01T00:00:00Z"
    }
  ]
}
```

## REST

A user can query the `authz` module using REST endpoints.

```bash
/cosmos/authz/v1beta1/grants
```

Example:

```bash
curl "localhost:1317/cosmos/authz/v1beta1/grants?granter=cosmos1..&grantee=cosmos1..&msg_type_url=/cosmos.bank.v1beta1.MsgSend"
```

Example Output:

```bash
{
  "grants": [
    {
      "authorization": {
        "@type": "/cosmos.bank.v1beta1.SendAuthorization",
        "spend_limit": [
          {
            "denom": "stake",
            "amount": "100"
          }
        ]
      },
      "expiration": "2022-01-01T00:00:00Z"
    }
  ],
  "pagination": null
}
```
