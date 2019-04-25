# 创世（Genesis）文件

本文档解释了 Cosmos Hub 主网的 genesis 文件是如何构建的。 它还解释了如何为自己的`gaia` testnet 创建一个 genesis 文件。

请注意，您可以通过运行以下命令为您自己的 testnet 生成默认的 genesis 文件：

```bash
gaiad init <moniker> --chain-id <chain-id>
```

genesis 文件存储在 `~/.gaiad/config/genesis.toml`.

## 什么是创世文件

genesis 文件是一个 JSON 文件，用于定义区块链的初始状态。 它可以看作是区块链的高度“0”。 高度为“1”的第一个块将引用 genesis 文件作为其父级。

genesis 文件中定义的状态包含所有必要的信息，如初始令牌分配、创建时间、默认参数等。 我们来分别描述这些信息。

## Genesis 时间和链ID

`genesis_time`定义在 genesis 文件的顶部。 它是一个“UTC”时间戳，指示区块链何时启动。 此时，创世记验证人应该上线并开始参与共识过程。 当超过2/3的生成验证人（通过投票权加权）在线时，区块链启动。

```json
"genesis_time": "2019-03-13T17:00:00.000000000Z",
```

`chain_id`是您的链的唯一标识符。 它有助于区分使用相同版本的软件的不同链。

```json
"chain_id": "cosmoshub-1",
```

## 共识参数

接下来，创世文件定义共识参数。 共识参数覆盖与共识层相关的所有参数，`gaia` 的共识层是 `Tendermint`。 我们来看看这些参数：

- `block`
  - `max_bytes`: 每个块的最大字节数。
  - `max_gas`: 每个块的最大 gas 数量。 该区块中包含的每笔交易都会消耗一些 gas，包含在一个区块内的交易所使用的总 gas 不能超出。
- `evidence`
  - `max_age`: 证据（evidence）是一种证明，表明验证者在同一高度（同一轮）签署了两个不同的区块。 这是一种明显的恶意行为，会在状态机层受到惩罚。 `max_age`定义**块**的最大数量，在经过`max_age`块之后证据不再有效。
- `validator`
  - `pub_key_types`: 可被验证人接受的公钥类型 (例如`ed25519`, `secp256k1`, ...) ，目前仅支持`ed25519`。

```json
"consensus_params": {
    "block_size": {
      "max_bytes": "150000",
      "max_gas": "1500000"
    },
    "evidence": {
      "max_age": "1000000"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    }
  },
```

## 应用程序状态

应用程序状态定义了状态机的初始状态。

### 创世账号

在本节中，定义了初始分配的 Token。 可以通过直接编辑 genesis 文件手动添加帐户，但也可以使用以下命令：

```bash
// Example: gaiad add-genesis-account cosmos1qs8tnw2t8l6amtzvdemnnsq9dzk0ag0z37gh3h 10000000uatom

gaiad add-genesis-account <account-address> <amount><denom>
```

这个命令在 `app_state.accounts` 下创建一个条目。

```json
"accounts": [
      {
        "address": "cosmos1qs8tnw2t8l6amtzvdemnnsq9dzk0ag0z37gh3h",
        "coins": [
          {
            "denom": "uatom",
            "amount": "10000000"
          }
        ],
        "sequence_number": "0",
        "account_number": "0",
        "original_vesting": [
          {
            "denom": "uatom",
            "amount": "26306000000"
          }
        ],
        "delegated_free": null,
        "delegated_vesting": null,
        "start_time": "0",
        "end_time": "10000"
      }
]
```

让我们来分别解读这些参数：

