<!--
order: 9
-->

# Client

## CLI

A user can query and interact with the `staking` module using the CLI.

### Query

The `query` commands allows users to query `staking` state.

```bash
simd query staking --help
```

#### delegation

The `delegation` command allows users to query delegations for an individual delegator on an individual validator.

Usage:

```bash
simd query staking delegation [delegator-addr] [validator-addr] [flags]
```

Example:

```bash
simd query staking delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

Example Output:

```bash
balance:
  amount: "10000000000"
  denom: stake
delegation:
  delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
  shares: "10000000000.000000000000000000"
  validator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

#### delegations

The `delegations` command allows users to query delegations for an individual delegator on all validators.

Usage:

```bash
simd query staking delegations [delegator-addr] [flags]
```

Example:

```bash
simd query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
```

Example Output:

```bash
delegation_responses:
- balance:
    amount: "10000000000"
    denom: stake
  delegation:
    delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
    shares: "10000000000.000000000000000000"
    validator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
- balance:
    amount: "10000000000"
    denom: stake
  delegation:
    delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
    shares: "10000000000.000000000000000000"
    validator_address: cosmosvaloper1x20lytyf6zkcrv5edpkfkn8sz578qg5sqfyqnp
pagination:
  next_key: null
  total: "0"
```

#### delegations-to

The `delegations-to` command allows users to query delegations on an individual validator.

Usage:

```bash
simd query staking delegations-to [validator-addr] [flags]
```

Example:

```bash
simd query staking delegations-to cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

Example Output:

```bash
- balance:
    amount: "504000000"
    denom: stake
  delegation:
    delegator_address: cosmos1q2qwwynhv8kh3lu5fkeex4awau9x8fwt45f5cp
    shares: "504000000.000000000000000000"
    validator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
- balance:
    amount: "78125000000"
    denom: uixo
  delegation:
    delegator_address: cosmos1qvppl3479hw4clahe0kwdlfvf8uvjtcd99m2ca
    shares: "78125000000.000000000000000000"
    validator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
pagination:
  next_key: null
  total: "0"
```

#### historical-info

The `historical-info` command allows users to query historical information at given height.

Usage:

```bash
simd query staking historical-info [height] [flags]
```

Example:

```bash
simd query staking historical-info 10
```

Example Output:

```bash
header:
  app_hash: Lbx8cXpI868wz8sgp4qPYVrlaKjevR5WP/IjUxwp3oo=
  chain_id: testnet
  consensus_hash: BICRvH3cKD93v7+R1zxE2ljD34qcvIZ0Bdi389qtoi8=
  data_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  evidence_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  height: "10"
  last_block_id:
    hash: RFbkpu6pWfSThXxKKl6EZVDnBSm16+U0l0xVjTX08Fk=
    part_set_header:
      hash: vpIvXD4rxD5GM4MXGz0Sad9I7//iVYLzZsEU4BVgWIU=
      total: 1
  last_commit_hash: Ne4uXyx4QtNp4Zx89kf9UK7oG9QVbdB6e7ZwZkhy8K0=
  last_results_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  next_validators_hash: nGBgKeWBjoxeKFti00CxHsnULORgKY4LiuQwBuUrhCs=
  proposer_address: mMEP2c2IRPLr99LedSRtBg9eONM=
  time: "2021-10-01T06:00:49.785790894Z"
  validators_hash: nGBgKeWBjoxeKFti00CxHsnULORgKY4LiuQwBuUrhCs=
  version:
    app: "0"
    block: "11"
valset:
- commission:
    commission_rates:
      max_change_rate: "0.010000000000000000"
      max_rate: "0.200000000000000000"
      rate: "0.100000000000000000"
    update_time: "2021-10-01T05:52:50.380144238Z"
  consensus_pubkey:
    '@type': /cosmos.crypto.ed25519.PubKey
    key: Auxs3865HpB/EfssYOzfqNhEJjzys2Fo6jD5B8tPgC8=
  delegator_shares: "10000000.000000000000000000"
  description:
    details: ""
    identity: ""
    moniker: myvalidator
    security_contact: ""
    website: ""
  jailed: false
  min_self_delegation: "1"
  operator_address: cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc
  status: BOND_STATUS_BONDED
  tokens: "10000000"
  unbonding_height: "0"
  unbonding_time: "1970-01-01T00:00:00Z"
```

#### params

The `params` command allows users to query values set as staking parameters.

Usage:

```bash
simd query staking params [flags]
```

Example:

```bash
simd query staking params
```

Example Output:

```bash
bond_denom: stake
historical_entries: 10000
max_entries: 7
max_validators: 50
unbonding_time: 1814400s
```

#### pool

The `pool` command allows users to query values for amounts stored in the staking pool.

Usage:

```bash
simd q staking pool [flags]
```

Example:

```bash
simd q staking pool
```

Example Output:

```bash
bonded_tokens: "10000000"
not_bonded_tokens: "0"
```

#### redelegation

The `redelegation` command allows users to query a redelegation record based on delegator and a source and destination validator address.

Usage:

```bash
simd query staking redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr] [flags]
```

Example:

```bash
simd query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

Example Output:

```bash
pagination: null
redelegation_responses:
- entries:
  - balance: "50000000"
    redelegation_entry:
      completion_time: "2021-10-24T20:33:21.960084845Z"
      creation_height: 2.382847e+06
      initial_balance: "50000000"
      shares_dst: "50000000.000000000000000000"
  - balance: "5000000000"
    redelegation_entry:
      completion_time: "2021-10-25T21:33:54.446846862Z"
      creation_height: 2.397271e+06
      initial_balance: "5000000000"
      shares_dst: "5000000000.000000000000000000"
  redelegation:
    delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
    entries: null
    validator_dst_address: cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm
    validator_src_address: cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm
```

#### redelegations

The `redelegations` command allows users to query all redelegation records for an individual delegator.

Usage:

```bash
simd query staking redelegations [delegator-addr] [flags]
```

Example:

```bash
simd query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
```

Example Output:

