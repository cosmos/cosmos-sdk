**原文路径:https://github.com/cosmos/cosmos-sdk/blob/master/docs/basics/accounts.md**

# 账户系统

# 必备阅读 {hide}

- [一个 SDK 程序的剖析](./app-anatomy.md) {prereq}

## 账户定义

在 Cosmos SDK 中，一个账户是指定的一个公私钥对。公钥可以用于派生出 `Addresses`，`Addresses` 可以在程序里面的各个模块间区分不同的用户。`Addresses` 同样可以和[消息](../building-modules/messages-and-queries.md#messages)进行关联用于确定发消息的账户。私钥一般用于生成签名来证明一个消息是被一个 `Addresses`(和私钥关联的`Addresses`) 所发送。

Cosmos SDK 使用一套称之为 [BIP32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki) 的标准来生成公私钥。这个标准定义了怎么去创建一个 HD 钱包(钱包就是一批账户的集合)。每一个账户的核心，都有一个种子，每一个种子都有一个 12 或 24 个字的助记符。使用这个助记符，使用一种单向的加密方法可以派生出任意数量的私钥。公钥可以通过私钥推导出来。当然，助记符是最敏感的信息，因为可以不停通过助记符来重新生成私钥。

```md
     Account 0                         Account 1                         Account 2

+------------------+              +------------------+               +------------------+
|                  |              |                  |               |                  |
|    Address 0     |              |    Address 1     |               |    Address 2     |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Public key 0    |              |  Public key 1    |               |  Public key 2    |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Private key 0   |              |  Private key 1   |               |  Private key 2   |
|        ^         |              |        ^         |               |        ^         |
+------------------+              +------------------+               +------------------+
         |                                 |                                  |
         |                                 |                                  |
         |                                 |                                  |
         +--------------------------------------------------------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Master PrivKey   |
                                 |                   |
                                 +-------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Mnemonic (Seed)  |
                                 |                   |
                                 +-------------------+
```

在 Cosmos SDK 中，账户可以在 [`Keybase`](#keybase) 中作为一个对象来储存和管理。

## Keybase

`Keybase` 是储存和管理账户的对象，在 Cosmos SDK 中，`Keybase` 要实现以下接口

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/types.go#L13-L86

在 Cosmos SDK 中，`Keybase` 接口的默认实现对象是 `dbKeybase`。

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/keybase.go

`dbKeybase` 上面对 `Keybase` 接口中方法实现的笔记:

- `Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)` 对 `message` 字节进行签名。需要做一些准备工作将 `message` 编码成 []byte 类型，可以参考 `auth` 模块 `message` 准备的例子。注意，SDK 上面没有实现签名的验证，签名验证被推迟到[`anteHandler`](#antehandler)中进行

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/auth/types/txbuilder.go#L176-L209

- `CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error)`创建一个新的助记符并打印在日志里，但是**并不保存在磁盘上**
- `CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error)` 基于[`bip44 path`](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)创建一个新的账户并将其保存在磁盘上。注意私钥在[保存前用密码加密](https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/mintkey/mintkey.go),**永远不会储存未加密的私钥**.在这个方法的上下文中, `account`和 `address` 参数指的是 BIP44 派生路径的段(例如`0`, `1`, `2`, ...)用于从助记符派生出私钥和公钥(注意：给相同的助记符和 `account` 将派生出相同的私钥，给相同的 `account` 和 `address` 也会派生出相同的公钥和 `Address`)。最后注意 `CreateAccount` 方法使用在 [Tendermint library](https://github.com/tendermint/tendermint/tree/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/secp256k1) 中的 `secp256k1` 派生出公私钥和 `Address`。总之，这个方法是用来创建用户的钥匙和地址的，并不是共识秘钥，参见[`Addresses`](#addresses) 获取更多信息

`dbKeybase` 的实现是最基本的，并没有根据需求提供锁定功能。锁定功能指如果一个`dbKeybase`实例被创建，底层的`db`就被锁定意味着除了实例化它的程序其他程序无法访问它。这就是 SDK 程序使用另外一套 `Keybase` 接口的实现 `lazyKeybase` 的原因

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/lazy_keybase.go

`lazyKeybase` 是 `dbKeybase` 的一个简单包装，它仅在要执行操作时锁定数据库，并在之后立即将其解锁。使用 `lazyKeybase`[命令行界面](../core/cli.md) 可以在 [rest server](../core/grpc_rest.md)运行时创建新的账户，它也可以同时传递多个 CLI 命令

## 地址和公钥

`Addresses` 和 `PubKey` 在程序里面都是标识一个参与者的公共信息。Cosmos SDK 默认提供 3 种类型的 `Addresses`和 `PubKey`

- 基于用户的 `Addresses` 和 `PubKey`，用于指定用户(例如 `message` 的发送者)。它们通过 **`secp256k1`** 曲线推导出来
- 基于验证节点的 `Addresses` 和 `PubKey` 用于指定验证者的操作员，它们通过 **`secp256k1`** 曲线推导出来
- 基于共识节点的 `Addresses` 和 `PubKey` 用于指定参与共识的验证着节点，它们通过 **`ed25519`** 曲线推导出来

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve       | Address byte length | Pubkey byte length |
| ------------------ | --------------------- | -------------------- | ----------- | ------------------- | ------------------ |
| Accounts           | cosmos                | cosmospub            | `secp256k1` | `20`                | `33`               |
| Validator Operator | cosmosvaloper         | cosmosvaloperpub     | `secp256k1` | `20`                | `33`               |
| Consensus Nodes    | cosmosvalcons         | cosmosvalconspub     | `ed25519`   | `20`                | `32`               |

### 公钥

在 Cosmos SDK 里面 `PubKey` 遵循在 tendermint 的 `crypto` 包中定义的 `Pubkey` 接口

+++ https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/crypto.go#L22-L27

对于 `secp256k1` 类型的秘钥，具体的实现可以在[这里](https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/secp256k1/secp256k1.go#L140)找到。对于`ed25519`类型的密钥，具体实现可以在[这里](https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/ed25519/ed25519.go#L135)找到。

请注意，在 Cosmos SDK 中，`Pubkeys` 并非以其原始格式进行操作。它使用 [`Amino`](../core/encoding.md#amino) 和 [`bech32`](https://en.bitcoin.it/wiki/Bech32) 进行 2 次编码。在 SDK 里面，`Pubkeys` 首先调用 `Bytes()` 方法在原始的 `Pubkey` 中(这里面提供 amino 编码)，然后使用 `bech32` 的 `ConvertAndEncode` 方法

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/address.go#L579-L729

### 地址

在 Cosmos SDK 默认提送 3 种类型的地址

- `AccAddress` 用于账户
- `ValAddress` 用于验证者操作员
- `ConsAddress` 用于验证者节点

这些地址类型都是一种长度为 20 的十六进制编码的 `[]byte` 数组的别名，这里有一种标准方法从`Pubkey pub`中获取到地址`aa`.

```go
aa := sdk.AccAddress(pub.Address().Bytes())
```

这些地址实现了 `Address` 接口

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/address.go#L71-L80

值得注意的是，`Marhsal()` 和 `Bytes()` 方法都返回相同的 `[]byte` 类型的地址，根据 protobuf 的兼容性要求我们需要前者。同样，`String()` 也被用来返回 `bech32` 编码类型的地址，这个应该是用户看到的最终编码形式。下面是一个例子:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/address.go#L229-L243

## 接下来 {hide}

学习[gas and fees](./gas-fees.md) {hide}
