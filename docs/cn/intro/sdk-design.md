# Cosmos SDK的主要组件

Cosmos SDK 是一个框架，可以促进基于Tendermint的安全状态机的开发。SDK的核心是一个基于Golang的[ABCI](https://docs.cosmos.network/master/intro/sdk-app-architecture.html#abci)样板实现。它带有一个用于存储数据的[`multistore`](https://docs.cosmos.network/master/core/store.html#multistore)，和一个用于处理Transaction的[`router`](https://docs.cosmos.network/master/core/baseapp.html#routing)。

下面的简化视图展示了当通过`DeliverTx`从Tendermint 转移transactions时，基于Cosmos SDK构建的应用程序如何处理这些transactions。

- 解码从Tendermint共识引擎中接收到的`transactions`（Tendermint只能处理 `[]bytes` 类型的数据）

- 从`transactions`中提取`messages`并进行基本的健全性检查。

- 将每个Message路由到对应的模块中，以进行相应处理。

- 提交状态更改。

## BaseApp

`baseapp` 是 Cosmos SDK 应用程序的样本实现，它拥有能够处理和底层共识引擎的连接的ABCI实现。通常，Cosmos SDK 应用程序通过嵌入[`app.go`](https://docs.cosmos.network/master/basics/app-anatomy.html#core-application-file)来实现拓展。查看示例请参考SDK应用教程：

```go
type nameServiceApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	// Keepers
	accountKeeper  auth.AccountKeeper
	bankKeeper     bank.Keeper
	stakingKeeper  staking.Keeper
	slashingKeeper slashing.Keeper
	distrKeeper    distr.Keeper
	supplyKeeper   supply.Keeper
	paramsKeeper   params.Keeper
	nsKeeper       nameservice.Keeper

	// Module Manager
	mm *module.Manager
}
```

`baseapp` 的目标是在存储和可拓展状态机之间提供安全的接口，同时尽可能少地定义状态机（对ABCI保持不变）。

更多关于`baseapp`的信息，请点击[这里](https://docs.cosmos.network/master/core/baseapp.html)。

## Multistore

Cosmos SDK 为状态持久化提供了`multistore`。Multistore允许开发者声明任意数量的`KVStores`。这些`KVStores`只接受`[]byte`类型的值，因此任何自定义的结构都需要在存储之前使用[codec](https://docs.cosmos.network/master/core/encoding.html)进行编码。

Multistore抽象用于区分不同模块的状态，每个都由其自己的模块管理。更多关于multistore的信息请点击[这里](https://docs.cosmos.network/master/core/store.html#multistore)。

## Modules

Cosmos SDK的强大之处在于其模块化开发的理念。SDK应用程序是通过组合一系列可互操作的模块而构建的。每个模块定义了状态子集，并包含其Messages与Transactions的处理器，同时SDK负责将每个Message路由到对应的模块中。

以下的简化视图展示了应用链中的每个全节点如何处理有效区块中的Transaction。

```
                                      +
                                      |
                                      |  Transaction relayed from the full-node's Tendermint engine
                                      |  to the node's application via DeliverTx
                                      |
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

每个模块都可以看成是一个小的状态机。开发者需要定义由模块处理的状态子集，同时自定义改变状态的Message类型（注意：`messages`是通过`baseapp`从`transactions`中提取的）。通常，每个模块会在`multistore`中声明自己的`KVStore`，以存储自定义的状态子集。大部分开发者在构建自己的模块时，需要访问其它第三方模块。由于Cosmos SDK是一个开放的框架，其中的一些模块可能是恶意的，这意味着需要一套安全原则去考虑模块间的交互。这些原则都基于[object-capabilities](https://docs.cosmos.network/master/core/ocap.html)。事实上，这也意味着，并不是要让每个模块都保留其他模块的访问控制列表，而是每个模块都实现了被称为`keepers`的特殊对象，它们可以被传递给其他模块，以授予一组预定义的功能。

SDK模块被定义在SDK的 `x/`文件夹中，一些核心的模块包括：

- `x/auth`：用于管理账户和签名。

- `x/bank`：用于启动 tokens 和 token 转账。

- `x/staking` + `s/slashing`：用于构建POS区块链。

除了`x/`文件夹中已经存在的任何人都可以使用的模块，SDK还允许您构建自己自定义的模块，您可以在[教程中查看示例](https://cosmos.network/docs/tutorial/keeper.html)。