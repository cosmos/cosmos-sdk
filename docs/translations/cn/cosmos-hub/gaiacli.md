# Gaia客户端

## Gaia CLI

`gaiacli`是一个工具，使您能够与 Cosmos Hub 网络中的节点进行交互，无论您是否自己运行它。 让我们恰当的设置它。 要安装它，请按照[安装步骤](./ installation.md)进行安装。

### 配置 gaiacli

设置`gaiacli`的主要命令如下：

```bash
gaiacli config <flag> <value>
```

该命令能为每个标志设置默认值。

首先，设置要连接的全节点的地址：

```bash
gaiacli config node <host>:<port

# example: gaiacli config node https://77.87.106.33:26657
```

如果您运行自己的全节点，只需使用`tcp://localhost:26657`地址即可。

然后，让我们设置`--trust-node`标志的默认值：

```bash
gaiacli config trust-node true

# Set to true if you trust the full-node you are connecting to, false otherwise
```

最后，设置我们想要与之交互链的`chain-id`：

```bash
gaiacli config chain-id cosmoshub-2
```

### Key

#### Key类型

有如下类型的key：

+ `cosmos` 
	+ 从通过`gaiacli keys add`生成的账户私钥中产生
	+ 用于接收资金
	+ 例如 `cosmos15h6vd5f0wqps26zjlwrc6chah08ryu4hzzdwhc`
+ `cosmosvaloper`
	+ 用于关联一个验证人和其操作者
	+ 用于发起staking操作命令
	+ 例如 `cosmosvaloper1carzvgq3e6y3z5kz5y6gxp3wpy3qdrv928vyah`
+ `cosmospub`
	+ 从通过`gaiacli keys add`生成的账户私钥中产生
	+ 例如 `cosmospub1zcjduc3q7fu03jnlu2xpl75s2nkt7krm6grh4cc5aqth73v0zwmea25wj2hsqhlqzm`
+ `cosmosvalconspub`
	+ 在使用`gaiad init`创建节点时生成
	+ 使用`gaiad tendermint show-validator`获得该值
	+ 例如 `cosmosvalconspub1zcjduepq0ms2738680y72v44tfyqm3c9ppduku8fs6sr73fx7m666sjztznqzp2emf`

#### 生成 Key

你需要一个帐户的私钥公钥对（分别称作`sk`，`pk`）才能接收资金，发送交易，绑定交易等等。

生成一个新的*secp256k1*密钥：

```bash
gaiacli keys add <account_name>
```

接下来，你必须创建一个密码来保护磁盘上的密钥。上述命令的输出将包含种子短语。建议将种子短语保存在安全的地方，以便在忘记密码的情况下，最终可以使用以下命令从种子短语重新生成密钥：

```bash
gaiacli keys add --recover
```

如果你检查你的私钥，你会看到`<account_name>` :

```bash
gaiacli keys show <account_name>
```

通过下面的命令查看验证人操作者的地址：

```bash
gaiacli keys show <account_name> --bech=val
```

你可以查看你所有的可以使用的密钥：

```bash
gaiacli keys list
```

查看你节点的验证人公钥：

```bash
gaiad tendermint show-validator
```

请注意，这是Tendermint的签名密钥，而不是你在委托交易中使用的操作员密钥。

::: danger Warning
我们强烈建议不要对多个密钥使用相同的密码。Tendermint 团队和 Interchain Foundation 将不承担资金损失的责任。
:::

#### 生成多签公钥
你可以生成一个多签公钥并将其打印：

```bash
gaiacli keys add --multisig=name1,name2,name3[...] --multisig-threshold=K new_key_name
```

`K`是将要对多签公钥发起的交易进行签名的最小私钥数。

`--multisig`标识必须包含要将组合成一个公钥的那些子公钥的名称，该公钥将在本地数据库中生成并存储为`new_key_name`。通过`--multisig`提供的所有名称必须已存在于本地数据库中。除非设置了`--nosort`标识，否则在命令行上提供密钥的顺序无关紧要，即以下命令生成两个相同的密钥：

```bash
gaiacli keys add --multisig=foo,bar,baz --multisig-threshold=2 multisig_address
gaiacli keys add --multisig=baz,foo,bar --multisig-threshold=2 multisig_address
```

