---
parent:
  order: false
---

# Cosmos SDK 文档

## 开始

- **[SDK 介绍](./intro/README.md)**：Cosmos SDK 的总体概览
- **[快速开始](./using-the-sdk/quick-start.md)**：构建一个标准的基于 cosmos sdk 的 app 并启动节点
- **[SDK 开发教程](https://github.com/cosmos/sdk-application-tutorial)**: 一个学习 SDK 的教程。它展示了如何从头开始基于 sdk 构建区块链, 并在此过程中解释了 SDK 的基本原理。

## 索引

- **[基础文档](./basics/)**：cosmos sdk 的基础概念文档，例如应用结构、交易的生命周期、账户管理等
- **[核心文档](./core/)**: cosmos sdk 的核心文档，例如`baseapp`，`store`，`server`等
- **[构建模块](./building-modules/)**: 对于模块开发者来说的一些重要概念，例如`message`，`keeper`，`handler`，`querier`
- **[接口](./run-node/)**: 为 cosmos 应用设计接口的文档

## 开发资源

- **[模块目录](../../x/)**：模块的实现和文档
- **[规范](./spec/):** Cosmos SDK 的模块及其他规范。
- **[SDK API 参考](https://godoc.org/github.com/cosmos/cosmos-sdk):** Cosmos SDK Godocs 文档 。
- **[REST API 规范](https://cosmos.network/rpc/):** 通过 REST 与 `gaia` 全节点交互的 API 列表。

## Cosmos Hub

Cosmos Hub (名为 `gaia`) 文档已经迁移到[这里](https://github.com/cosmos/gaia/tree/master/docs).

## 开发语言

Cosmos-SDK 目前是用 [Golang](https://golang.org/)编写的, 尽管该框架同样可以在其他语言中实现。请联系我们获取有关资助其他语言实现的信息。

## 贡献

参考 [文档说明](https://github.com/cosmos/cosmos-sdk/blob/master/docs/DOCS_README.md) 了解构建细节及更新时注意事项。