- `sequence_number`: 此编号用于计算此帐户发送的交易数。 每次事务包含在块中时它都会递增，并用于防止重放攻击，初始值为“0”。
- `account_number`: 帐户的唯一标识符，它在包含此帐户的首次被打包到块的交易中生成。
- `original_vesting`: 锁仓（Vesting） 由`gaia`原生支持。 您可以定义帐户需要锁仓 token 数量，这些 token 在一定时间之后才能流通。 锁仓中的 token 可用于委托。 默认值为“null”。
- `delegated_free`: 在 vest 过期后可转让的委托 token 数量。在创世文件中，大部分情况是 `null`。
- `delegated_vesting`: 锁仓中的 token 数量。在创世文件中，大部分情况是 `null`。
- `start_time`: vesting 期开始区块高度。创世文件中，大部分情况是`0`。
- `end_time`: vesting 期结束区块高度。如果没有 token 在 vesting 期，这个值是`0`。

### 银行（Bank）

`bank`模块负责 token。在本节中唯一 需要定义的参数是“转账”是否在创世文件启用。

```json
"bank": {
      "send_enabled": false
    }
```

### 权益（Staking）

`staking`模块处理状态机中的大多数 POS 逻辑。 此部分应如下所示：

```json
"staking": {
      "pool": {
        "not_bonded_tokens": "10000000",
        "bonded_tokens": "0"
      },
      "params": {
        "unbonding_time": "1814400000000000",
        "max_validators": 100,
        "max_entries": 7,
        "bond_denom": "uatom"
      },
      "last_total_power": "0",
      "last_validator_powers": null,
      "validators": null,
      "bonds": null,
      "unbonding_delegations": null,
      "redelegations": null,
      "exported": false
    }
```

让我们来分别解读这些参数：

- `pool`
  - `not_bonded_tokens`: 在创世文件中没有绑定（即委托）的 token 数量。 通常情况下，它与权益 token （本例中是 `uatom`）的总供应量相等。
  - `bonded_tokens`: 在创世文件中没绑定的 token 数量，通常是0。
- `params`
  - `unbonding_time`: 以**纳秒**为单位的解绑延迟时间。
  - `max_validators`: 最大验证人节点数量。
  - `max_entries`: 可同时进行解委托、重新委托的最大条目数。
  - `bond_denom`: 权益代币符号。
- `last_total_power`: 总投票权重。在创世文件通常是0（除非创世文件使用了之前的状态）。
- `last_validator_powers`: 最后一个区块的状态中每个验证人的投票权重。在创世文件中通常是 null（除非创世文件使用了之前的状态）。
- `validators`: 最后一个区块中的验证人列表。在创世文件中通常是 null（除非创世文件使用了之前的状态）。
- `bonds`: 最后一个区块中的委托列表。在创世文件中通常是 null（除非创世文件使用了之前的状态）。
- `unbonding_delegations`: 最后一个区块中解绑的委托列表。在创世文件中通常是 null（除非创世文件使用了之前的状态）。
- `redelegations`: 最后一个区块中的重新委托列表。在创世文件中通常是 null（除非创世文件使用了之前的状态）。
- `exported`: 创世文件是否是从之前的状态导出得到的。

### 挖矿（Mint）

`mint`模块管理 token 供应的通胀逻辑。 创世文件中的`mint`部分如下所示：

```json
"mint": {
      "minter": {
        "inflation": "0.070000000000000000",
        "annual_provisions": "0.000000000000000000"
      },
      "params": {
        "mint_denom": "uatom",
        "inflation_rate_change": "0.130000000000000000",
        "inflation_max": "0.200000000000000000",
        "inflation_min": "0.070000000000000000",
        "goal_bonded": "0.670000000000000000",
        "blocks_per_year": "6311520"
      }
    }
```

让我们来分别解读这些参数：

- `minter`
  - `inflation`：总 token 供应量的年化通胀百分比，每周更新。值 “0.070000000000000000” 意味着目标是每年通货膨胀率为“7％”，每周重新计算一次。
  - `annual_provisions`: 每块重新计算。初始值是 `0.000000000000000000`。