```bash
pagination:
  next_key: null
  total: "0"
redelegation_responses:
- entries:
  - balance: "50000000"
    redelegation_entry:
      completion_time: "2021-10-24T20:33:21.960084845Z"
      creation_height: 2.382847e+06
      initial_balance: "50000000"
      shares_dst: "50000000.000000000000000000"
  - balance: "5000000000"
    redelegation_entry:
      completion_time: "2021-10-25T21:33:54.446846862Z"
      creation_height: 2.397271e+06
      initial_balance: "5000000000"
      shares_dst: "5000000000.000000000000000000"
  redelegation:
    delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
    entries: null
    validator_dst_address: cosmosvaloper1uccl5ugxrm7vqlzwqr04pjd320d2fz0z3hc6vm
    validator_src_address: cosmosvaloper1zppjyal5emta5cquje8ndkpz0rs046m7zqxrpp
- entries:
  - balance: "562770000000"
    redelegation_entry:
      completion_time: "2021-10-25T21:42:07.336911677Z"
      creation_height: 2.39735e+06
      initial_balance: "562770000000"
      shares_dst: "562770000000.000000000000000000"
  redelegation:
    delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
    entries: null
    validator_dst_address: cosmosvaloper1uccl5ugxrm7vqlzwqr04pjd320d2fz0z3hc6vm
    validator_src_address: cosmosvaloper1zppjyal5emta5cquje8ndkpz0rs046m7zqxrpp
```

#### redelegations-from

The `redelegations-from` command allows users to query delegations that are redelegating _from_ a validator.

Usage:

```bash
simd query staking redelegations-from [validator-addr] [flags]
```

Example:

```bash
simd query staking redelegations-from cosmosvaloper1y4rzzrgl66eyhzt6gse2k7ej3zgwmngeleucjy
```

Example Output:

```bash
pagination:
  next_key: null
  total: "0"
redelegation_responses:
- entries:
  - balance: "50000000"
    redelegation_entry:
      completion_time: "2021-10-24T20:33:21.960084845Z"
      creation_height: 2.382847e+06
      initial_balance: "50000000"
      shares_dst: "50000000.000000000000000000"
  - balance: "5000000000"
    redelegation_entry:
      completion_time: "2021-10-25T21:33:54.446846862Z"
      creation_height: 2.397271e+06
      initial_balance: "5000000000"
      shares_dst: "5000000000.000000000000000000"
  redelegation:
    delegator_address: cosmos1pm6e78p4pgn0da365plzl4t56pxy8hwtqp2mph
    entries: null
    validator_dst_address: cosmosvaloper1uccl5ugxrm7vqlzwqr04pjd320d2fz0z3hc6vm
    validator_src_address: cosmosvaloper1y4rzzrgl66eyhzt6gse2k7ej3zgwmngeleucjy
- entries:
  - balance: "221000000"
    redelegation_entry:
      completion_time: "2021-10-05T21:05:45.669420544Z"
      creation_height: 2.120693e+06
      initial_balance: "221000000"
      shares_dst: "221000000.000000000000000000"
  redelegation:
    delegator_address: cosmos1zqv8qxy2zgn4c58fz8jt8jmhs3d0attcussrf6
    entries: null
    validator_dst_address: cosmosvaloper10mseqwnwtjaqfrwwp2nyrruwmjp6u5jhah4c3y
    validator_src_address: cosmosvaloper1y4rzzrgl66eyhzt6gse2k7ej3zgwmngeleucjy
```

#### unbonding-delegation

The `unbonding-delegation` command allows users to query unbonding delegations for an individual delegator on an individual validator.

Usage:

```bash
simd query staking unbonding-delegation [delegator-addr] [validator-addr] [flags]
```

Example:

```bash
simd query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

Example Output:

```bash
delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
entries:
- balance: "52000000"
  completion_time: "2021-11-02T11:35:55.391594709Z"
  creation_height: "55078"
  initial_balance: "52000000"
validator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

#### unbonding-delegations

The `unbonding-delegations` command allows users to query all unbonding-delegations records for one delegator.

Usage:

```bash
simd query staking unbonding-delegations [delegator-addr] [flags]
```

Example:

```bash
simd query staking unbonding-delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
```

Example Output:

```bash
pagination:
  next_key: null
  total: "0"
unbonding_responses:
- delegator_address: cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
  entries:
  - balance: "52000000"
    completion_time: "2021-11-02T11:35:55.391594709Z"
    creation_height: "55078"
    initial_balance: "52000000"
  validator_address: cosmosvaloper1t8ehvswxjfn3ejzkjtntcyrqwvmvuknzmvtaaa

```

#### unbonding-delegations-from

The `unbonding-delegations-from` command allows users to query delegations that are unbonding _from_ a validator.

Usage:

```bash
simd query staking unbonding-delegations-from [validator-addr] [flags]
```

Example:

```bash
simd query staking unbonding-delegations-from cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

Example Output:

```bash
pagination:
  next_key: null
  total: "0"
unbonding_responses:
- delegator_address: cosmos1qqq9txnw4c77sdvzx0tkedsafl5s3vk7hn53fn
  entries:
  - balance: "150000000"
    completion_time: "2021-11-01T21:41:13.098141574Z"
    creation_height: "46823"
    initial_balance: "150000000"
  validator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
- delegator_address: cosmos1peteje73eklqau66mr7h7rmewmt2vt99y24f5z
  entries:
  - balance: "24000000"
    completion_time: "2021-10-31T02:57:18.192280361Z"
    creation_height: "21516"
    initial_balance: "24000000"
  validator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

#### validator

The `validator` command allows users to query details about an individual validator.

Usage:

```bash
simd query staking validator [validator-addr] [flags]
```

Example:

```bash
simd query staking validator cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
```

Example Output:

```bash
commission:
  commission_rates:
    max_change_rate: "0.020000000000000000"
    max_rate: "0.200000000000000000"
    rate: "0.050000000000000000"
  update_time: "2021-10-01T19:24:52.663191049Z"
consensus_pubkey:
  '@type': /cosmos.crypto.ed25519.PubKey
  key: sIiexdJdYWn27+7iUHQJDnkp63gq/rzUq1Y+fxoGjXc=
delegator_shares: "32948270000.000000000000000000"
description:
  details: Witval is the validator arm from Vitwit. Vitwit is into software consulting
    and services business since 2015. We are working closely with Cosmos ecosystem
    since 2018. We are also building tools for the ecosystem, Aneka is our explorer
    for the cosmos ecosystem.
  identity: 51468B615127273A
  moniker: Witval
  security_contact: ""
  website: ""
jailed: false
min_self_delegation: "1"
operator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
status: BOND_STATUS_BONDED
tokens: "32948270000"
unbonding_height: "0"
unbonding_time: "1970-01-01T00:00:00Z"
```

