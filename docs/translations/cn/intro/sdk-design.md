# Cosmos SDK设计概览

Cosmos SDK是一个方便开发者开发基于Tendermint的安全可靠状态机的一套框架。其核心是Golang版的ABCI的实现。它附带一个`multistore`来持久化存储数据还有一个`router`来处理事务。

下面是一个简单的视图，展示了当从Tendermint的`DeliverTx`请求（`CheckTx`的处理流程与其相同，除了不会执行状态的改变）中接收到一笔交易时，基于Cosmos SDK构建的应用程序是如何处理交易的：

1. 解码从Tendermint共识引擎接收到的交易（记住Tendermint只处理`[]bytes`）
2. 从交易中提取消息并进行基本的合理性检查。
3. 将每条消息路由至对应的模块进行处理。
4. 提交状态变更。

该应用程序能让你生成交易，编码它们并传递给地城的Tendermint来进行广播

## `baseapp`

`baseApp`是Cosmos SDK的ABCI的实现样板。它有一个`router`用作把交易路由到对应的模块。你的应用程序的的主体`app.go`文件将定义你自定义的`app`类型，将嵌入`baseapp`。这样，你自定义的`app`类型将自动继承`baseapp`的所有方法。阅览[SDK应用程序教程](https://github.com/cosmos/sdk-application-tutorial/blob/master/app.go#L27)。

`baseapp`的目的是在存储和可扩展状态机之间提供安全接口，同时尽可能少地定义该状态机（保持对ABCI的真实性）。

有关`baseapp`的更多信息，请点击[这里](../concepts/baseapp.md)。

## Multistore

Cosmos SDK为将状态持久化提供了multistore。multistore允许开发者声明任意数量的[`KVStores`](https://github.com/blocklayerhq/chainkit)。`KVStores`只接受`[]byte`类型作为值，因此任何自定义的类型都需要在存储之前使用[go-amino](https://github.com/tendermint/go-amino)进行编码。

multistore用于区分不同的模块的状态，每个都有其模块管理。要了解更多关于multistore的信息，点击[这里](../concepts/store.md)

## Modules

Cosmos SDK的强大之处在于其模块发开发的理念。SDK应用程序通过把一组可以互相操作的模块组合起来生成。每个模块定义状态的自己并包含其自己的消息/交易处理器，而SDK负责将每条消息路由到其各自归属的模块。

```
                                      +
                                      |
                                      |  Transaction relayed from Tendermint
                                      |  via DeliverTx
                                      |
                                      |
                +---------------------v--------------------------+
                |                 APPLICATION                    |
                |                                                |
                |     Using baseapp's methods: Decode the Tx,    |
                |     extract and route the message(s)           |
                |                                                |
                +---------------------+--------------------------+
                                      |
                                      |
                                      |
                                      +---------------------------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |  Message routed to the correct
                                                                  |  module to be processed
                                                                  |
                                                                  |
+----------------+  +---------------+  +----------------+  +------v----------+
|                |  |               |  |                |  |                 |
|  AUTH MODULE   |  |  BANK MODULE  |  | STAKING MODULE |  |   GOV MODULE    |
|                |  |               |  |                |  |                 |
|                |  |               |  |                |  | Handles message,|
|                |  |               |  |                |  | Updates state   |
|                |  |               |  |                |  |                 |
+----------------+  +---------------+  +----------------+  +------+----------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |
                                       +--------------------------+
                                       |
                                       | Return result to Tendermint
                                       | (0=Ok, 1=Err)
                                       v
```

每个模块都可以看做一个小型的状态机。开发人员需要定义模块所处理的状态的自己，已经用作修改状态的message类型(*注意*:message是由`baseapp`的方法从交易中提取的)。通常，每个模块在multistore声明它自己的`KVStore`来持久化保存它所定义的状态子集。大多数开发者在构建自己的模块时也需要访问其他的第三方模块。鉴于Cosmos-SDK是一个开源的框架，一些模块可能是恶意的，这就意味着需要安全原则来合理化模块之间的交互。这些原则基于[object-capabilities](./ocap.md)。实际上，这意味着不是让每个模块保留其他模块的访问控制列表，而是每个模块都实现称作keeper的特殊对象，这些对象可以传递给其他模块并授予预先定义的一组功能。

SDK模块在SDK的`x/`目录下定义。一些核心模块包括：
+ `x/auth`: 用于管理账户和签名.
+ `x/bank`: 用于实现token和token转账.
+ `x/staking` + `x/slashing`: 用于构建POS区块链.

除了`x/`中已有的模块，任何人都可以在他们的应用程序中使用它们自己定义的模块。你可以查看[示例教程](https://cosmos.network/docs/tutorial/keeper.html)。

### 接下来 学习Cosmos SDK的安全模型[ocap](./ocap.md)