多签地址也可以在运行中生成并通过以下命令打印：

```bash
gaiacli keys show --multisig-threshold K name1 name2 name3 [...]
```

有关如何生成多签帐户，使用其签名和广播多签交易的详细信息，请参阅[多签交易](#多签交易)

### Tx 广播

在广播交易时，`gaiacli`接受`--broadcast-mode`标识。 这个标识的值可以是`sync`（默认值）、`async`或`block`，其中`sync`使客户端返回 CheckTx 响应，`async`使客户端立即返回，而`block`使得 客户端等待 tx 被提交（或超时）。

值得注意的是，在大多数情况下**不**应该使用`block`模式。 这是因为广播可以超时但是 tx 仍然可能存在在块中，这可能导致很多不良结果。 因此，最好使用`sync`或`async`并通过 tx hash 查询以确定 tx 何时包含在块中。

### Fees 和 Gas

每笔交易可能会提供 fees 或 gas price，但不能同时提供。

验证人可以配置最低 gas price（多币种的），并且在决定它们是否能被包含在区块中的`CheckTx`期间使用改值，其中 `gasPrices >= minGasPrices`。请注意，你的交易必须提供大于或等于验证人要求的任何接受币种的费用。

**注意**：有了这样的机制，验证人可能会开始在 mempool 中通过 gasPrice 来优先处理某些 txs，因此提供更高 fee 或 gas price可能会产生更高的tx优先级。

比如：

```bash
gaiacli tx send ... --fees=50000uatom
```

或：

```bash
gaiacli tx send ... --gas-prices=0.025uatom
```


### 账户

#### 获取 Token

获取token的最佳方式是通过[Cosmos测试网水龙头](https://faucetcosmos.network)。如果水龙头对你不生效，尝试在[#cosmos-validator](https://riot.im/app/#/room/#cosmos-validators:matrix.org)上向人索要。水龙头需要你打算用于抵押股权的`cosmos`开头的地址。

#### 查询账户余额

在你的地址收到token后，你可以通过以下命令查看账户的余额：

```bash
gaiacli query account <account_cosmos>
```

::: warning Note
当你查询余额为零的帐户时，你将收到以下错误：`No account with address <account_cosmos> was found in the state.` 如果你在节点与区块链完全同步之前就查询，也会发生这种情况。这些都很正常。
:::

#### 发送 Token

你可以通过如下命令从一个账户发送资金到另一个账户：

```bash
gaiacli tx send <destination_cosmos> 10faucetToken \
  --chain-id=<chain_id> \
  --from=<key_name> 
```

::: warning Note
`--amount`标识接收格式：`--amount=<value|coin_name>`
:::

::: tip Note
你可能希望通过`--gas`标识限制交易可以消耗的最大燃料。如果你通过`--gas=auto`，将在执行交易前自动估gas。gas估算可能是不准确的，因为状态变化可能发生在模拟结束和交易的实际执行之间，因此在原始估计之上应用调整以确保能够成功地广播交易。可以通过`--gas-adjustment`标识控制调整，其默认值为1.0。
:::

现在，查看源账户和目标账户的更新后的余额：

```bash
gaiacli query account <account_cosmos>
gaiacli query account <destination_cosmos>
```

你还可以使用`--block`标识查询在特定高度区块下你的余额：

```bash
gaiacli query account <account_cosmos> --block=<block_height>
```

你可以通过在命令行中附加`--dry-run`标识来模拟交易而不实际广播它：

```bash
gaiacli tx send <destination_cosmosaccaddr> 10faucetToken \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --dry-run
```

此外，你可以通过将`--generate-only`附加到命令行参数列表来构建交易并将其JSON格式打印到STDOUT：

```bash
gaiacli tx send <destination_cosmosaccaddr> 10faucetToken \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --generate-only > unsignedSendTx.json
```

```bash
gaiacli tx sign \
  --chain-id=<chain_id> \
  --from=<key_name>
  unsignedSendTx.json > signedSendTx.json
```

::: tip Note
标识 `--generate-only` 只能在访问本地 keybase 时使用。
:::

你可以通过下面的命令验证交易的签名：

```bash
gaiacli tx sign --validate-signatures signedSendTx.json
```

你可以将由JSON文件提供的已签名的交易广播至指定节点：

```bash
gaiacli tx broadcast --node=<node> signedSendTx.json
```

### 查询交易

#### 匹配一组tag

你可以使用交易搜索命令查询与每个交易上添加的特定`标签集`匹配的交易。

每个标签都由`<tag>:<value>`形式的键值对形成。还可以使用`＆`符号组合标签来查询更具体的结果。

使用`标签`查询交易的命令如下：

```bash
gaiacli query txs --tags='<tag>:<value>'
```

使用多个`标签`:

```bash
gaiacli query txs --tags='<tag1>:<value1>&<tag2>:<value2>'
```

通过`page`和`limit`来实现分页:

```bash
gaiacli query txs --tags='<tag>:<value>' --page=1 --limit=20
```

::: tip 注意

action标签始终等于相关message的`Type()`函数返回的消息类型。

你可以在每个SDK的模块中找到目前的标签列表：
- [Common tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/types/tags.go#L57-L63)
- [Staking tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/staking/tags/tags.go#L8-L24)
- [Governance tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/gov/tags/tags.go#L8-L22)
- [Slashing tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/slashing/handler.go#L52)
- [Distribution tags](https://github.com/cosmos/cosmos-sdk/blob/develop/x/distribution/tags/tags.go#L8-L17)
- [Bank tags](https://github.com/cosmos/cosmos-sdk/blob/d1e76221d8e28824bb4791cb4ad8662d2ae9051e/x/bank/keeper.go#L193-L206)
  :::

#### 匹配一笔交易的hash

你一可以通过指定hash值查询该笔交易：

```bash
gaiacli query tx [hash]
```

### Slashing

#### Unjailing

将你入狱的验证人释放出狱：

```bash
gaiacli tx slashing unjail --from <validator-operator-addr>
```

#### Signing Info

检索一个验证人的签名信息：

```bash
gaiacli query slashing signing-info <validator-pubkey>
```


#### 查询参数

你可以查询当前的slashing参数：

```bash
gaiacli query slashing params
```

### Staking

#### 设置一个验证人

有关如何设置验证人候选者的完整指南，请参阅[验证人设置]()章节

#### 向一个验证人委托

一旦主网上线，你可以把`atom`委托给一个验证人。这些委托人可以收到部分验证人的收益。阅读[Cosmos Token Model](https://github.com/cosmos/cosmos/raw/master/Cosmos_Token_Model.pdf)了解更多信息。

#### 查询验证人

你可以查询指定链的验证人：

```bash
gaiacli query staking validators
```

如果你想要获得单个验证人的信息，你可以使用下面的命令：

```bash
gaiacli query staking validator <account_cosmosval>
```

#### 绑定 Token

在Cosmos Hub主网中，我们绑定`uatom`，`1atom = 1000000uatom`。你可以把token绑定在一个测试网验证人节点上（即委托）：

```bash
gaiacli tx staking delegate \
  --amount=10000000uatom \
  --validator=<validator> \
  --from=<key_name> \
  --chain-id=<chain_id>
```

`<validator>`是你要委托的验证人的操作者地址。如果你运行的是本地testnet，可以通过以下方式找到：

```bash
gaiacli keys show [name] --bech val
```

其中`[name]`是初始化`gaiad`时指定的键的名称。

虽然token是绑定的，但它们与网络中的所有其他绑定的token汇集在一起。验证人和委托人获得一定比例的股权，这些股权等于他们在这个资产池中的抵押。

#### 查询委托

一旦提交了一笔对验证人的委托，你可以使用下面的命令查看委托详情：

```bash
gaiacli query staking delegation <delegator_addr> <validator_addr>
```

或者你想查看所有当前的委托：

```bash
gaiacli query staking delegations <delegator_addr>
```

你还可以通过添加`--height`标识来获取先前的委托状态。

#### 解绑 Token
如果出于一些原因验证人行为异常，或者你想解绑一定数量的token，请使用以下命令。

```bash
gaiacli tx staking unbond \
  <validator_addr> \
  10atom \
  --from=<key_name> \
  --chain-id=<chain_id>
```

经过解绑期后，解绑自动完成。

#### 查询Unbonding-Delegations

一旦你开始了一笔unbonding-delegation，你可以使用以下命令查看信息：

```bash
gaiacli query staking unbonding-delegation <delegator_addr> <validator_addr>
```

或者你可以查看当前你所有的unbonding-delegation:

```bash
gaiacli query staking unbonding-delegations <account_cosmos>
```

此外，你可以从特定验证人获取所有unbonding-delegation：

```bash
gaiacli query staking unbonding-delegations-from <account_cosmosval>
```

要获取指定区块时的unbonding-delegation状态，请尝试添加`--height`标识。

#### 重新委托token

重新授权是一种委托类型，允许你将非流动token从一个验证人上绑定到另一个验证人：

```bash
gaiacli tx staking redelegate \
  <src-validator-operator-addr> \
  <dst-validator-operator-addr> \
  10atom \
  --from=<key_name> \
  --chain-id=<chain_id>
```

这里，你还可以使用`shares-amount`或`shares-fraction`标识重新委托。

经过解绑期后，重新委托自动完成。

#### 查询重新委托

开始重新授权后，你可以使用以下命令查看其信息：

```bash
gaiacli query staking redelegation <delegator_addr> <src_val_addr> <dst_val_addr>
```

或者，如果你可以检查所有当前的unbonding-delegation：

```bash
gaiacli query staking redelegations <account_cosmos>
```

此外，你可以查询某个特定验证人的所有转出的重新绑定：

```bash
gaiacli query staking redelegations-from <account_cosmosval>
```

添加`--height`标识来查询之前某个特定区块的redelegation。

#### 查询参数

参数定义了staking的高级参数。你可以使用以下方法获取：

```bash
gaiacli query staking params
```

使用上面的命令，你将获得以下值：

+ unbonding时间
+ 验证人的最大数量
+ 用于抵押的币种

所有这些值都将通过对一个`ParameterChange`提案的`governance`流程进行更新。

#### 查询抵押池
一个抵押池定义了当前状态的动态参数。你可以通过以下命令查询：

```bash
gaiacli query staking pool
```

使用`pool`命令，你将获得以下值：

+ 未绑定和已绑定的token
+ token总量
+ 当前的年度通货膨胀率以及上次发生通货膨胀的区块
+ 最后记录的绑定股权

#### 查询对验证人的绑定
你可以查询对某个验证人的所有绑定：

```bash
gaiacli query delegations-to <account_cosmosval>
```

### 治理

治理是Cosmos Hub的用户可以就软件升级，主网的参数或自定义文本提案并达成共识的过程。这是通过对提案进行投票来完成的，提案将由主要网络上的`Atom`持有者提交。

关于投票过程的一些考虑因素：
+ 投票由绑定`Atom`的持有者以1个绑定的`Atom`对应1票方式投出
+ 委托人不投票的话会将票权继承给其验证人
+ **验证人必须对每个提案进行投票**。如果验证人未对提案进行投票，则会对其进行削减处罚。
+ 投票期结束时（主网上是2周）统计投票。每个地址可以多次投票以更新其`Option`值（每次支付交易费用），只有最后一次投票将被视为有效。
+ 选民可以选择`Yes`，`No`，`NoWithVeto`和`Abstain`选项。在投票结束时，如果`( YesVotes / ( YesVotes + NoVotes + NoWithVetoVotes ) ) > 1/2`且`( NoWithVetoVotes / ( YesVotes + NoVotes + NoWithVetoVotes )) < 1/3`提案通过，否则就拒绝。

有关治理流程及其工作原理的更多信息，请查看Governance模块[规范]()。

#### 创建一个治理提案

要创建治理提案，您必须提交初始抵押以及标题和说明。治理之外的其它模块可以实现自己的提议类型和处理程序（例如：参数更改），其中治理模块本身支持`Text`提议。治理之外的任何模块都将命令绑定在`submit-proposal`上。

提交一个文本类型的提案：

```bash
gaiacli tx gov submit-proposal \
  --title=<title> \
  --description=<description> \
  --type="Text" \
  --deposit="1000000uatom" \
  --from=<name> \
  --chain-id=<chain_id>
```

您也可以直接通过`--proposal`指向包含提案的 JSON 文件。

要提交更改参数的提案，您必须提供提案文件，因为其内容对 CLI 输入不太友好：

```bash
gaiacli tx gov submit-proposal param-change <path/to/proposal.json> \
  --from=<name> \
  --chain-id=<chain_id>
```

其中`proposal.json`包含以下内容：

```json
{
  "title": "Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 105
    }
  ],
  "deposit": [
    {
      "denom": "stake",
      "amount": "10000000"
    }
  ]
}
```

::: danger Warning

Currently parameter changes are _evaluated_ but not _validated_, so it is very important
that any `value` change is valid (ie. correct type and within bounds) for its
respective parameter, eg. `MaxValidators` should be an integer and not a decimal.

Proper vetting of a parameter change proposal should prevent this from happening
(no deposits should occur during the governance process), but it should be noted
regardless. 

目前，参数更改已经过*评估*但未*经过验证*，因此`value`对于其相应参数，任何更改都是有效的（即正确类型和边界内）非常重要，例如 `MaxValidators` 应该是整数而不是小数。

正确审查参数变更提案应该可以防止这种情况发生（在治理过程中不会发生抵押），但无论如何都应该注意。

:::

::: tip Note

目前不支持`SoftwareUpgrade`，因为它没有实现，目前与`Text`提议的语义没有区别。

:::

#### 查询提案

一旦创建，你就可以查询提案的信息：

```bash
gaiacli query gov proposal <proposal_id>
```

或者查询所有的有效提案：

```bash
gaiacli query gov proposals
```

你还可以使用`voter`或`depositor`标识来过滤查询提案。

要查询特定提案的提议人：

```bash
gaiacli query gov proposer <proposal_id>
```

#### 增加存入金

为了将提案广播到网络，存入的金额必须高于`minDeposit`值（初始值：`10steak`）。如果你之前创建的提案不符合此要求，你仍可以增加存入的总金额以激活它。达到最低存入金后，提案进入投票期：

```bash
gaiacli tx gov deposit <proposal_id> "10000000uatom" \
  --from=<name> \
  --chain-id=<chain_id>
```

> 注意：达到`MaxDepositPeriod`后，将删除不符合此要求的提案。


#### 查询存入金
创建新提案后，你可以查询提交其所有存款：

```bash
gaiacli query gov deposits <proposal_id>
```

你还可以查询特定地址提交的存入金：

```bash
gaiacli query gov deposit <proposal_id> <depositor_address>
```

#### 投票给一个提案

在提案的存入金达到`MinDeposit`后，投票期将开放。抵押了`Atom`的持有人可以投票：

```bash
gaiacli tx gov vote <proposal_id> <Yes/No/NoWithVeto/Abstain> \
  --from=<name> \
  --chain-id=<chain_id>
```

#### 查询投票

使用您刚才提交的参数检查投票：

```bash
gaiacli query gov vote <proposal_id> <voter_address>
```

你还可以查询提交给所有此前投给指定提案的投票：

```bash
gaiacli query gov votes <proposal_id>
```

#### 查询提案的计票结果

要检查指定提案的当前计票，你可以使用`tally`命令：

```bash
gaiacli query gov tally <proposal_id>
```

#### 查询治理参数

要检查当前的治理参数，请运行：

```bash
gaiacli query gov params
```

查询运行的治理参数的子集：

```bash
gaiacli query gov param voting
gaiacli query gov param tallying
gaiacli query gov param deposit
```

### 费用分配

#### 查询分配参数

查询当前的分配参数：

```bash
gaiacli query distr params
```

#### 查询

查询当前未结算的（未提取）的奖励：

```bash
gaiacli query distr outstanding-rewards
```

#### 查询验证人佣金

查询对一个验证人的未结算的佣金：

```bash
gaiacli query distr commission <validator_address>
```

#### 查询验证人的削减处罚

查询一个验证人的处罚历史记录：

```bash
gaiacli query distr slashes <validator_address> <start_height> <end_height>
```

#### 查询委托人奖励

查询某笔委托当前的奖励（如果要取回）：

```bash
gaiacli query distr rewards <delegator_address> <validator_address>
```

#### 查询所有的委托人奖励

要查询委托人的所有当前奖励（如果要取回），请运行：

```bash
gaiacli query distr rewards <delegator_address>
```

### 多签交易

多签交易需要多个私钥的签名。因此，从多签账户生成和签署交易涉及有关各方之间的合作。密钥持有者的任何一方都可以发起多签，并且至少要有其中一方需要将其他账户的公钥导入到本地的数据库并生成多签公钥来完成和广播该笔交易。

例如，给定包含密钥`p1`，`p2`和`p3`的多签密钥，每个密钥由不同方持有，持有`p1`的用户将需要导入`p2`和`p3`的公钥以生成多签帐户公钥：

```bash
gaiacli keys add \
  p2 \
  --pubkey=cosmospub1addwnpepqtd28uwa0yxtwal5223qqr5aqf5y57tc7kk7z8qd4zplrdlk5ez5kdnlrj4

gaiacli keys add \
  p3 \
  --pubkey=cosmospub1addwnpepqgj04jpm9wrdml5qnss9kjxkmxzywuklnkj0g3a3f8l5wx9z4ennz84ym5t

gaiacli keys add \
  p1p2p3 \
  --multisig-threshold=2 \
  --multisig=p1,p2,p3
```

已存储新的多签公钥`p1p2p3`，其地址将用作多签交易的签名者：

```bash
gaiacli keys show --address p1p2p3
```

您还可以通过查看 key 的 JSON 输出或增加`--show-multisig`标识来查看multisig阈值，pubkey构成和相应的权重：

```bash
gaiacli keys show p1p2p3 -o json

gaiacli keys show p1p2p3 --show-multisig
```

创建多签交易的第一步是使用上面创建的多签地址初始化：

```bash
gaiacli tx send cosmos1570v2fq3twt0f0x02vhxpuzc9jc4yl30q2qned 10000000uatom \
  --from=<multisig_address> \
  --generate-only > unsignedTx.json
```

`unsignedTx.json`文件包含以JSON编码的未签署交易。`p1`现在可以使用自己的私钥对交易进行签名：

```bash
gaiacli tx sign \
  unsignedTx.json \
  --multisig=<multisig_address> \
  --from=p1 \
  --output-document=p1signature.json 
```

生成签名后，`p1`将`unsignedTx.json`和`p1signature.json`都发送到`p2`或`p3`，然后`p2`或`p3`将生成它们各自的签名:

```bash
gaiacli tx sign \
  unsignedTx.json \
  --multisig=<multisig_address> \
  --from=p2 \
  --output-document=p2signature.json
```

`p1p2p3` is a 2-of-3 multisig key, therefore one additional signature is sufficient. Any the key holders can now generate the multisig transaction by combining the required signature files:

p1p2p3` 是 2-of-3 多签key，因此一个的签名就足够了。 现在，任何密钥持有者都可以通过组合所需的签名文件来生成多签交易：

```bash
gaiacli tx multisign \
  unsignedTx.json \
  p1p2p3 \
  p1signature.json p2signature.json > signedTx.json
```

现在可以把交易发送给节点：

```bash
gaiacli tx broadcast signedTx.json
```

## shell 自动补全脚本

可以通过完全命令生成主流的UNIX shell解释器（如`Bash`和`Zsh`）的`completion`命令，该命令可用于`gaiad`和`gaiacli`。

如果要生成Bash完成脚本，请运行以下命令：

```bash
gaiad completion > gaiad_completion
gaiacli completion > gaiacli_completion
```

如果要生成Zsh完成脚本，请运行以下命令：

```bash
gaiad completion --zsh > gaiad_completion
gaiacli completion --zsh > gaiacli_completion
```

::: tip Note
在大多数UNIX系统上，可以在`.bashrc`或`.bash_profile`中加载此类脚本以启用Bash自动完成：

```bash
echo '. gaiad_completion' >> ~/.bashrc
echo '. gaiacli_completion' >> ~/.bashrc
```

有关如何启用shell自动完成的信息，请参阅操作系统提供的解释器用户手册。
:::