**原文路径:https://github.com/cosmos/cosmos-sdk/blob/master/docs/basics/gas-fees.md**

# Gas and Fees

## 必备阅读 {hide}

- [一个 SDK 程序的剖析](./app-anatomy.md) {prereq}

## `Gas` and `Fees`的介绍

在 Cosmos SDK 中,`gas`是一种特殊的单位，用于跟踪执行期间的资源消耗。每当对储存进行读写操作的时候会消耗`gas`，如果要进行比较复杂的计算的话也会消耗`gas`。它主要有两个目的:

- 确保区块不会消耗太多资源而且能顺利完成。这个默认在 SDK 的 [block gas meter](#block-gas-meter) 中保证
- 防止来自终端用户的垃圾消息和滥用。为此，通常会为 [`message`](../building-modules/messages-and-queries.md#messages) 执行期间的消耗进行定价，并产生 `fee`(`fees = gas * gas-prices`)。`fees` 通常必须由 `message` 的发送者来支付。请注意，SDK 并没有强制执行对 `gas` 定价，因为可能会有其他的方法来防止垃圾消息(例如带宽方案)。尽管如此，大多数应用程序仍然会使用`fee` 方式来防止垃圾消息。这个机制通过 [`AnteHandler`](#antehandler) 来完成.

## Gas Meter

在 Cosmos SDK 中 `gas` 是一种简单的 `uint64` 类型，被称之为 `gas meter` 的对象进行管理，Gas meters 实现了 `GasMeter` 接口

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/gas.go#L31-L39

这里:

- `GasConsumed()` 返回 `GasMeter` 实例中消耗的 `gas` 的数量
- `GasConsumedToLimit()` 返回 `GasMeter` 实例消耗的 gas 数量，如果达到上限的话就返回上限
- `Limit()` 返回 `GasMeter` 实例的上限，如果是 0 则表示对 `gas` 的数量没有限制
- `ConsumeGas(amount Gas, descriptor string)` 消耗提供的 `gas`，如果 `gas` 溢出了，使用 `descriptor` 内容进行报错，如果 `gas` 并不是无限的，则超过限制就会报错。
- `IsPastLimit()` 如果 `gas` 消耗超过了 `GasMeter` 的限制则返回 `true`，其它返回 `false`
- `IsOutOfGas()` 如果 `gas` 消耗大于或等于了 `GasMeter` 的限制则返回 `true`，其它返回 `false`

  `GasMeter` 通常保存在 [`ctx`](../core/context.md) 中，`gas` 消耗的方式如下:

```go
ctx.GasMeter().ConsumeGas(amount, "description")
```

通常，Cosmos SDK 使用两种不同的 `GasMeter`，[main gas meter](#main-gas-metter[) 和 [block gas meter](#block-gas-meter)。

### 主 Gas Meter

`ctx.GasMeter()` 是应用程序的主 `GasMeter`，主 `GasMeter` 通过 `BeginBlock` 中的 `setDeliverState` 进行初始化，然后跟踪导致状态转换的执行序列中的 `gas` 消耗。也即是说它的更新由 [`BeginBlock`](../core/baseapp.md#beginblock)，[`DeliverTx`](../core/baseapp.md#delivertx) 和 [`EndBlock`](../core/baseapp.md#endblock) 进行操作。主 `GasMeter` 必须在 [`AnteHandler`](#antehandler)中 **设置为 0**，以便它能获取每个 transaction 的 Gas 消耗

`gas`消耗可以手工完成，模块开发者通常在 [`BeginBlocker`,`EndBlocker`](../building-modules/beginblock-endblock.md) 或者 [`handler`](../building-modules/handler.md) 上执行，但大多数情况下，只要对储存区进行了读写，它就会自动完成。这种自动消耗的逻辑在[`GasKv`](../core/store.md#gaskv-store)中完成.

### 块 Gas Meter

`ctx.BlockGasMeter()` 是跟踪每个区块 `gas` 消耗并保证它没有超过限制的 `GasMeter`。每当 [`BeginBlock`](../core/baseapp.md#beginblock) 被调用的时候一个新的 `BlockGasMeter` 实例将会被创建。`BlockGasMeter` 的 `gas` 是有限的，每个块的 `gas` 限制应该在应用程序的共识参数中定义，Cosmos SDK 应用程序使用 Tendermint 提供的默认共识参数：

+++ https://github.com/tendermint/tendermint/blob/f323c80cb3b78e123ea6238c8e136a30ff749ccc/types/params.go#L65-L72

当通过 `DeliverTx` 处理新的 [transaction](../core/transactions.md) 的时候，`BlockGasMeter` 的当前值会被校验是否超过上限，如果超过上限，`DeliverTx` 直接返回，由于 `BeginBlock` 会消耗 `gas`，这种情况可能会在第一个 `transaction` 到来时发生，如果没有发生这种情况，`transaction`将会被正常的执行。在 `DeliverTx` 的最后，`ctx.BlockGasMeter()` 会追踪 `gas` 消耗并将它增加到处理 `transaction` 的 `gas` 消耗中.

```go
ctx.BlockGasMeter().ConsumeGas(
    ctx.GasMeter().GasConsumedToLimit(),
    "block gas meter",
)
```

## AnteHandler

`AnteHandler` 是一个特殊的处理程序，它在 `CheckTx` 和 `DeliverTx` 期间为每一个 `transaction` 的每个 `message` 处理之前执行。`AnteHandler` 相比 `handler` 有不同的签名:

```go
// AnteHandler authenticates transactions, before their internal messages are handled.
// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx, simulate bool) (newCtx Context, result Result, abort bool)
```

`AnteHandler` 不是在核心 SDK 中实现的，而是在每一个模块中实现的，这使开发者可以使用适合其程序需求的`AnteHandler`版本，也就是说当前大多数应用程序都使用 [`auth` module](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth) 中定义的默认实现。下面是 `AnteHandler` 在普通 Cosmos SDK 程序中的作用:

- 验证事务的类型正确。事务类型在实现 `anteHandler` 的模块中定义，它们遵循事务接口：

  +++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/tx_msg.go#L33-L41

  这使开发人员可以使用各种类型的应用程序进行交易。 在默认的 auth 模块中，标准事务类型为 StdTx：

  +++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/auth/types/stdtx.go#L22-L29

- 验证交易中包含的每个 [`message`](../building-modules/messages-and-queries.md#messages) 的签名，每个 `message` 应该由一个或多个发送者签名，这些签名必须在 `anteHandler` 中进行验证.
- 在 `CheckTx` 期间，验证 `transaction` 提供的 `gas prices` 是否大于本地配置 `min-gas-prices`(提醒一下，`gas-prices` 可以从以下等式中扣除`fees = gas * gas-prices`)`min-gas-prices` 是每个独立节点的本地配置，在`CheckTx`期间用于丢弃未提供最低费用的交易。这确保了内存池不会被垃圾交易填充.
- 设置 `newCtx.GasMeter` 到 0，限制为`GasWanted`。**这一步骤非常重要**，因为它不仅确保交易不会消耗无限的天然气，而且还会在每个 `DeliverTx` 重置 `ctx.GasMeter`(每次 `DeliverTx` 被调用的时候都会执行 `anteHandler`，`anteHandler` 运行之后 `ctx` 将会被设置为 `newCtx`)

如上所述，`anteHandler` 返回 `transaction` 执行期间所能消耗的最大的 `gas` 数量，称之为 `GasWanted`。最后实际 `gas` 消耗数量记为 `GasUsed`，因此我们必须使 `GasUsed =< GasWanted`。当返回时 `GasWanted` 和 `GasUsed` 都会被中继到共识引擎中.