#### validators

The `validators` command allows users to query details about all validators on a network.

Usage:

```bash
simd query staking validators [flags]
```

Example:

```bash
simd query staking validators
```

Example Output:

```bash
pagination:
  next_key: FPTi7TKAjN63QqZh+BaXn6gBmD5/
  total: "0"
validators:
commission:
  commission_rates:
    max_change_rate: "0.020000000000000000"
    max_rate: "0.200000000000000000"
    rate: "0.050000000000000000"
  update_time: "2021-10-01T19:24:52.663191049Z"
consensus_pubkey:
  '@type': /cosmos.crypto.ed25519.PubKey
  key: sIiexdJdYWn27+7iUHQJDnkp63gq/rzUq1Y+fxoGjXc=
delegator_shares: "32948270000.000000000000000000"
description:
    details: Witval is the validator arm from Vitwit. Vitwit is into software consulting
      and services business since 2015. We are working closely with Cosmos ecosystem
      since 2018. We are also building tools for the ecosystem, Aneka is our explorer
      for the cosmos ecosystem.
    identity: 51468B615127273A
    moniker: Witval
    security_contact: ""
    website: ""
  jailed: false
  min_self_delegation: "1"
  operator_address: cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
  status: BOND_STATUS_BONDED
  tokens: "32948270000"
  unbonding_height: "0"
  unbonding_time: "1970-01-01T00:00:00Z"
- commission:
    commission_rates:
      max_change_rate: "0.100000000000000000"
      max_rate: "0.200000000000000000"
      rate: "0.050000000000000000"
    update_time: "2021-10-04T18:02:21.446645619Z"
  consensus_pubkey:
    '@type': /cosmos.crypto.ed25519.PubKey
    key: GDNpuKDmCg9GnhnsiU4fCWktuGUemjNfvpCZiqoRIYA=
  delegator_shares: "559343421.000000000000000000"
  description:
    details: Noderunners is a professional validator in POS networks. We have a huge
      node running experience, reliable soft and hardware. Our commissions are always
      low, our support to delegators is always full. Stake with us and start receiving
      your Cosmos rewards now!
    identity: 812E82D12FEA3493
    moniker: Noderunners
    security_contact: info@noderunners.biz
    website: http://noderunners.biz
  jailed: false
  min_self_delegation: "1"
  operator_address: cosmosvaloper1q5ku90atkhktze83j9xjaks2p7uruag5zp6wt7
  status: BOND_STATUS_BONDED
  tokens: "559343421"
  unbonding_height: "0"
  unbonding_time: "1970-01-01T00:00:00Z"
```

### Transactions

The `tx` commands allows users to interact with the `staking` module.

```bash
simd tx staking --help
```

#### create-validator

The command `create-validator` allows users to create new validator initialized with a self-delegation to it.

Usage:

```bash
simd tx staking create-validator [flags]
```

Example:

```bash
simd tx staking create-validator \
  --amount=1000000stake \
  --pubkey=$(simd tendermint show-validator) \
  --moniker="my-moniker" \
  --website="https://myweb.site" \
  --details="description of your validator" \
  --chain-id="name_of_chain_id" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
  --gas-adjustment="1.2" \
  --gas-prices="0.025stake" \
  --from=mykey
```

#### delegate

The command `delegate` allows users to delegate liquid tokens to a validator.

Usage:

```bash
simd tx staking delegate [validator-addr] [amount] [flags]
```

Example:

```bash
simd tx staking delegate cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey
```

#### edit-validator

The command `edit-validator` allows users to edit an existing validator account.

Usage:

```bash
simd tx staking edit-validator [flags]
```

Example:

```bash
simd tx staking edit-validator --moniker "new_moniker_name" --website "new_webiste_url" --from mykey
```

#### redelegate

The command `redelegate` allows users to redelegate illiquid tokens from one validator to another.

Usage:

```bash
simd tx staking redelegate [src-validator-addr] [dst-validator-addr] [amount] [flags]
```

Example:

```bash
simd tx staking redelegate cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100stake --from mykey
```

#### unbond

The command `unbond` allows users to unbond shares from a validator.

Usage:

```bash
simd tx staking unbond [validator-addr] [amount] [flags]
```

Example:

```bash
simd tx staking unbond cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake --from mykey
```
#### cancel unbond 
The command `cancel-unbond` allow users to cancel the unbonding delegation entry and delegate back to the original validator.

Usage:
```bash
simd tx staking cancel-unbond [validator-addr] [amount] [creation-height]
```

Example:
```bash
simd tx staking cancel-unbond cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake 123123 --from mykey
```


## gRPC

A user can query the `staking` module using gRPC endpoints.

### Validators

The `Validators` endpoint queries all validators that match the given status.

```bash
cosmos.staking.v1beta1.Query/Validators
```

Example:

```bash
grpcurl -plaintext localhost:9090 cosmos.staking.v1beta1.Query/Validators
```

Example Output:

