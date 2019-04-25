# Gaia是什么

`gaia`是作为Cosmos SDK应用程序的Cosmos Hub的名称。它有两个主要的入口：

+ `gaiad` : Gaia的服务进程，运行着`gaia`程序的全节点。
+ `gaiacli` : Gaia的命令行界面，用于同一个Gaia的全节点交互。

`gaia`基于Cosmos SDK构建，使用了如下模块:

+ `x/auth` : 账户和签名
+ `x/bank` : token转账
+ `x/staking` : 抵押逻辑
+ `x/mint` : 增发通胀逻辑
+ `x/distribution` : 费用分配逻辑
+ `x/slashing` : 处罚逻辑
+ `x/gov` : 治理逻辑
+ `x/ibc` : 跨链交易
+ `x/params` : 处理应用级别的参数

> 关于Cosmos Hub : Cosmos Hub是第一个在Cosmos Network中上线的枢纽。枢纽的作用是用以跨链转账。如果区块链通过IBC协议连接到枢纽，它会自动获得对其它连接至枢纽的区块链的访问能力。Cosmos Hub是一个公开的PoS区块链。它的权益代币称为Atom。

接着，学习如何[安装Gaia](./installation.md)