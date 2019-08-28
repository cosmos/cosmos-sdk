# Cosmos SDK 设计概览

Cosmos SDK是一个方便开发者开发基于Tendermint的安全可靠状态机的一套框架。其核心是Golang版的ABCI的实现。它附带一个`multistore`来持久化存储数据还有一个`router`来处理交易。

下面一个简单的视图展示了当从Tendermint的`DeliverTx`请求（`CheckTx`的处理流程与其相同，除了不会执行状态的改变）中接收到一笔交易时，基于Cosmos SDK构建的应用程序是如何处理交易的：

1. 解码从Tendermint共识引擎接收到的交易（记住Tendermint只处理`[]bytes`）
2. 从交易中提取消息并进行基本的合理性检查。
3. 将每条消息路由至对应的模块进行处理。
4. 提交状态变更。

应用同样可以生成交易，进行编码并传递给底层的Tendermint来进行广播

## `baseapp`

`baseApp` 是Cosmos SDK的ABCI的实现样板。里面的 `router` 用来把交易路由到对应的模块。我们应用程序的主体文件`app.go` 将自定义`app`类型，它将嵌入`baseapp`。这样，自定义的`app`类型将自动继承`baseapp`的所有方法。阅览[SDK应用教程](https://github.com/cosmos/sdk-application-tutorial/blob/master/app.go#L27)代码示例。

`baseapp`的目的是在存储和可扩展状态机的之间提供安全接口，同时尽可能少地定义该状态机（保持对ABCI的真实性）。

有关`baseapp`的更多信息，请点击[这里](../concepts/baseapp.md)。

## Multistore

Cosmos SDK 为状态持久化提供了 multistore 。multistore 允许开发者声明任意数量的[`KVStores`](https://github.com/blocklayerhq/chainkit)。`KVStores`只接受`[]byte`类型作为值，因此任何自定义的类型都需要在存储之前使用[go-amino](https://github.com/tendermint/go-amino)进行编码。

multistore 抽象用于区分不同的模块的状态，每个都由其自身模块管理。要了解更多关于 multistore 的信息，点击[这里](../concepts/store.md)

## Modules

Cosmos SDK 的强大之处在于其模块化开发的理念。应用程序通过把一组可以互相操作的模块组合起来进行构建。每个模块定义状态子集，并包含其自己的消息/交易处理器，而SDK负责将每条消息路由到其各自归属的模块。

下面是一个简化视图, 旨在说明每个应用链的全节点是如何处理接收的有效块中交易的:

```
                                      +
                                      |
                                      |  交易通过全节点的 Tendermint 引擎的DeliverTx
                                      |  传递到应用层
                                      |
                                      |
                +---------------------v--------------------------+
                |                    应用（层）                    |
                |                                                |
                |         用 baseapp 的方法: 解码 Tx,              |
                |             提取及路由消息                       |
                |                                                |
                +---------------------+--------------------------+
                                      |
                                      |
                                      |
                                      +---------------------------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |  消息传给相应的模块处理
                                                                  |
                                                                  |
+----------------+  +---------------+  +----------------+  +------v----------+
|                |  |               |  |                |  |                 |
|  AUTH MODULE   |  |  BANK MODULE  |  | STAKING MODULE |  |   GOV MODULE    |
|                |  |               |  |                |  |                 |
|                |  |               |  |                |  | 处理消息, 更改状态 |
|                |  |               |  |                |  |                 |
|                |  |               |  |                |  |                 |
+----------------+  +---------------+  +----------------+  +------+----------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |
                                       +--------------------------+
                                       |
                                       |  返回结果到 Tendermint
                                       | (0=Ok, 1=Err)
                                       v
```

每个模块都可以看做一个小型的状态机。开发人员需要定义模块所处理的状态的子集，以及修改状态的message自定义类型(*注意* : message 是由 `baseapp` 的方法从交易中提取的)。通常，每个模块在 multistore 声明它自己的`KVStore` 来持久化保存它所定义的状态子集。大多数开发者在构建自己的模块时也需要访问其他的第三方模块。鉴于Cosmos-SDK是一个开源的框架，一些模块可能是恶意的，这就意味着需要安全原则来合理化模块之间的交互。这些原则基于[object-capabilities](./ocap.md)。实际上，这意味着不是让每个模块保留其他模块的访问控制列表，而是每个模块都实现称作`keeper`的特殊对象，这些对象可以传递给其他模块并授予预先定义的一组能力。

SDK模块在SDK的`x/`目录下定义。一些核心模块包括：
+ `x/auth`: 用于管理账户和签名.
+ `x/bank`: 用于实现token和token转账.
+ `x/staking` + `x/slashing`: 用于构建POS区块链.

除了`x/`中已有的模块，任何人都可以在他们的应用程序中使用它们自己定义的模块。你可以查看[示例教程](https://learnblockchain.cn/docs/cosmos/tutorial/04-keeper.html)。

### 接下来 学习 Cosmos SDK 安全模型，[ocap](./ocap.md)
