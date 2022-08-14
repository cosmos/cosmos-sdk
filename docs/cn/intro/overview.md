# Cosmos SDK 介绍

## 什么是 Cosmos SDK

[Cosmos SDK](https://github.com/cosmos/cosmos-sdk)是开源框架，用于构建类似 Cosmos Hub 等基于 POS 共识算法的多元资产公有区块链，以及基于权威证明共识算法的许可链。使用 Cosmos SDK 构建的区块链通常被称为特定应用区块链（专用区块链）（application-specific blockchains）。

Cosmos SDK 的目标是让开发者可以快速地构建一条能与其他区块链以原生的方式进行互操作的可定制区块链。在我们的设想中，这套 SDK 就像 Web 应用框架一样，可以让开发者迅速构建出基于[Tendermint](https://github.com/tendermint/tendermint)的安全区块链应用程序。 基于 Cosmos SDK 的区块链由组合式[模块](https://docs.cosmos.network/master/building-modules/intro.html)构建，其中大部分模块都是开源的，且任何开发者均可使用。任何人都能为 Cosmos SDK 创建新的模块，集成已经构建的模块就像将他们导入你的区块链应用程序一样简单。还有一点，Cosmos SDK 是基于功能（capabilities）的系统，这允许开发者可以更好地考虑模块之间交互的安全性。更深入地了解功能，请跳至[本节](https://docs.cosmos.network/master/core/ocap.html)。

## 什么是特定应用区块链

目前在区块链领域中，一种开发模式是通过像以太坊这样的虚拟机区块链展开，即开发者在现有的区块链上通过智能合约的方式去构建去中心化应用。虽然智能合约在单用途应用场景（如 ICO）下非常有用，但在构建复杂的去中心化平台时无法达到要求。更具体地说，智能合约在灵活性、所有权、性能方面会受到限制。

特定应用区块链提供了与虚拟机区块链截然不同的开发模式。特定应用区块链是面向单个具体应用程序的高度定制化区块链：开发者可以完全自由地做出让应用程序可以达到最佳运行状态的设计决策。他们也可以提供更好的主导权、安全性和性能。

了解更多可参考[特定应用区块链](https://docs.cosmos.network/master/intro/why-app-specific.html)。

## 为什么选择 Cosmos SDK？

Cosmos SDK 是目前最先进的构建可定制化特定应用区块链的框架。以下是一些可能让你希望通过 Cosmos SDK 构建去中心化应用的原因：

- Cosmos SDK 默认的共识引擎是[Tendermint Core](https://github.com/tendermint/tendermint). Tendermint 是目前最成熟的、唯一的 BFT 共识引擎。它被广泛应用于行业中，被认为是构建 POS 系统的最佳标准共识引擎。

- Cosmos SDK 是开源的，你可以通过组合式[modules](https://docs.cosmos.network/master/x/)轻松地构建出区块链。随着 SDK 生态中各种开源模块的发展，通过 Cosmos SDK 构建复杂的去中心化平台会变得越来越容易。

- Cosmos SDK 受基于功能的安全性所启发，并受益于多年来在区块链状态机领域的经验。这让 Cosmos SDK 成为一个非常安全的构建区块链的环境。

- 最重要的是，Cosmos SDK 已经构建出了多个正在运行中的特定应用区块链。例如，Cosmos HUB，IRIS HUB，Binance Chain, Terra 和 Kava。更多基于 Cosmos SDK 构建的区块链参考[这里](https://cosmos.network/ecosystem)。

## 开始使用 Cosmos SDK

了解更多请参考[SDK 应用架构](https://docs.cosmos.network/master/intro/sdk-app-architecture.html)。

了解如何从头建立特定应用区块链，请参考[SDK 教程](https://cosmos.network/docs/tutorial)。
