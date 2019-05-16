## 安装Gaia

本教程将详细说明如何在你的系统上安装`gaiad`和`gaiacli`。安装后，你可以作为[全节点](./join-mainnet.md)或是[验证人节点](./validators/validator-setup.md)加入到主网。

### 安装Go

按照[官方文档](https://golang.org/doc/install)安装`go`。记得设置环境变量`$GOPATH`,`$GOBIN`和`$PATH`:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.bash_profile
echo "export GO111MODULE=on" >> ~/.bash_profile
source ~/.bash_profile
```

::: tip
Cosmos SDK需要安装**Go 1.12+**
:::

### 安装二进制执行程序

接下来，安装最新版本的Gaia。需要确认您 `git checkout 了正确的[发布版本](https://github.com/cosmos/cosmos-sdk/releases)。

::: warning
对于主网，请确保你的版本大于或等于`v0.33.0`
:::

```bash
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk && git checkout <latest-release-tag>
make tools install
```

> *注意*: 如果在这一步中出现问题，请检查你是否安装的是Go的最新稳定版本。

等`gaiad`和`gaiacli`可执行程序安装完之后，请检查:

```bash
$ gaiad version --long
$ gaiacli version --long
```

`gaiacli`的返回应该类似于：

```
cosmos-sdk: 0.33.0
git commit: 7b4104aced52aa5b59a96c28b5ebeea7877fc4f0
go.sum hash: d156153bd5e128fec3868eca9a1397a63a864edb5cfa0ac486db1b574b8eecfe
build tags: netgo ledger
go version go1.12 linux/amd64
```

##### Build Tags

build tags指定了可执行程序具有的特殊特性。

| Build Tag | Description                                     |
| --------- | ----------------------------------------------- |
| netgo     | Name resolution will use pure Go code           |
| ledger    | 支持Ledger设备(硬件钱包) |

### 接下来
然后你可以选择 加入[主网](./join-mainnet.md)、[公共测试网](./join-testnet.md) 或是 [创建私有测试网](./deploy-testnet.md)。
