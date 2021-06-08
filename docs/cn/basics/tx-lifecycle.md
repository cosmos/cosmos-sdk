# Transaction 的生命周期

本文档描述了 Transaction 从创建到提交的生命周期，Transaction 的定义在[其他文档](https://docs.cosmos.network/master/core/transactions.html)中有详细描述，后文中 Transaction 将统一被称为`Tx`。

## 创建

### Transaction 的创建

命令行界面是主要的应用程序界面之一，`Tx` 可以由用户输入[以下命令](https://docs.cosmos.network/master/core/cli.html)来创建，其中 `[command]` 是 `Tx` 的类型，`[args]` 是相关参数，`[flags]` 是相关配置例如 gas price：

```bash
[appname] tx [command] [args] [flags]
```

此命令将自动**创建** `Tx`，使用帐户的私钥对其进行**签名**，并将其**广播**到其他节点。

创建 `Tx` 有一些必需的和可选的参数，其中 `--from` 指定该 `Tx` 的发起[账户](https://docs.cosmos.network/master/basics/accounts.html)，例如一个发送代币的`Tx`，则将从 `from` 指定的账户提取资产。

#### Gas 和 Fee

此外，用户可以使用这几个[参数](https://docs.cosmos.network/master/core/cli.html)来表明他们愿意支付多少 [fee](https://docs.cosmos.network/master/basics/gas-fees.html)：

- `--gas` 指的是 [gas](https://docs.cosmos.network/master/basics/gas-fees.html) 的数量，gas 代表 `Tx` 消耗的计算资源，需要消耗多少 gas 取决于具体的 `Tx`，在 `Tx` 执行之前无法被精确计算出来，但可以通过在 `--gas` 后带上参数 `auto` 来进行估算。
- `--gas-adjustment`（可选）可用于适当的增加 `gas`，以避免其被低估。例如，用户可以将 `gas-adjustment` 设为 1.5，那么被指定的 gas 将是被估算 gas 的 1.5 倍。
- `--gas-prices` 指定用户愿意为每单位 gas 支付多少 fee，可以是一种或多种代币。例如，`--gas-prices=0.025uatom, 0.025upho` 就表明用户愿意为每单位的 gas 支付 0.025uatom 和 0.025upho。
- `--fees` 指定用户总共愿意支付的 fee。

所支付 fee 的最终价值等于 gas 的数量乘以 gas 的价格。换句话说，`fees = ceil(gas * gasPrices)`。由于可以使用 gas 价格来计算 fee，也可以使用 fee 来计算 gas 价格，因此用户仅指定两者之一即可。

随后，验证者通过将给定的或计算出的 `gas-prices` 与他们本地的 `min-gas-prices` 进行比较，来决定是否在其区块中写入该 `Tx`。如果 `gas-prices` 不够高，该 `Tx` 将被拒绝，因此鼓励用户支付更多 fee。

#### CLI 示例

应用程序的用户可以在其 CLI 中输入以下命令，用来生成一个将 1000uatom 从 `senderAddress` 发送到 `recipientAddress` 的 `Tx`，该命令指定了用户愿意支付的 gas（其中 gas 数量为自动估算的 1.5 倍，每单位 gas 价格为 0.025uatom）。

```bash
appcli tx send <recipientAddress> 1000uatom --from <senderAddress> --gas auto --gas-adjustment 1.5 --gas-prices 0.025uatom
```

#### 其他的 Transaction 创建方法

命令行是与应用程序进行交互的一种简便方法，但是 `Tx` 也可以使用 [REST interface](https://docs.cosmos.network/master/core/grpc_rest.html) 或应用程序开发人员定义的某些其他入口点来创建命令行。从用户的角度来看，交互方式取决于他们正在使用的是页面还是钱包（例如， `Tx` 使用 [Lunie.io](https://lunie.io/#/) 创建并使用 Ledger Nano S 对其进行签名）。

## 添加到交易池

每个全节点（Tendermint 节点）接收到 `Tx` 后都会发送一个名为 `CheckTx` 的 [ABCI message](https://tendermint.com/docs/spec/abci/abci.html#messages)，用来检查 `Tx` 的有效性，`CheckTx` 会返回 `abci.ResponseCheckTx`。
如果 `Tx` 通过检查，则将其保留在节点的 [**交易池**](https://tendermint.com/docs/tendermint-core/mempool.html#mempool)（每个节点唯一的内存事务池）中等待出块，`Tx` 如果被发现无效，诚实的节点将丢弃该 `Tx`。在达成共识之前，节点会不断检查传入的 `Tx` 并将其广播出去。

### 检查的类型

全节点在 `CheckTx` 期间对 `Tx` 先执行无状态检查，然后进行有状态检查，目的是尽早识别并拒绝无效 `Tx`，以免浪费计算资源。

**_无状态检查_**不需要知道节点的状态，即轻客户端或脱机节点都可以检查，因此计算开销较小。无状态检查包括确保地址不为空、强制使用非负数、以及定义中指定的其他逻辑。

**_状态检查_**根据提交的状态验证 `Tx` 和 `Message`。例如，检查相关值是否存在并能够进行交易，账户是否有足够的资产，发送方是否被授权或拥有正确的交易所有权。在任何时刻，由于不同的原因，全节点通常具有应用程序内部状态的[多种版本](https://docs.cosmos.network/master/core/baseapp.html#volatile-states)。例如，节点将在验证 `Tx` 的过程中执行状态更改，但仍需要最后的提交状态才能响应请求，节点不能使用未提交的状态更改来响应请求。

为了验证 `Tx`，全节点调用的 `CheckTx` 包括无状态检查和有状态检查，进一步的验证将在 [`DeliverTx`](#delivertx) 阶段的后期进行。其中 `CheckTx` 从对 `Tx` 进行解码开始。

### 解码

当 `Tx` 从应用程序底层的共识引擎（如 Tendermint）被接收时，其仍处于 `[]byte`[编码](https://docs.cosmos.network/master/core/encoding.html) 形式，需要将其解码才能进行操作。随后，[`runTx`](https://docs.cosmos.network/master/core/baseapp.html#runtx-and-runmsgs) 函数会被调用，并以 `runTxModeCheck` 模式运行，这意味着该函数将运行所有检查，但是会在执行 `Message` 和写入状态更改之前退出。

### ValidateBasic

[Message](https://docs.cosmos.network/master/core/transactions.html#messages) 是由 module 的开发者实现的 `Msg` 接口中的一个方法。它应包括基本的**无状态**完整性检查。例如，如果 `Message` 是要将代币从一个账户发送到另一个账户，则 `ValidateBasic` 会检查账户是否存在，并确认账户中代币金额为正，但不需要了解状态，例如帐户余额。

### AnteHandler

[`AnteHandler`](https://docs.cosmos.network/master/basics/gas-fees.html#antehandler)是可选的，但每个应用程序都需要定义。`AnteHandler` 使用副本为特定的 `Tx` 执行有限的检查，副本可以使对 `Tx` 进行状态检查时无需修改最后的提交状态，如果执行失败，还可以还原为原始状态。

例如，[`auth`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth/spec) 模块的 `AnteHandler` 检查并增加序列号，检查签名和帐号，并从 `Tx` 的第一个签名者中扣除费用，这个过程中所有状态更改都使用 `checkState`

### Gas

[`Context`](https://docs.cosmos.network/master/core/context.html) 相当于`GasMeter`，会计算出在 `Tx` 的执行过程中多少 `gas` 已被使用。用户提供的 `Tx` 所需的 `gas` 数量称为 `GasWanted`。`Tx` 在实际执行过程中消耗的 `gas` 被称为`GasConsumed`，如果 `GasConsumed` 超过 `GasWanted`，将停止执行，并且对状态副本的修改不会被提交。否则，`CheckTx` 设置 `GasUsed` 等于 `GasConsumed` 并返回结果。在计算完 gas 和 fee 后，验证器节点检查用户指定的值 `gas-prices` 是否小于其本地定义的值 `min-gas-prices`。

### 丢弃或添加到交易池

如果在 `CheckTx` 期间有任何失败，`Tx` 将被丢弃，并且 `Tx` 的生命周期结束。如果 `CheckTx` 成功，则 `Tx` 将被广播到其他节点，并会被添加到交易池，以便成为待出区块中的候选 `Tx`。

**交易池**保存所有全节点可见的 `Tx`，全节点会将其最近的 `Tx` 保留在**交易池缓存**中，作为防止重放攻击的第一道防线。理想情况下，`mempool.cache_size` 的大小足以容纳整个交易池中的所有 `Tx`。如果交易池缓存太小而无法跟踪所有 `Tx`，`CheckTx` 会识别出并拒绝重放的 `Tx`。

现有的预防措施包括 fee 和`序列号`计数器，用来区分重放 `Tx` 和相同的 `Tx`。如果攻击者尝试向某个节点发送多个相同的 `Tx`，则保留交易池缓存的完整节点将拒绝相同的 `Tx`，而不是在所有 `Tx` 上运行 `CheckTx`。如果 `Tx` 有不同的`序列号`，攻击者会因为需要支付费用而取消攻击。

验证器节点与全节点一样，保留一个交易池以防止重放攻击，但它也用作出块过程中未经验证的交易池。请注意，即使 `Tx` 在此阶段通过了所有检查，仍然可能会被发现无效，因为 `CheckTx` 没有完全验证 `Tx`（`CheckTx` 实际上并未执行 `message`）。

## 写入区块

共识是验证者节点就接受哪些 `Tx` 达成协议的过程，它是**反复进行**的。每个回合都始于出块节点创建一个包含最近 `Tx` 的区块，并由验证者节点（具有投票权的特殊全节点）负责达成共识，同意接受该区块或出一个空块。验证者节点执行共识算法，例如[Tendermint BFT](https://tendermint.com/docs/spec/consensus/consensus.html#terms)，调用 ABCI 请求确认 `Tx`，从而达成共识。

达成共识的第一步是**区块提案**，共识算法从验证者节点中选择一个出块节点来创建和提议一个区块，用来写入 `Tx`，`Tx` 必须在该提议者的交易池中。

## 状态变更

共识的下一步是执行 `Tx` 以完全验证它们，所有的全节点收到出块节点广播的区块并调用 ABCI 函数[`BeginBlock`](https://docs.cosmos.network/master/basics/app-anatomy.html#beginblocker-and-endblocker)，`DeliverTx`，和 [`EndBlock`](https://docs.cosmos.network/master/basics/app-anatomy.html#beginblocker-and-endblocker)。全节点在本地运行的每个过程将产生一个明确的结果，因为 `message` 的状态转换是确定性的，并且 `Tx` 在提案中有明确的顺序。

```
         -----------------------------
        |Receive Block Proposal|
         -----------------------------
                         |
                         v
         -----------------------------
        |         BeginBlock         |
         -----------------------------
                         |
                         v
        -----------------------------
        |      DeliverTx(tx0)      |
        |      DeliverTx(tx1)      |
        |      DeliverTx(tx2)      |
        |      DeliverTx(tx3)      |
        |               .                 |
        |               .                 |
        |               .                 |
        -----------------------------
                         |
                         v
        -----------------------------
        |          EndBlock          |
        -----------------------------
                         |
                         v
        -----------------------------
        |          Consensus        |
        -----------------------------
                         |
                         v
        -----------------------------
        |           Commit          |
        -----------------------------
```

### DeliverTx

[`baseapp`](https://docs.cosmos.network/master/core/baseapp.html) 中定义的 ABCI 函数 `DeliverTx` 会执行大部分状态转换，`DeliverTx` 会针对共识中确定的顺序，对块中的每个 `Tx` 按顺序运行。`DeliverTx` 几乎和 `CheckTx` 相同，但是会以 deliver 模式调用[`runTx`](../core/baseapp.md#runtx)函数而不是 check 模式。全节点不使用 `checkState`，而是使用 `deliverState`。

- **解码：** 由于 `DeliverTx` 是通过 ABCI 调用的，因此 `Tx` 会以 `[]byte` 的形式被接收。节点首先会对 `Tx` 进行解码，然后在 `runTxModeDeliver` 中调用 `runTx`，`runTx` 除了会执行 `CheckTx` 中的检查外，还会执行 `Tx` 和并写入状态的变化。

- **检查：** 全节点会再次调用 `validateBasicMsgs` 和 `AnteHandler`。之所以进行第二次检查，是因为在 `Tx` 进交易池的过程中，可能没有相同的 `Tx`，但恶意出块节点的区块可能包括无效 `Tx`。但是这次检查特殊的地方在于，`AnteHandler` 不会将 `gas-prices` 与节点的 `min-gas-prices` 比较，因为每个节点的 `min-gas-prices` 可能都不同，这样比较的话可能会产生不确定的结果。

- **路由和 Handler：** `CheckTx` 退出后，`DeliverTx` 会继续运行 [`runMsgs`](https://docs.cosmos.network/master/core/baseapp.html#runtx-and-runmsgs) 来执行 `Tx` 中的每个 `Msg`。由于 `Tx` 可能具有来自不同模块的 `message`，因此 `baseapp` 需要知道哪个模块可以找到适当的 `Handler`。因此，`路由`通过[模块管理器](https://docs.cosmos.network/master/building-modules/module-manager.html)来检索路由名称并找到对应的[`Handler`](https://docs.cosmos.network/master/building-modules/handler.html)。

- **Handler：** `handler` 是用来执行 `Tx` 中的每个 `message`，并且使状态转换到从而保持 `deliverTxState`。`handler` 在 `Msg` 的模块中定义，并写入模块中的适当存储区。

- **Gas：** 在 `Tx` 被传递的过程中，`GasMeter` 是用来记录有多少 gas 被使用，如果执行完成，`GasUsed` 会被赋值并返回 `abci.ResponseDeliverTx`。如果由于 `BlockGasMeter` 或者 `GasMeter` 耗尽或其他原因导致执行中断，程序则会报出相应的错误。

如果由于 `Tx` 无效或 `GasMeter` 用尽而导致任何状态更改失败，`Tx` 的处理将被终止，并且所有状态更改都将还原。区块提案中无效的 `Tx` 会导致验证者节点拒绝该区块并投票给空块。

### 提交

最后一步是让节点提交区块和状态更改，在重跑了区块中所有的 `Tx` 之后，验证者节点会验证区块的签名以最终确认它。不是验证者节点的全节点不参与共识（即无法投票），而是接受投票信息以了解是否应提交状态更改。

当收到足够的验证者票数（2/3+的加权票数）时，完整的节点将提交一个新的区块，以添加到区块链网络中并最终确定应用程序层中的状态转换。此过程会生成一个新的状态根，用作状态转换的默克尔证明。应用程序使用从[Baseapp](https://docs.cosmos.network/master/core/baseapp.html)继承的 ABCI 方法[`Commit`](https://docs.cosmos.network/master/core/baseapp.html#commit)，`Commit` 通过将 `deliverState` 写入应用程序的内部状态来同步所有的状态转换。提交状态更改后，`checkState` 从最近提交的状态重新开始，并将 `deliverState` 重置为空以保持一致并反映更改。

请注意，并非所有区块都具有相同数量的 `Tx`，并且共识可能会导致一个空块。在公共区块链网络中，验证者可能是**拜占庭恶意**的，这可能会阻止将 `Tx` 提交到区块链中。可能的恶意行为包括出块节点将某个 `Tx` 排除在区块链之外，或者投票反对某个出块节点。

至此，`Tx`的生命周期结束，节点已验证其有效性，并提交了这些更改。`Tx`本身，以 `[]byte` 的形式被存储在区块上进入了区块链网络。

## 下一节

了解 [accounts](./accounts.md)
