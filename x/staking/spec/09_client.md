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

The `historical-info` command allows users to query historical info at given height.

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
      your Juno rewards now!
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

## gRPC

A user can query the `staking` module using gRPC endpoints.

