# SDK 应用程序架构

## 状态机

区块链应用的核心是[具有最终确定性的复制状态机](https://en.wikipedia.org/wiki/State_machine_replication)。

状态机是计算机科学概念，一台机器可以具有多个状态，但在任何给定时间只有一个`状态`，其描述了系统的当前状态，及触发这些状态转变的交易（译者注：也对应数据库中事务的概念）。

给定一个状态S和交易T，状态机会返回一个新的状态S'。

```
+--------+                 +--------+
|        |                 |        |
|   S    +---------------->+   S'   |
|        |    apply(T)     |        |
+--------+                 +--------+
```

实际上，交易以区块的形式打包在一起以提高过程的效率。给定状态S和包含交易的区块B，状态机将返回新状态S'。

```
+--------+                              +--------+
|        |                              |        |
|   S    +----------------------------> |   S'   |
|        |   For each T in B: apply(T)  |        |
+--------+                              +--------+
```

在区块链上下文环境中，状态机是确定性的。这意味着如果你从一个给定的状态开始，重放相同顺序的交易，将始终以相同的最终状态结束。

Cosmos SDK 为你提供了最大的灵活性用以定义自身应用程序的状态、交易类型和状态转换函数。在接下来的章节中会更深入细致的描述如何使用 SDK 来构建状态机。但首先，让我们看看状态机是如何使用 **Tendermint** 进行复制的。

### Tendermint

作为一个开发者，你只需要使用 Cosmos-SDK 定义状态机，而[Tendermint](https://tendermint.com/docs/introduction/introduction.html)将会为你处理网络层的状态复制。

```
                ^  +-------------------------------+  ^
                |  |                               |  |   通过 Cosmos SDK 构建
                |  |         状态机 = 应用（层）      |  |
                |  |                               |  v
                |  +-------------------------------+
                |  |                               |  ^
       链节点    |  |             共识层             |  |
                |  |                               |  |
                |  +-------------------------------+  |   Tendermint Core
                |  |                               |  |
                |  |              网络层            |  |
                |  |                               |  |
                v  +-------------------------------+  v
```

Tendermint是一个与应用程序无关的引擎，负责处理区块链的*网络层*和*共识层*。实际上，这意味着Tendermint负责传播和排序交易字节。Tendermint Core 依赖于拜占庭容错（BFT）算法来达成交易顺序的共识。要深入了解Tendermint，可点击[这里](https://tendermint.com/docs/introduction/what-is-tendermint.html)。

Tendermint一致性算法通过一组称为*验证人*的特殊节点一起运作。验证人负责向区块链添加交易区块。对于任何给定的区块，有一组验证人V。通过算法选择V中的验证人A作为下一个区块的提议人。如果超过三分之二的V签署了[prevote](https://tendermint.com/docs/spec/consensus/consensus.html#state-machine-spec)和[precommit](https://tendermint.com/docs/spec/consensus/consensus.html#state-machine-spec)，并且区块包含的所有交易都是有效的，则该区块被认为是有效的。验证人集合可以通过状态机中编写的规则进行更改。要深入了解算法，[点击](https://tendermint.com/docs/introduction/what-is-tendermint.html#consensus-overview)。

Cosmos SDK 应用程序的主要部分是一个区块链服务后台（daemon），它在每个网络节点的本地运行。如果验证人集合中三分之一以下的是拜占庭（即恶意的），则每个节点在同时查询状态时应获得相同的结果。


## ABCI

Tendermint通过名为[ABCI](https://github.com/tendermint/tendermint/tree/master/abci)的接口将交易从网络层传递给应用程序，因此应用程序必须要实现 ABCI 。

```
+---------------------+
|                     |
|         应用         |
|                     |
+--------+---+--------+
         ^   |
         |   | ABCI
         |   v
+--------+---+--------+
|                     |
|                     |
|     Tendermint      |
|                     |
|                     |
+---------------------+
```

注意，Tendermint 仅处理交易字节。它不知道这些字节究竟是什么意思。Tendermint 所做的只是对交易确定性地排序。赋予这些字节意义是应用程序的工作。Tendermint通过ABCI将交易字节传递给应用程序，并期望返回代码以知晓消息是否成功。

以下是ABCI中最重要的消息类型：

- `CheckTx` : 当 Tendermint Core 收到交易时，如果符合一些的基本的要求会将其传递给应用程序。`Checkx` 用于保护全节点的交易池免受垃圾邮件的侵害。一个名为“Ante Handler”的特殊处理器用于执行一系列验证步骤，例如检查手续费用是否足够和验证签名是否合法。如果交易有效，则将交易添加到[交易池（mempool）](https://tendermint.com/docs/spec/reactors/mempool/functionality.html#mempool-functionality)中并广播到对等节点。注意， `CheckTx` 不会处理交易（即不会对修改状态），因为它们尚未包含在区块中。
- `DeliverTx` : 当 Tendermint Core 接收到[有效区块](https://tendermint.com/docs/spec/blockchain/blockchain.html#validation)时，块中的每条交易都将通过 `DeliverTx `传递给应用程序进行处理。正是在这一阶段发生了状态转换。“Ante Handler”也将连同实际处理交易中每条消息的handler一起再次执行。
- `BeginBlock`/`EndBlock` : 无论区块是否包含交易，这两个消息都将在每个区块的开头和结尾执行。触发自动的逻辑执行是很有用的。过程中要足够小心，因为计算成本高昂的循环运算可能会减慢区块链的速度，甚至发生无限循环引起区块链本身停滞。


有关 ABCI 方法和类型的详细介绍请[点击](https://tendermint.com/docs/spec/abci/abci.html#overview)。

在 Tendermint 上构建的任何应用程序都需要实现ABCI接口，来同本地的底层 Tendermint 引擎进行通信。幸运的是，你不用必需实现ABCI接口。Cosmos SDK以[baseapp](./sdk-design.md#baseapp)的形式提供了样板实现。

### 接下来，让我们学习[SDK的高级设计原则](./sdk-design.md)