- `params`
  - `mint_denom`: 增发权益代币面值，此处是 `uatom`。
  - `inflation_rate_change`: 通胀每年最大变化。 
  - `inflation_max`: 最高通胀水平。
  - `inflation_min`: 最低通胀水平。
  - `goal_bonded`: 目标绑定量占总供应量百分比。如果委托 token 的百分比低于此目标，则通胀率会增加（在`inflation_rate_change`之后），直至达到`inflation_max`。 如果委托 token 的百分比高于此目标，则通胀率会下降（在`inflation_rate_change`之后），直至达到`inflation_min`。
  - `blocks_per_year`: 每年出块量估算。用于计算出块收益中权益 token 的通胀部分（称之为块供给）。

### 分配（Distribution）

`distr`模块处理每个块中发给验证人和委托人的挖矿及手续费的分配逻辑。 创世文件中的`distr`部分如下所示：

```json
    "distr": {
      "fee_pool": {
        "community_pool": null
      },
      "community_tax": "0.020000000000000000",
      "base_proposer_reward": "0.010000000000000000",
      "bonus_proposer_reward": "0.040000000000000000",
      "withdraw_addr_enabled": false,
      "delegator_withdraw_infos": null,
      "previous_proposer": "",
      "outstanding_rewards": null,
      "validator_accumulated_commissions": null,
      "validator_historical_rewards": null,
      "validator_current_rewards": null,
      "delegator_starting_infos": null,
      "validator_slash_events": null
    }
```

让我们来分别解读这些参数：

- `fee_pool`
  - `community_pool`: 用于支付奖励的 token 放在公共池中，它通过治理提案分配。在创世文件中通常是 null。
- `community_tax`: 税率，即交易费和出块收益中需要放入公共池部分的百分比。
- `base_proposer_reward`: 区块提议者在有效区块中收取的交易费用奖励的基础部分。 如果值为`0.010000000000000000`，则1％的费用将转给提议者。
- `bonus_proposer_reward`: 如果预提交取得了 2/3 （该块有效的最小值）的加权投票，他们会获得 `base_proposer_reward` 奖励。  如果预提交获得100％的加权投票，则此奖励线性增加至`bonus_proposer_reward`。
- `withdraw_addr_enabled`: 如果是`true`，委托人可以设置不同的地址来取回他们的奖励。 如果要在创世时禁用转账，则要设置为`false`，因为它可以绕过转账限制。
- `delegator_withdraw_infos`: 委托人收益地址列表。 如果没有从之前的状态导出，一般是`null`。
- `previous_proposer`: 上一个块的提议者，  如果没有从之前的状态导出，则设置为""。
- `outstanding_rewards`: 未付（未提取）奖励。如果没有从之前的状态导出，设置为`null`。
- `validator_accumulated_commission`: 未付（未提取）验证人佣金。如果没有从之前的状态导出，设置为`null`。
- `validator_historical_rewards`: 验证人的历史奖励相关的信息，由`distr`模块用于各种计算。 如果没有从之前的状态导出，设置为`null`。
- `validators_current_rewards`: 验证人的当前奖励相关的信息，由`distr`模块用于各种计算。 如果没有从之前的状态导出，设置为`null`。
- `delegator_starting_infos`: Tracks the previous validator period, the delegation's amount of staking token, and the creation height (to check later on if any slashes have occurred). 跟踪先前的验证人时期，委托的 token 数量和创建高度（稍后检查是否发生了需要惩罚的事件）。  如果没有从之前的状态导出，设置为`null`。
- `validator_slash_events`: Set of information related to the past slashing of validators. Set to `null` if genesis was not exported from previous state. 过往验证人惩罚事件相关的信息集。 如果没有从之前的状态导出，设置为`null`。

### 治理（Governance）

`gov`模块处理所有与治理相关的事务。 `gov`部分的初始状态如下所示：

