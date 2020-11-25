# 对象能力模型（Object-Capability Model）

## 介绍

在考虑安全性时，最好从特定的威胁模型开始。我们的威胁模型如下:

> 我们假设蓬勃发展的 Cosmos-SDK 模块生态中会包含错误或恶意的模块。

Cosmos SDK 旨在通过以对象能力系统作为基础来解决此威胁。

> 对象能力系统的结构特性有利于代码设计模块化，并确保代码实现的可靠封装。
>
> 这些结构上的特性便于分析一个对象能力程序或操作系统的某些安全属性。其中一些 - 特别是信息流属性 - 可以在对象引用和连接级别进行分析，而不需要依赖于了解或分析（决定对象行为的）代码。
>
> 因此，可以在存在包含未知或（可能）恶意代码的新对象的情况下建立和维护这些安全属性。
>
> 这些结构属性源于管理对已存在对象的访问的两个规则：
>
> 1. 只有在对象 A 持有对象 B 的引用，A 才可以向 B 发送一条消息
> 2. 只有对象 A 收到了一条包含对象 C 引用的消息，A 才可以获得 C 的引用
>
> 根据这两条规则，一个对象只有通过一条先前存在的引用链获得另一个对象的引用，简而言之，“只有连接才能产生连接”。

关于对象能力（object-capabilities），可以阅读这边[文章](http://habitatchronicles.com/2017/05/what-are-capabilities/)了解更多。

## 对象能力模式实践

想法就是只暴露完成工作所需要的部分。

比如，下面的代码片段违反了对象能力原则：

```go
type AppAccount struct {...}
var account := &AppAccount{
    Address: pub.Address(),
    Coins: sdk.Coins{sdk.NewInt64Coin("ATM", 100)},
}
var sumValue := externalModule.ComputeSumValue(account)
```

方法名`ComputeSumValue`暗示了这是一个不修改状态的纯函数，但传入指针值意味着函数可以修改其值。更好的函数定义是使用一个拷贝来替代：

```go
var sumValue := externalModule.ComputeSumValue(*account)
```

在 Cosmos SDK 中，你可以看到[gaia app](https://github.com/cosmos/cosmos-sdk/blob/master/simapp/app.go)中对该原则的实践。

+++ https://github.com/cosmos/gaia/blob/master/app/app.go#L197-L209
