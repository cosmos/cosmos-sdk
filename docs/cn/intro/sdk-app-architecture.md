# 区块链架构

## 状态机

区块链的核心是[复制确定状态机](https://en.wikipedia.org/wiki/State_machine_replication)（replicated deterministic state machine）。

状态机是计算机科学领域的一个概念，即一台机器可以具有多个状态，但在任意给定时刻只具有一个确定的状态。我们用`state`描述系统当前状态，`transactions`触发状态转换。

给定一个状态S和Transaction T，状态机会返回新的状态S'。

```
+--------+                 +--------+
|        |                 |        |
|   S    +---------------->+   S'   |
|        |    apply(T)     |        |
+--------+                 +--------+
```

在实际中，Transaction集会被打包进区块中，以让处理过程更加高效。给定一个状态S和一个包含Transaction集 B的区块，状态机就会返回新的状态S'。

```
+--------+                              +--------+
|        |                              |        |
|   S    +----------------------------> |   S'   |
|        |   For each T in B: apply(T)  |        |
+--------+                              +--------+
```

在区块链的上下文环境中，状态机是确定的。这意味着节点从给定状态开始，重放相同的Transaction序列，总能得到相同的最终状态。

Cosmos SDK为开发者提供了最大程度的灵活性去定义应用程序的状态，Transaction类型和状态转换功能。接下来的章节中会更详细地介绍使用SDK构建状态机的过程。在此之前，先让我们看看如何使用Tendermint复制状态机。

## Tendermint

得益于Cosmos SDK，开发者只需要定义好状态机，[Tendermint](https://tendermint.com/docs/introduction/what-is-tendermint.html)就会处理好状态复制的工作。

```
                ^  +-------------------------------+  ^
                |  |                               |  |   Built with Cosmos SDK
                |  |  State-machine = Application  |  |
                |  |                               |  v
                |  +-------------------------------+
                |  |                               |  ^
Blockchain node |  |           Consensus           |  |
                |  |                               |  |
                |  +-------------------------------+  |   Tendermint Core
                |  |                               |  |
                |  |           Networking          |  |
                |  |                               |  |
                v  +-------------------------------+  v
```

[Tendermint](https://tendermint.com/docs/introduction/what-is-tendermint.html) 是一个与应用程序无关的引擎，负责处理区块链的网络层和共识层。这意味着Tendermint负责对Transaction字节进行传播和排序。Tendermint Core 通过同名的拜占庭容错算法来达成Transaction顺序的共识。

Tendermint[共识算法](https://tendermint.com/docs/introduction/what-is-tendermint.html#consensus-overview)与一组被称为Validator的特殊节点共同运作。Validator负责向区块链中添加包含transaction的区块。在任何给定的区块中，都有一组Validator集合V。算法会从集合V中选出一个Validator作为下一个区块的Proposer。如果一个区块被集合V中超过三分之二的Validator签署了[prevote](https://tendermint.com/docs/spec/consensus/consensus.html#prevote-step-height-h-round-r)和[precommit](https://tendermint.com/docs/spec/consensus/consensus.html#precommit-step-height-h-round-r)，且区块中所有Transaction都是有效的，则认为该区块有效。Validator集合可以按照状态机中写定的规则更改。

## ABCI

Tendermint通过被称为[ABCI](https://tendermint.com/docs/spec/abci/)的接口向应用程序传递Transactions，该接口必须由应用程序实现。

```
              +---------------------+
              |                     |
              |     Application     |
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

需要注意的是，Tendermint仅处理transaction字节，它并不知道这些字节的含义。Tendermint所做的只是对transaction字节进行确定性地排序。Tendermint通过ABCI向应用程序传递字节，并期望返回状态码以获知包含在transactions中的messages是否成功处理。

以下是ABCI最重要的Messages：

`CheckTx`：当Tendermint Core接收到一个Transaction时，它会传递给应用程序以检查是否满足一些基本要求。`CheckTx` 用于保护全节点的内存池免受垃圾transactions攻击。`AnteHandler`这一特殊处理程序用于执行一系列验证步骤，例如检查手续费是否足够以及验证签名。如果检查通过，该transaction会被添加进[mempool](https://tendermint.com/docs/spec/reactors/mempool/functionality.html#mempool-functionality)，并广播给其他共识节点。请注意，此时transactions尚未被`CheckTx`处理（即未进行状态修改），因为它们还没有被包含在区块中。

`DeliverTx`：当Tendermint Core收到一个[有效区块](https://tendermint.com/docs/spec/blockchain/blockchain.html#validation)时，区块中的每一个Transaction都会通过`DeliverTx`传递给应用程序以进行处理。状态转换会在这个阶段中发生。`AnteHandler`会与Transaction中每个Message的实际[`handlers`](https://docs.cosmos.network/master/building-modules/handler.html)一起再次执行。

`BeginBlock/EndBlock`：无论区块中是否包含transaction，messages都会在每个区块的开头和结尾处执行。触发自动执行的逻辑是很有用的。但需要谨慎使用，因为计算量庞大的循环会严重降低区块链的性能，而无限循环甚至会导致区块链宕机。

获知更多关于ABCI的详细内容可以访问[Tendermint docs](https://tendermint.com/docs/spec/abci/abci.html#overview).

基于Tendermint构建的任何程序都需要实现ABCI接口，以便和底层的本地Tendermint引擎通信。幸运的是，您不需要实现ABCI接口，Cosmos SDK 以 [baseapp](https://docs.cosmos.network/master/intro/sdk-design.html#baseapp) 的形式提供了样板实现。