```json
"gov": {
      "starting_proposal_id": "1",
      "deposits": null,
      "votes": null,
      "proposals": null,
      "deposit_params": {
        "min_deposit": [
          {
            "denom": "uatom",
            "amount": "512000000"
          }
        ],
        "max_deposit_period": "1209600000000000"
      },
      "voting_params": {
        "voting_period": "1209600000000000"
      },
      "tally_params": {
        "quorum": "0.4",
        "threshold": "0.5",
        "veto": "0.334",
        "governance_penalty": "0.0"
      }
    }
```

让我们来分别解读这些参数：

- `starting_proposal_id`: 此参数定义第一个提案的ID，每个提案都由唯一ID标识。
- `deposits`: 每个提案 ID 的保证金列表。如果没有从之前的状态导出，设置为`null`。
- `votes`: 每个提案 ID 的投票列表。 如果没有从之前的状态导出，设置为`null`。
- `proposals`: 所有提案列表。如果没有从之前的状态导出，设置为`null`。
- `deposit_params`
  - `min_deposit`: 使提案进入投票期的最小抵押数量，如果提供了多种面值，满足其一即可。
  - `max_deposit_period`: 最长抵押等待时间（单位**纳秒**），之后就不能再进行抵押了。
- `voting_params`
  - `voting_period`: 投票期时长（单位**纳秒**）。
- `tally_params`
  - `quorum`: 提议生效所需的投票数占总抵押数的百分比。
  - `threshold`: 提议生效所需 `YES` 票最小百分比。
  - `veto`: 若提议生效，`NO_WITH_VETO` 票最大百分比.
  - `governance_penalty`: 对未给特定提案进行投票的验证人的处罚。

### 惩罚（Slashing ）

The `slashing` module handles the logic to slash delegators if their validator misbehave. The `slashing` section in genesis looks as follows:

`slashing`模块处理对验证人行为不当的惩罚逻辑。 创世文件中的`slashing`部分如下：

```json
"slashing": {
      "params": {
        "max_evidence_age": "1814400000000000",
        "signed_blocks_window": "10000",
        "min_signed_per_window": "0.050000000000000000",
        "downtime_jail_duration": "600000000000",
        "slash_fraction_double_sign": "0.050000000000000000",
        "slash_fraction_downtime": "0.000100000000000000"
      },
      "signing_infos": {},
      "missed_blocks": {}
    }
```

让我们来分别解读这些参数：

- `params`
  - `max_evidence_age`: 证据最长有效期，单位 **纳秒**。
  - `signed_blocks_window`: 用于识别离线验证人节点的块滑动窗口。
  - `min_signed_per_window`: 在滑动窗口中预提交的数量少于此值，认为验证人节点离线。
  - `downtime_jail_duration`: 如果验证人离线时间超过此处设定的**纳秒**数，验证人节点将被关小黑屋。
  - `slash_fraction_double_sign`: 验证人节点双签时，需缴纳罚金占总委托数量的百分比。
  - `slash_fraction_downtime`: 验证人节点离线时，需缴纳罚金占总委托数量的百分比。
- `signing_infos`:`slashing` 模块所需的每个验证人节点的各种信息。如果没有从之前的状态导出，设置为`{}`。
- `missed_blocks`: `slashing` 模块所需的与丢块相关的各种信息。如果没有从之前的状态导出，设置为`{}`。

### 创世交易（Genesis Transactions）

默认情况下，genesis文件不包含任何`gentxs`。 `gentx`是一种交易，在创世文件中的将`accounts`下的 token 委托给验证人节点，本质上就是在创世时创建验证人。 在`genesis_time`之后，一旦有超过 2/3 的验证人（加权投票）作为有效`gentx`的接收者上线，该链就会启动。

可以手动将`gentx`添加到genesis文件，或通过以下命令：

```bash
gaiad collect-gentxs
```

此命令将存储在`~/.gaiad/config/gentx`中的所有`gentxs`添加到genesis文件中。 要创建创世纪交易，请单击[此处](./validators/validator-setup.md#participation-in-genesis-as-a-validator)。