```bash
{
  "validators": [
    {
      "operatorAddress": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
      "consensusPubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"Auxs3865HpB/EfssYOzfqNhEJjzys2Fo6jD5B8tPgC8="},
      "status": "BOND_STATUS_BONDED",
      "tokens": "10000000",
      "delegatorShares": "10000000000000000000000000",
      "description": {
        "moniker": "myvalidator"
      },
      "unbondingTime": "1970-01-01T00:00:00Z",
      "commission": {
        "commissionRates": {
          "rate": "100000000000000000",
          "maxRate": "200000000000000000",
          "maxChangeRate": "10000000000000000"
        },
        "updateTime": "2021-10-01T05:52:50.380144238Z"
      },
      "minSelfDelegation": "1"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### Validator

The `Validator` endpoint queries validator information for given validator address.

```bash
cosmos.staking.v1beta1.Query/Validator
```

Example:

```bash
grpcurl -plaintext -d '{"validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/Validator
```

Example Output:

```bash
{
  "validator": {
    "operatorAddress": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
    "consensusPubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"Auxs3865HpB/EfssYOzfqNhEJjzys2Fo6jD5B8tPgC8="},
    "status": "BOND_STATUS_BONDED",
    "tokens": "10000000",
    "delegatorShares": "10000000000000000000000000",
    "description": {
      "moniker": "myvalidator"
    },
    "unbondingTime": "1970-01-01T00:00:00Z",
    "commission": {
      "commissionRates": {
        "rate": "100000000000000000",
        "maxRate": "200000000000000000",
        "maxChangeRate": "10000000000000000"
      },
      "updateTime": "2021-10-01T05:52:50.380144238Z"
    },
    "minSelfDelegation": "1"
  }
}
```

### ValidatorDelegations

The `ValidatorDelegations` endpoint queries delegate information for given validator.

```bash
cosmos.staking.v1beta1.Query/ValidatorDelegations
```

Example:

```bash
grpcurl -plaintext -d '{"validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/ValidatorDelegations
```

Example Output:

```bash
{
  "delegationResponses": [
    {
      "delegation": {
        "delegatorAddress": "cosmos1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgy3ua5t",
        "validatorAddress": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
        "shares": "10000000000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "10000000"
      }
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### ValidatorUnbondingDelegations

The `ValidatorUnbondingDelegations` endpoint queries delegate information for given validator.

```bash
cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations
```

Example:

```bash
grpcurl -plaintext -d '{"validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1z3pzzw84d6xn00pw9dy3yapqypfde7vg6965fy",
      "validator_address": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
      "entries": [
        {
          "creation_height": "25325",
          "completion_time": "2021-10-31T09:24:36.797320636Z",
          "initial_balance": "20000000",
          "balance": "20000000"
        }
      ]
    },
    {
      "delegator_address": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
      "validator_address": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
      "entries": [
        {
          "creation_height": "13100",
          "completion_time": "2021-10-30T12:53:02.272266791Z",
          "initial_balance": "1000000",
          "balance": "1000000"
        }
      ]
    },
  ],
  "pagination": {
    "next_key": null,
    "total": "8"
  }
}
```

### Delegation

The `Delegation` endpoint queries delegate information for given validator delegator pair.

```bash
cosmos.staking.v1beta1.Query/Delegation
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77", validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/Delegation
```

Example Output:

```bash
{
  "delegation_response":
  {
    "delegation":
      {
        "delegator_address":"cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
        "validator_address":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
        "shares":"25083119936.000000000000000000"
      },
    "balance":
      {
        "denom":"stake",
        "amount":"25083119936"
      }
  }
}
```

### UnbondingDelegation

The `UnbondingDelegation` endpoint queries unbonding information for given validator delegator.

```bash
cosmos.staking.v1beta1.Query/UnbondingDelegation
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77", validator_addr":"cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc"}' \
localhost:9090 cosmos.staking.v1beta1.Query/UnbondingDelegation
```

Example Output:

```bash
{
  "unbond": {
    "delegator_address": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
    "validator_address": "cosmosvaloper1rne8lgs98p0jqe82sgt0qr4rdn4hgvmgp9ggcc",
    "entries": [
      {
        "creation_height": "136984",
        "completion_time": "2021-11-08T05:38:47.505593891Z",
        "initial_balance": "400000000",
        "balance": "400000000"
      },
      {
        "creation_height": "137005",
        "completion_time": "2021-11-08T05:40:53.526196312Z",
        "initial_balance": "385000000",
        "balance": "385000000"
      }
    ]
  }
}
```

### DelegatorDelegations

The `DelegatorDelegations` endpoint queries all delegations of a given delegator address.

```bash
cosmos.staking.v1beta1.Query/DelegatorDelegations
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorDelegations
```

Example Output:

```bash
{
  "delegation_responses": [
    {"delegation":{"delegator_address":"cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77","validator_address":"cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8","shares":"25083339023.000000000000000000"},"balance":{"denom":"stake","amount":"25083339023"}}
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### DelegatorUnbondingDelegations

The `DelegatorUnbondingDelegations` endpoint queries all unbonding delegations of a given delegator address.

```bash
cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1y8nyfvmqh50p6ldpzljk3yrglppdv3t8phju77",
      "validator_address": "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9uxyejze",
      "entries": [
        {
          "creation_height": "136984",
          "completion_time": "2021-11-08T05:38:47.505593891Z",
          "initial_balance": "400000000",
          "balance": "400000000"
        },
        {
          "creation_height": "137005",
          "completion_time": "2021-11-08T05:40:53.526196312Z",
          "initial_balance": "385000000",
          "balance": "385000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### Redelegations

The `Redelegations` endpoint queries redelegations of given address.

```bash
cosmos.staking.v1beta1.Query/Redelegations
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1ld5p7hn43yuh8ht28gm9pfjgj2fctujp2tgwvf", "src_validator_addr" : "cosmosvaloper1j7euyj85fv2jugejrktj540emh9353ltgppc3g", "dst_validator_addr" : "cosmosvaloper1yy3tnegzmkdcm7czzcy3flw5z0zyr9vkkxrfse"}' \
localhost:9090 cosmos.staking.v1beta1.Query/Redelegations
```

Example Output:

```bash
{
  "redelegation_responses": [
    {
      "redelegation": {
        "delegator_address": "cosmos1ld5p7hn43yuh8ht28gm9pfjgj2fctujp2tgwvf",
        "validator_src_address": "cosmosvaloper1j7euyj85fv2jugejrktj540emh9353ltgppc3g",
        "validator_dst_address": "cosmosvaloper1yy3tnegzmkdcm7czzcy3flw5z0zyr9vkkxrfse",
        "entries": null
      },
      "entries": [
        {
          "redelegation_entry": {
            "creation_height": 135932,
            "completion_time": "2021-11-08T03:52:55.299147901Z",
            "initial_balance": "2900000",
            "shares_dst": "2900000.000000000000000000"
          },
          "balance": "2900000"
        }
      ]
    }
  ],
  "pagination": null
}
```

### DelegatorValidators

The `DelegatorValidators` endpoint queries all validators information for given delegator.

```bash
cosmos.staking.v1beta1.Query/DelegatorValidators
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1ld5p7hn43yuh8ht28gm9pfjgj2fctujp2tgwvf"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorValidators
```

Example Output:

```bash
{
  "validators": [
    {
      "operator_address": "cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "UPwHWxH1zHJWGOa/m6JB3f5YjHMvPQPkVbDqqi+U7Uw="
      },
      "jailed": false,
      "status": "BOND_STATUS_BONDED",
      "tokens": "347260647559",
      "delegator_shares": "347260647559.000000000000000000",
      "description": {
        "moniker": "BouBouNode",
        "identity": "",
        "website": "https://boubounode.com",
        "security_contact": "",
        "details": "AI-based Validator. #1 AI Validator on Game of Stakes. Fairly priced. Don't trust (humans), verify. Made with BouBou love."
      },
      "unbonding_height": "0",
      "unbonding_time": "1970-01-01T00:00:00Z",
      "commission": {
        "commission_rates": {
          "rate": "0.061000000000000000",
          "max_rate": "0.300000000000000000",
          "max_change_rate": "0.150000000000000000"
        },
        "update_time": "2021-10-01T15:00:00Z"
      },
      "min_self_delegation": "1"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### DelegatorValidator

The `DelegatorValidator` endpoint queries validator information for given delegator validator

```bash
cosmos.staking.v1beta1.Query/DelegatorValidator
```

Example:

```bash
grpcurl -plaintext \
-d '{"delegator_addr": "cosmos1eh5mwu044gd5ntkkc2xgfg8247mgc56f3n8rr7", "validator_addr": "cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8"}' \
localhost:9090 cosmos.staking.v1beta1.Query/DelegatorValidator
```

Example Output:

```bash
{
  "validator": {
    "operator_address": "cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fww3vc8",
    "consensus_pubkey": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "UPwHWxH1zHJWGOa/m6JB3f5YjHMvPQPkVbDqqi+U7Uw="
    },
    "jailed": false,
    "status": "BOND_STATUS_BONDED",
    "tokens": "347262754841",
    "delegator_shares": "347262754841.000000000000000000",
    "description": {
      "moniker": "BouBouNode",
      "identity": "",
      "website": "https://boubounode.com",
      "security_contact": "",
      "details": "AI-based Validator. #1 AI Validator on Game of Stakes. Fairly priced. Don't trust (humans), verify. Made with BouBou love."
    },
    "unbonding_height": "0",
    "unbonding_time": "1970-01-01T00:00:00Z",
    "commission": {
      "commission_rates": {
        "rate": "0.061000000000000000",
        "max_rate": "0.300000000000000000",
        "max_change_rate": "0.150000000000000000"
      },
      "update_time": "2021-10-01T15:00:00Z"
    },
    "min_self_delegation": "1"
  }
}
```

### HistoricalInfo

```bash
cosmos.staking.v1beta1.Query/HistoricalInfo
```

Example:

```bash
grpcurl -plaintext -d '{"height" : 1}' localhost:9090 cosmos.staking.v1beta1.Query/HistoricalInfo
```

Example Output:

```bash
{
  "hist": {
    "header": {
      "version": {
        "block": "11",
        "app": "0"
      },
      "chain_id": "simd-1",
      "height": "140142",
      "time": "2021-10-11T10:56:29.720079569Z",
      "last_block_id": {
        "hash": "9gri/4LLJUBFqioQ3NzZIP9/7YHR9QqaM6B2aJNQA7o=",
        "part_set_header": {
          "total": 1,
          "hash": "Hk1+C864uQkl9+I6Zn7IurBZBKUevqlVtU7VqaZl1tc="
        }
      },
      "last_commit_hash": "VxrcS27GtvGruS3I9+AlpT7udxIT1F0OrRklrVFSSKc=",
      "data_hash": "80BjOrqNYUOkTnmgWyz9AQ8n7SoEmPVi4QmAe8RbQBY=",
      "validators_hash": "95W49n2hw8RWpr1GPTAO5MSPi6w6Wjr3JjjS7AjpBho=",
      "next_validators_hash": "95W49n2hw8RWpr1GPTAO5MSPi6w6Wjr3JjjS7AjpBho=",
      "consensus_hash": "BICRvH3cKD93v7+R1zxE2ljD34qcvIZ0Bdi389qtoi8=",
      "app_hash": "ZZaxnSY3E6Ex5Bvkm+RigYCK82g8SSUL53NymPITeOE=",
      "last_results_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
      "evidence_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
      "proposer_address": "aH6dO428B+ItuoqPq70efFHrSMY="
    },
  "valset": [
      {
        "operator_address": "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcqcnylw",
        "consensus_pubkey": {
          "@type": "/cosmos.crypto.ed25519.PubKey",
          "key": "/O7BtNW0pafwfvomgR4ZnfldwPXiFfJs9mHg3gwfv5Q="
        },
        "jailed": false,
        "status": "BOND_STATUS_BONDED",
        "tokens": "1426045203613",
        "delegator_shares": "1426045203613.000000000000000000",
        "description": {
          "moniker": "SG-1",
          "identity": "48608633F99D1B60",
          "website": "https://sg-1.online",
          "security_contact": "",
          "details": "SG-1 - your favorite validator on Witval. We offer 100% Soft Slash protection."
        },
        "unbonding_height": "0",
        "unbonding_time": "1970-01-01T00:00:00Z",
        "commission": {
          "commission_rates": {
            "rate": "0.037500000000000000",
            "max_rate": "0.200000000000000000",
            "max_change_rate": "0.030000000000000000"
          },
          "update_time": "2021-10-01T15:00:00Z"
        },
        "min_self_delegation": "1"
      }
    ]
  }
}

```

### Pool

The `Pool` endpoint queries the pool information.

```bash
cosmos.staking.v1beta1.Query/Pool
```

Example:

```bash
grpcurl -plaintext -d localhost:9090 cosmos.staking.v1beta1.Query/Pool
```

Example Output:

```bash
{
  "pool": {
    "not_bonded_tokens": "369054400189",
    "bonded_tokens": "15657192425623"
  }
}
```

### Params

The `Params` endpoint queries the pool information.

```bash
cosmos.staking.v1beta1.Query/Params
```

Example:

```bash
grpcurl -plaintext localhost:9090 cosmos.staking.v1beta1.Query/Params
```

Example Output:

```bash
{
  "params": {
    "unbondingTime": "1814400s",
    "maxValidators": 100,
    "maxEntries": 7,
    "historicalEntries": 10000,
    "bondDenom": "stake"
  }
}
```

## REST

A user can query the `staking` module using REST endpoints.

### DelegatorDelegations

The `DelegtaorDelegations` REST endpoint queries all delegations of a given delegator address.

```bash
/cosmos/staking/v1beta1/delegations/{delegatorAddr}
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/delegations/cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5" -H  "accept: application/json"
```

Example Output:

```bash
{
  "delegation_responses": [
    {
      "delegation": {
        "delegator_address": "cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5",
        "validator_address": "cosmosvaloper1quqxfrxkycr0uzt4yk0d57tcq3zk7srm7sm6r8",
        "shares": "256250000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "256250000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5",
        "validator_address": "cosmosvaloper194v8uwee2fvs2s8fa5k7j03ktwc87h5ym39jfv",
        "shares": "255150000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "255150000"
      }
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### Redelegations

The `Redelegations` REST endpoint queries redelegations of given address.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/redelegations
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1thfntksw0d35n2tkr0k8v54fr8wxtxwxl2c56e/redelegations?srcValidatorAddr=cosmosvaloper1lzhlnpahvznwfv4jmay2tgaha5kmz5qx4cuznf&dstValidatorAddr=cosmosvaloper1vq8tw77kp8lvxq9u3c8eeln9zymn68rng8pgt4" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "redelegation_responses": [
    {
      "redelegation": {
        "delegator_address": "cosmos1thfntksw0d35n2tkr0k8v54fr8wxtxwxl2c56e",
        "validator_src_address": "cosmosvaloper1lzhlnpahvznwfv4jmay2tgaha5kmz5qx4cuznf",
        "validator_dst_address": "cosmosvaloper1vq8tw77kp8lvxq9u3c8eeln9zymn68rng8pgt4",
        "entries": null
      },
      "entries": [
        {
          "redelegation_entry": {
            "creation_height": 151523,
            "completion_time": "2021-11-09T06:03:25.640682116Z",
            "initial_balance": "200000000",
            "shares_dst": "200000000.000000000000000000"
          },
          "balance": "200000000"
        }
      ]
    }
  ],
  "pagination": null
}
```

### DelegatorUnbondingDelegations

The `DelegatorUnbondingDelegations` REST endpoint queries all unbonding delegations of a given delegator address.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/unbonding_delegations
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1nxv42u3lv642q0fuzu2qmrku27zgut3n3z7lll/unbonding_delegations" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1nxv42u3lv642q0fuzu2qmrku27zgut3n3z7lll",
      "validator_address": "cosmosvaloper1e7mvqlz50ch6gw4yjfemsc069wfre4qwmw53kq",
      "entries": [
        {
          "creation_height": "2442278",
          "completion_time": "2021-10-12T10:59:03.797335857Z",
          "initial_balance": "50000000000",
          "balance": "50000000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### DelegatorValidators

The `DelegatorValidators` REST endpoint queries all validators information for given delegator address.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/validators
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1xwazl8ftks4gn00y5x3c47auquc62ssune9ppv/validators" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "validators": [
    {
      "operator_address": "cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuvynw64",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "5v4n3px3PkfNnKflSgepDnsMQR1hiNXnqOC11Y72/PQ="
      },
      "jailed": false,
      "status": "BOND_STATUS_BONDED",
      "tokens": "21592843799",
      "delegator_shares": "21592843799.000000000000000000",
      "description": {
        "moniker": "jabbey",
        "identity": "",
        "website": "https://twitter.com/JoeAbbey",
        "security_contact": "",
        "details": "just another dad in the cosmos"
      },
      "unbonding_height": "0",
      "unbonding_time": "1970-01-01T00:00:00Z",
      "commission": {
        "commission_rates": {
          "rate": "0.100000000000000000",
          "max_rate": "0.200000000000000000",
          "max_change_rate": "0.100000000000000000"
        },
        "update_time": "2021-10-09T19:03:54.984821705Z"
      },
      "min_self_delegation": "1"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### DelegatorValidator

The `DelegatorValidator` REST endpoint queries validator information for given delegator validator pair.

```bash
/cosmos/staking/v1beta1/delegators/{delegatorAddr}/validators/{validatorAddr}
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/delegators/cosmos1xwazl8ftks4gn00y5x3c47auquc62ssune9ppv/validators/cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuvynw64" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "validator": {
    "operator_address": "cosmosvaloper1xwazl8ftks4gn00y5x3c47auquc62ssuvynw64",
    "consensus_pubkey": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "5v4n3px3PkfNnKflSgepDnsMQR1hiNXnqOC11Y72/PQ="
    },
    "jailed": false,
    "status": "BOND_STATUS_BONDED",
    "tokens": "21592843799",
    "delegator_shares": "21592843799.000000000000000000",
    "description": {
      "moniker": "jabbey",
      "identity": "",
      "website": "https://twitter.com/JoeAbbey",
      "security_contact": "",
      "details": "just another dad in the cosmos"
    },
    "unbonding_height": "0",
    "unbonding_time": "1970-01-01T00:00:00Z",
    "commission": {
      "commission_rates": {
        "rate": "0.100000000000000000",
        "max_rate": "0.200000000000000000",
        "max_change_rate": "0.100000000000000000"
      },
      "update_time": "2021-10-09T19:03:54.984821705Z"
    },
    "min_self_delegation": "1"
  }
}
```

### HistoricalInfo

The `HistoricalInfo` REST endpoint queries the historical information for given height.

```bash
/cosmos/staking/v1beta1/historical_info/{height}
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/historical_info/153332" -H  "accept: application/json"
```

Example Output:

```bash
{
  "hist": {
    "header": {
      "version": {
        "block": "11",
        "app": "0"
      },
      "chain_id": "cosmos-1",
      "height": "153332",
      "time": "2021-10-12T09:05:35.062230221Z",
      "last_block_id": {
        "hash": "NX8HevR5khb7H6NGKva+jVz7cyf0skF1CrcY9A0s+d8=",
        "part_set_header": {
          "total": 1,
          "hash": "zLQ2FiKM5tooL3BInt+VVfgzjlBXfq0Hc8Iux/xrhdg="
        }
      },
      "last_commit_hash": "P6IJrK8vSqU3dGEyRHnAFocoDGja0bn9euLuy09s350=",
      "data_hash": "eUd+6acHWrNXYju8Js449RJ99lOYOs16KpqQl4SMrEM=",
      "validators_hash": "mB4pravvMsJKgi+g8aYdSeNlt0kPjnRFyvtAQtaxcfw=",
      "next_validators_hash": "mB4pravvMsJKgi+g8aYdSeNlt0kPjnRFyvtAQtaxcfw=",
      "consensus_hash": "BICRvH3cKD93v7+R1zxE2ljD34qcvIZ0Bdi389qtoi8=",
      "app_hash": "fuELArKRK+CptnZ8tu54h6xEleSWenHNmqC84W866fU=",
      "last_results_hash": "p/BPexV4LxAzlVcPRvW+lomgXb6Yze8YLIQUo/4Kdgc=",
      "evidence_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
      "proposer_address": "G0MeY8xQx7ooOsni8KE/3R/Ib3Q="
    },
    "valset": [
      {
        "operator_address": "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcqcnylw",
        "consensus_pubkey": {
          "@type": "/cosmos.crypto.ed25519.PubKey",
          "key": "/O7BtNW0pafwfvomgR4ZnfldwPXiFfJs9mHg3gwfv5Q="
        },
        "jailed": false,
        "status": "BOND_STATUS_BONDED",
        "tokens": "1416521659632",
        "delegator_shares": "1416521659632.000000000000000000",
        "description": {
          "moniker": "SG-1",
          "identity": "48608633F99D1B60",
          "website": "https://sg-1.online",
          "security_contact": "",
          "details": "SG-1 - your favorite validator on cosmos. We offer 100% Soft Slash protection."
        },
        "unbonding_height": "0",
        "unbonding_time": "1970-01-01T00:00:00Z",
        "commission": {
          "commission_rates": {
            "rate": "0.037500000000000000",
            "max_rate": "0.200000000000000000",
            "max_change_rate": "0.030000000000000000"
          },
          "update_time": "2021-10-01T15:00:00Z"
        },
        "min_self_delegation": "1"
      },
      {
        "operator_address": "cosmosvaloper1t8ehvswxjfn3ejzkjtntcyrqwvmvuknzmvtaaa",
        "consensus_pubkey": {
          "@type": "/cosmos.crypto.ed25519.PubKey",
          "key": "uExZyjNLtr2+FFIhNDAMcQ8+yTrqE7ygYTsI7khkA5Y="
        },
        "jailed": false,
        "status": "BOND_STATUS_BONDED",
        "tokens": "1348298958808",
        "delegator_shares": "1348298958808.000000000000000000",
        "description": {
          "moniker": "Cosmostation",
          "identity": "AE4C403A6E7AA1AC",
          "website": "https://www.cosmostation.io",
          "security_contact": "admin@stamper.network",
          "details": "Cosmostation validator node. Delegate your tokens and Start Earning Staking Rewards"
        },
        "unbonding_height": "0",
        "unbonding_time": "1970-01-01T00:00:00Z",
        "commission": {
          "commission_rates": {
            "rate": "0.050000000000000000",
            "max_rate": "1.000000000000000000",
            "max_change_rate": "0.200000000000000000"
          },
          "update_time": "2021-10-01T15:06:38.821314287Z"
        },
        "min_self_delegation": "1"
      }
    ]
  }
}
```

### Parameters

The `Parameters` REST endpoint queries the staking parameters.

```bash
/cosmos/staking/v1beta1/params
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/params" -H  "accept: application/json"
```

Example Output:

```bash
{
  "params": {
    "unbonding_time": "2419200s",
    "max_validators": 100,
    "max_entries": 7,
    "historical_entries": 10000,
    "bond_denom": "stake"
  }
}
```

### Pool

The `Pool` REST endpoint queries the pool information.

```bash
/cosmos/staking/v1beta1/pool
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/pool" -H  "accept: application/json"
```

Example Output:

```bash
{
  "pool": {
    "not_bonded_tokens": "432805737458",
    "bonded_tokens": "15783637712645"
  }
}
```

### Validators

The `Validators` REST endpoint queries all validators that match the given status.

```bash
/cosmos/staking/v1beta1/validators
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/validators" -H  "accept: application/json"
```

Example Output:

```bash
{
  "validators": [
    {
      "operator_address": "cosmosvaloper1q3jsx9dpfhtyqqgetwpe5tmk8f0ms5qywje8tw",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "N7BPyek2aKuNZ0N/8YsrqSDhGZmgVaYUBuddY8pwKaE="
      },
      "jailed": false,
      "status": "BOND_STATUS_BONDED",
      "tokens": "383301887799",
      "delegator_shares": "383301887799.000000000000000000",
      "description": {
        "moniker": "SmartNodes",
        "identity": "D372724899D1EDC8",
        "website": "https://smartnodes.co",
        "security_contact": "",
        "details": "Earn Rewards with Crypto Staking & Node Deployment"
      },
      "unbonding_height": "0",
      "unbonding_time": "1970-01-01T00:00:00Z",
      "commission": {
        "commission_rates": {
          "rate": "0.050000000000000000",
          "max_rate": "0.200000000000000000",
          "max_change_rate": "0.100000000000000000"
        },
        "update_time": "2021-10-01T15:51:31.596618510Z"
      },
      "min_self_delegation": "1"
    },
    {
      "operator_address": "cosmosvaloper1q5ku90atkhktze83j9xjaks2p7uruag5zp6wt7",
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "GDNpuKDmCg9GnhnsiU4fCWktuGUemjNfvpCZiqoRIYA="
      },
      "jailed": false,
      "status": "BOND_STATUS_UNBONDING",
      "tokens": "1017819654",
      "delegator_shares": "1017819654.000000000000000000",
      "description": {
        "moniker": "Noderunners",
        "identity": "812E82D12FEA3493",
        "website": "http://noderunners.biz",
        "security_contact": "info@noderunners.biz",
        "details": "Noderunners is a professional validator in POS networks. We have a huge node running experience, reliable soft and hardware. Our commissions are always low, our support to delegators is always full. Stake with us and start receiving your cosmos rewards now!"
      },
      "unbonding_height": "147302",
      "unbonding_time": "2021-11-08T22:58:53.718662452Z",
      "commission": {
        "commission_rates": {
          "rate": "0.050000000000000000",
          "max_rate": "0.200000000000000000",
          "max_change_rate": "0.100000000000000000"
        },
        "update_time": "2021-10-04T18:02:21.446645619Z"
      },
      "min_self_delegation": "1"
    }
  ],
  "pagination": {
    "next_key": "FONDBFkE4tEEf7yxWWKOD49jC2NK",
    "total": "2"
  }
}
```

### Validator

The `Validator` REST endpoint queries validator information for given validator address.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "validator": {
    "operator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
    "consensus_pubkey": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "sIiexdJdYWn27+7iUHQJDnkp63gq/rzUq1Y+fxoGjXc="
    },
    "jailed": false,
    "status": "BOND_STATUS_BONDED",
    "tokens": "33027900000",
    "delegator_shares": "33027900000.000000000000000000",
    "description": {
      "moniker": "Witval",
      "identity": "51468B615127273A",
      "website": "",
      "security_contact": "",
      "details": "Witval is the validator arm from Vitwit. Vitwit is into software consulting and services business since 2015. We are working closely with Cosmos ecosystem since 2018. We are also building tools for the ecosystem, Aneka is our explorer for the cosmos ecosystem."
    },
    "unbonding_height": "0",
    "unbonding_time": "1970-01-01T00:00:00Z",
    "commission": {
      "commission_rates": {
        "rate": "0.050000000000000000",
        "max_rate": "0.200000000000000000",
        "max_change_rate": "0.020000000000000000"
      },
      "update_time": "2021-10-01T19:24:52.663191049Z"
    },
    "min_self_delegation": "1"
  }
}
```

