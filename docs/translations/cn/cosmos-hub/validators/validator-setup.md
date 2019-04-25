# 在主网上运行一个验证人

::: 提示
加入主网所需的信息(`genesis.json`和种子节点)在[`lauch` repo](https://github.com/cosmos/launch/tree/master/latest)中可以找到。
:::

在启动你验证人节点前，确定你已经完成了[启动全节点](../join-mainnet.md)教程。

## 什么是验证人?

[验证人](./overview.md)负责通过投票来向区块链提交新区块。如果验证人不可访问或者对多个相同高度的区块签名，将会遭受到削减处罚。如果变得不可用或者在同一高度上签名，则会被削减。请阅读有关Sentry节点架构的信息，以保护您的节点免受DDOS攻击并确保高可用性。请阅读[哨兵节点网络架构]()来保护你的节点免于DDOS攻击并保证高的可访问性。

::: 警告
如果你想要成为Cosmos Hub主网的验证人，你应该[安全研究](./security.md)。
:::

如果你已经[启动了一个全节点](../join-mainnet.md)，可以跳过下一节的内容。

## 创建你的验证人

你的`cosmosvalconspub`可以用于通过抵押token来创建一个新的验证人。你可以通过运行下面的命令来查看你的验证人公钥：

```bash
gaiad tendermint show-validator
```

使用下面的命令创建你的验证人：

::: 注意
不要使用多于你所拥有的`uatom`!
:::

```bash
gaiacli tx staking create-validator \
  --amount=1000000uatom \
  --pubkey=$(gaiad tendermint show-validator) \
  --moniker="choose a moniker" \
  --chain-id=<chain_id> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
  --gas-prices="0.025uatom" \
  --from=<key_name>
```

::: 提示
在指定commission参数时，`commission-max-change-rate`用于度量`commission-rate`的百分比点数的变化。比如，1%到2%增长了100%，但反映到`commission-rate`上只有1个百分点。
:::

::: 提示
如果没有指定，`consensus_pubkey`将默认为`gaiad tendermint show-validator`命令的输出。`key_name`是将用于对交易进行签名的私钥的名称。
:::

你可以在第三方区块链浏览器上确定你是否处于验证人行列。

## 以初始验证人的形式加入到genesis文件

::: 警告
这一节内容只针对想要在Cosmos Hub主网启动前就作为初始验证人身份的节点。如果主网已经启动，请跳过这一节。
:::

如果你想作为初始验证人被写入到genesis.json文件，你需要证明你在创世状态中有一些权益代币，创建一个（或多个）交易以将股权与你的验证人地址联系起来，并将此交易包含在genesis文件中。

你的`cosmosvalconspub`可以用于通过抵押token来创建一个新的验证人。运行如下命令来获取你的验证人节点公钥：

```bash
gaiad tendermint show-validator
```

然后执行`gaiad gentx`命令:

::: 提示
`gentx`是持有self-delegation的JSON文件。所有的创世交易会被`创世协调员`收集起来验证并初始化成一个`genesis.json`
:::

::: 注意
不要使用多于你所拥有的`uatom`!
:::

```bash
gaiad gentx \
  --amount <amount_of_delegation_uatom> \
  --commission-rate <commission_rate> \
  --commission-max-rate <commission_max_rate> \
  --commission-max-change-rate <commission_max_change_rate> \
  --pubkey <consensus_pubkey> \
  --name <key_name>
```

::: 提示
在指定佣金相关的参数时，`commission-max-change-rate`用于标识`commission-rate`每日变动的最大百分点数。比如从1%到2%按比率是增长了100%，但只增加了1个百分点。
:::

你可以提交你的`gentx`到[launch repository](https://github.com/cosmos/launch). 这些`gentx`将会组成最终的genesis.json.

## 编辑验证人的描述信息

你可以编辑验证人的公开说明。此信息用于标识你的验证人节点，委托人将根据此信息来决定要委托的验证人节点。确保为下面的每个标识提供输入，否则该字段将默认为空（ `--moniker`默认为机器名称）。

<key_name>指定你要编辑的验证人。如果你选择不包含此标识，记住必须要含有--from标识来指定你要更新的验证人。

`--identity`可用于验证和Keybase或UPort这样的系统一起验证身份。与Keybase一起使用时，`--identity`应使用由一个[keybase.io](https://keybase.io/)帐户生成的16位字符串。它是一种加密安全的方法，可以跨多个在线网络验证您的身份。 Keybase API允许我们检索你的Keybase头像。这是你可以在验证人配置文件中添加徽标的方法。

```bash
gaiacli tx staking edit-validator
  --moniker="choose a moniker" \
  --website="https://cosmos.network" \
  --identity=6A0D65E29A4CBC8E \
  --details="To infinity and beyond!" \
  --chain-id=<chain_id> \
  --gas="auto" \
  --gas-prices="0.025uatom" \
  --from=<key_name> \
  --commission-rate="0.10"
```

**注意** : `commission-rate`的值必须符合如下的不变量检查：

+ 必须在 0 和 验证人的`commission-max-rate` 之间
+ 不得超过 验证人的`commission-max-change-rate`, 该参数标识**每日**最大的百分点变化数。也就是，一个验证人在`commission-max-change-rate`的界限内每日一次可调整的最大佣金变化。


## 查看验证人的描述信息

通过该命令查看验证人的描述信息:

```bash
gaiacli query staking validator <account_cosmos>
```

## 跟踪验证人的签名信息

你可以通过`signing-info`命令跟踪过往的验证人签名：

```bash
gaiacli query slashing signing-info <validator-pubkey>\
  --chain-id=<chain_id>
```

## unjail验证人

当验证人因停机而"jailed"(入狱)时，你必须用节点操作人帐户提交一笔`Unjail`交易，使其再次能够获得区块提交的奖励（奖励多少取决于分区的fee分配）。

```bash
gaiacli tx slashing unjail \
	--from=<key_name> \
	--chain-id=<chain_id>
```

## 确认你的验证人节点正在运行

如果下面的命令返回有内容就证明你的验证人正处于活跃状态:

```bash
gaiacli query tendermint-validator-set | grep "$(gaiad tendermint show-validator)"
```

你必须要在[区块浏览器](https://explorecosmos.network/validators)中看见你的验证人节点信息。你可以在`~/.gaiad/config/priv_validator.json`文件中找到`bech32`编码格式的`address`。

::: warning 注意
为了能进入验证人集合，你的权重必须超过第100名的验证人。
:::

## 常见问题

### 问题 #1 : 我的验证人的`voting_power: 0`

你的验证人已经是jailed状态。如果验证人在最近`10000`个区块中有超过`500`个区块没有进行投票，或者被发现双签，就会被jail掉。

如果被因为掉线而遭到jail，你可以重获你的投票股权以重回验证人队伍。首先，如果`gaiad`没有运行，请再次启动：

```bash
gaiad start
```

等待你的全节点追赶上最新的区块高度。然后，运行如下命令。接着，你可以[unjail你的验证人]()。

最后，检查你的验证人看看投票股权是否恢复：

```bash
gaiacli status
```

你可能会注意到你的投票权比之前要少。这是由于你的下线受到的削减处罚！


### 问题 #2 : 我的`gaiad`由于`too many open files`而崩溃

Linux可以打开的默认文件数（每个进程）是1024。已知`gaiad`可以打开超过1024个文件。这会导致进程崩溃。快速修复运行`ulimit -n 4096`（增加允许的打开文件数）来快速修复，然后使用`gaiad start`重新启动进程。如果你使用`systemd`或其他进程管理器来启动`gaiad`，则可能需要在该级别进行一些配置。解决此问题的示例`systemd`文件如下：

```toml
# /etc/systemd/system/gaiad.service
[Unit]
Description=Cosmos Gaia Node
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu
ExecStart=/home/ubuntu/go/bin/gaiad start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```
