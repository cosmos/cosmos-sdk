<!--
order: 6
-->

# Client

## CLI

A user can query and interact with the `mint` module using the CLI.

### Query

The `query` commands allow users to query `mint` state.

```sh
simd query mint --help
```

#### annual-provisions

The `annual-provisions` command allow users to query the current minting annual provisions value

```sh
simd query mint annual-provisions [flags]
```

Example:

```sh
simd query mint annual-provisions
```

Example Output:

```sh
22268504368893.612100895088410693
```

#### inflation

The `inflation` command allow users to query the current minting inflation value

```sh
simd query mint inflation [flags]
```

Example:

```sh
simd query mint inflation
```

Example Output:

```sh
0.199200302563256955
```

#### params

The `params` command allow users to query the current minting parameters

```sh
simd query mint params [flags]
```

Example:

```yml
blocks_per_year: "4360000"
goal_bonded: "0.670000000000000000"
inflation_max: "0.200000000000000000"
inflation_min: "0.070000000000000000"
inflation_rate_change: "0.130000000000000000"
mint_denom: stake
```

## gRPC

A user can query the `mint` module using gRPC endpoints.

### AnnualProvisions

The `AnnualProvisions` endpoint allow users to query the current minting annual provisions value

```sh
/cosmos.mint.v1beta1.Query/AnnualProvisions
```

Example:

```sh
grpcurl -plaintext localhost:9090 cosmos.mint.v1beta1.Query/AnnualProvisions
```

Example Output:

```json
{
  "annualProvisions": "1432452520532626265712995618"
}
```

### Inflation

The `Inflation` endpoint allow users to query the current minting inflation value

```sh
/cosmos.mint.v1beta1.Query/Inflation
```

Example:

```sh
grpcurl -plaintext localhost:9090 cosmos.mint.v1beta1.Query/Inflation
```

Example Output:

```json
{
  "inflation": "130197115720711261"
}
```

### Params

The `Params` endpoint allow users to query the current minting parameters

```sh
/cosmos.mint.v1beta1.Query/Params
```

Example:

```sh
grpcurl -plaintext localhost:9090 cosmos.mint.v1beta1.Query/Params
```

Example Output:

```json
{
  "params": {
    "mintDenom": "stake",
    "inflationRateChange": "130000000000000000",
    "inflationMax": "200000000000000000",
    "inflationMin": "70000000000000000",
    "goalBonded": "670000000000000000",
    "blocksPerYear": "6311520"
  }
}
```

## REST

A user can query the `mint` module using REST endpoints.

### annual-provisions

```sh
/cosmos/mint/v1beta1/annual_provisions
```

Example:

```sh
curl "localhost:1317/cosmos/mint/v1beta1/annual_provisions"
```

Example Output:

```json
{
  "annualProvisions": "1432452520532626265712995618"
}
```

### inflation

```sh
/cosmos/mint/v1beta1/inflation
```

Example:

```sh
curl "localhost:1317/cosmos/mint/v1beta1/inflation"
```

Example Output:

```json
{
  "inflation": "130197115720711261"
}
```

### params

```sh
/cosmos/mint/v1beta1/params
```

Example:

```sh
curl "localhost:1317/cosmos/mint/v1beta1/params"
```

Example Output:

```json
{
  "params": {
    "mintDenom": "stake",
    "inflationRateChange": "130000000000000000",
    "inflationMax": "200000000000000000",
    "inflationMin": "70000000000000000",
    "goalBonded": "670000000000000000",
    "blocksPerYear": "6311520"
  }
}
```