### ValidatorDelegations

The `ValidatorDelegations` REST endpoint queries delegate information for given validator.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/delegations
```

Example:

```bash
curl -X GET "http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q/delegations" -H  "accept: application/json"
```

Example Output:

```bash
{
  "delegation_responses": [
    {
      "delegation": {
        "delegator_address": "cosmos190g5j8aszqhvtg7cprmev8xcxs6csra7xnk3n3",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "31000000000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "31000000000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos1ddle9tczl87gsvmeva3c48nenyng4n56qwq4ee",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "628470000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "628470000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos10fdvkczl76m040smd33lh9xn9j0cf26kk4s2nw",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "838120000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "838120000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos1n8f5fknsv2yt7a8u6nrx30zqy7lu9jfm0t5lq8",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "500000000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "500000000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos16msryt3fqlxtvsy8u5ay7wv2p8mglfg9hrek2e",
        "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
        "shares": "61310000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "61310000"
      }
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "5"
  }
}
```

### Delegation

The `Delegation` REST endpoint queries delegate information for given validator delegator pair.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/delegations/{delegatorAddr}
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q/delegations/cosmos1n8f5fknsv2yt7a8u6nrx30zqy7lu9jfm0t5lq8" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "delegation_response": {
    "delegation": {
      "delegator_address": "cosmos1n8f5fknsv2yt7a8u6nrx30zqy7lu9jfm0t5lq8",
      "validator_address": "cosmosvaloper16msryt3fqlxtvsy8u5ay7wv2p8mglfg9g70e3q",
      "shares": "500000000.000000000000000000"
    },
    "balance": {
      "denom": "stake",
      "amount": "500000000"
    }
  }
}
```

### UnbondingDelegation

The `UnbondingDelegation` REST endpoint queries unbonding information for given validator delegator pair.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/delegations/{delegatorAddr}/unbonding_delegation
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu/delegations/cosmos1ze2ye5u5k3qdlexvt2e0nn0508p04094ya0qpm/unbonding_delegation" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "unbond": {
    "delegator_address": "cosmos1ze2ye5u5k3qdlexvt2e0nn0508p04094ya0qpm",
    "validator_address": "cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu",
    "entries": [
      {
        "creation_height": "153687",
        "completion_time": "2021-11-09T09:41:18.352401903Z",
        "initial_balance": "525111",
        "balance": "525111"
      }
    ]
  }
}
```

### ValidatorUnbondingDelegations

The `ValidatorUnbondingDelegations` REST endpoint queries unbonding delegations of a validator.

```bash
/cosmos/staking/v1beta1/validators/{validatorAddr}/unbonding_delegations
```

Example:

```bash
curl -X GET \
"http://localhost:1317/cosmos/staking/v1beta1/validators/cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu/unbonding_delegations" \
-H  "accept: application/json"
```

Example Output:

```bash
{
  "unbonding_responses": [
    {
      "delegator_address": "cosmos1q9snn84jfrd9ge8t46kdcggpe58dua82vnj7uy",
      "validator_address": "cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu",
      "entries": [
        {
          "creation_height": "90998",
          "completion_time": "2021-11-05T00:14:37.005841058Z",
          "initial_balance": "24000000",
          "balance": "24000000"
        }
      ]
    },
    {
      "delegator_address": "cosmos1qf36e6wmq9h4twhdvs6pyq9qcaeu7ye0s3dqq2",
      "validator_address": "cosmosvaloper13v4spsah85ps4vtrw07vzea37gq5la5gktlkeu",
      "entries": [
        {
          "creation_height": "47478",
          "completion_time": "2021-11-01T22:47:26.714116854Z",
          "initial_balance": "8000000",
          "balance": "8000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```
