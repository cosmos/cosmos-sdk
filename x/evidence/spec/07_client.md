# Client

## CLI

A user can query and interact with the `evidence` module using the CLI.

### Query

The `query` commands allows users to query `evidence` state.

```bash
simd query evidence --help
```

### evidence

The `evidence` command allows users to list all evidence or evidence by hash.

Usage:

```bash
simd query evidence [flags]
```

To query evidence by hash

Example:

```bash
simd query evidence "DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660"
```

Example Output:

```bash
evidence:
  consensus_address: cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h
  height: 11
  power: 100
  time: "2021-10-20T16:08:38.194017624Z"
```

To get all evidence

Example:

```bash
simd query evidence
```

Example Output:

```bash
evidence:
  consensus_address: cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h
  height: 11
  power: 100
  time: "2021-10-20T16:08:38.194017624Z"
pagination:
  next_key: null
  total: "1"
```

## REST

A user can query the `evidence` module using REST endpoints.

### Evidence

Get evidence by hash

```bash
/cosmos/evidence/v1beta1/evidence/{evidence_hash}
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/evidence/v1beta1/evidence/DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660"
```

Example Output:

```bash
{
  "evidence": {
    "consensus_address": "cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h",
    "height": "11",
    "power": "100",
    "time": "2021-10-20T16:08:38.194017624Z"
  }
}
```

### All evidence

Get all evidence

```bash
/cosmos/evidence/v1beta1/evidence
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/evidence/v1beta1/evidence"
```

Example Output:

```bash
{
  "evidence": [
    {
      "consensus_address": "cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h",
      "height": "11",
      "power": "100",
      "time": "2021-10-20T16:08:38.194017624Z"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

## gRPC

A user can query the `evidence` module using gRPC endpoints.

### Evidence

Get evidence by hash

```bash
cosmos.evidence.v1beta1.Query/Evidence
```

Example:

```bash
grpcurl -plaintext -d '{"evidence_hash":"DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660"}' localhost:9090 cosmos.evidence.v1beta1.Query/Evidence
```

Example Output:

```bash
{
  "evidence": {
    "consensus_address": "cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h",
    "height": "11",
    "power": "100",
    "time": "2021-10-20T16:08:38.194017624Z"
  }
}
```

### All evidence

Get all evidence

```bash
cosmos.evidence.v1beta1.Query/AllEvidence
```

Example:

```bash
grpcurl -plaintext localhost:9090 cosmos.evidence.v1beta1.Query/AllEvidence
```

Example Output:

```bash
{
  "evidence": [
    {
      "consensus_address": "cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h",
      "height": "11",
      "power": "100",
      "time": "2021-10-20T16:08:38.194017624Z"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```
