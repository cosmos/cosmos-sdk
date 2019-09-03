# SDK介绍

## 什么是Cosmos SDK?

[Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) 是一个架构，用于构建多资产股权证明(PoS)的区块链，比如Cosmos Hub，以及权益证明(PoA)的区块链。使用Cosmos SDK构建的区块链通常称为**特定应用区块链**。

Cosmos SDK的目标是允许开发者从头开始轻松创建原生就能同其他区块链相互操作的自定义区块链。我们设想SDK类似于Ruby-on-Rails框架之上构建应用一样，可以很方便在[Tendermint](https://github.com/tendermint/tendermint)之上构建安全的区块链应用。 基于SDK的区块链通过可组合的模块构建出来的，大部分模块是开源的，并且可供任何开发人员使用。 任何人都可以为Cosmos-SDK 创建一个模块，集成已经构建的模块就像将它们导入到区块链应用程序一样简单。 更重要的是，Cosmos SDK是一个基于**能力**（capabilities）的系统，开发人员可以更好地了解模块之间交互的安全性。 要深入了解能力，请跳到[OCAP](./ocap.md)。

## 什么是特定应用区块链?


今天区块链的一个发展模式是像以太坊这样的虚拟机区块链，开发通常围绕着在现有区块链之上通过智能合约构建一个去中心化的应用程序。 虽然智能合约对于像单用途应用程序（如ICO）这样的一些场景非常有用，但对于构建复杂的去中心化平台往往是不够的。 更一般地说，智能合约在灵活性、主权和性能方面受到限制。

特定应用区块链提供了与虚拟机区块链截然不同的开发模式。 特定应用区块链是一个定制的区块链来服务单个应用程序：开发人员可以自由地做出应用程序运行最佳所需的设计决策。 它们还可以提供更好的主权、安全和性能。

要了解有关特定应用区块链的更多信息，可参考[这里](./why-app-specific.md)。

## 为什么是 Cosmos SDK?

Cosmos SDK 是目前用于构建自定义的特定应用区块链的最先进的框架。 以下是一些你可能需要考虑使用 Cosmos SDK 构建去中心化应用的原因：

* SDK中默认共识引擎是  [Tendermint Core](https://github.com/tendermint/tendermint) 。 Tendermint 是已存在的最成熟（也是唯一的）的BFT共识引擎。 它被广泛应用于行业，被认为是建立股权证明系统（POS）的黄金标准共识引擎。
* SDK是开源的，旨在使其易于从可组合模块中构建区块链。 随着开源SDK模块生态系统的发展，使用它构建复杂的去中心化平台将变得越来越容易。
* SDK 受基于能力的安全性启发，及多年来解决区块链状态机的经验。 这使得 Cosmos SDK 成为构建区块链的非常安全的环境。
* 最重要的是，Cosmos SDK已经被许多特定应用区块链产品所使用。 如：[Cosmos Hub](https://hub.cosmos.network), [Iris](https://irisnet.org), [Binance Chain](https://docs.binance.org/), [Terra](https://terra.money/) or [Lino](https://lino.network/) ，除此之外还有很多建立在Cosmos SDK的项目。 你可以在这里查看[生态系统](https://cosmos.network/ecosystem)。


## 开始使用 Cosmos SDK 

* 了解[SDK 应用体系架构](./sdk-app-architecture.md)的详细信息
* 了解如何从头构建特定应用区块链，参考[SDK教程](/docs/tutorial) 。