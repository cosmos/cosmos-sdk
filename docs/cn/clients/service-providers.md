# 服务提供商（Service Providers）

我们将“服务提供商”定义为可以为最终用户提供服务的实体，这些实体涉及与基于 Cosmos-SDK 的区块链（包括 Cosmos Hub）的某种形式的交互。更具体地说，本文档将集中于与 token 的交互。

本节不涉及想要提供[轻客户端](https://docs.tendermint.com/master/tendermint-core/light-client.html)功能的钱包开发者。服务提供商将作为最终用户的区块链的可信接入点。

## 架构的高级描述

有三个主要部分需要考虑：

- 全节点：与区块链交互。
- Rest Server：它充当 HTTP 调用的中继者。
- Rest API：为 Rest Server 定义可用端点。

## 运行全节点

### 安装和配置

我们将描述为 Cosmos Hub 运行和交互全节点的步骤。对于其他基于 SDK 的区块链，该过程是类似的。

首先，您需要[安装软件](../cosmos-hub/installation.md).

然后，您可以开始[运行全节点](../cosmos-hub/join-testnet.md).

### 命令行界面（CLI）

接下来，您将用一些 CLI 命令与全节点交互。

#### 创建秘钥对

生成新秘钥（默认使用 secp256k1 椭圆曲线算法）：

```bash
gaiacli keys add <your_key_name>
```

系统将要求您为此密钥对输入密码（至少 8 个字符）。该命令返回 4 个信息：

- `NAME`: 秘钥名称。
- `TYPE`：秘钥类型，总是`local`。
- `ADDRESS`：您的地址，用于接收资金。
- `PUBKEY`：您的公钥, 用于验证者.
- `MNEMONIC`： 由 24 个单词组成的助记词。 **将这个助记词保存在安全的地方**，它用于在您忘记密码时恢复您的私钥。

您可以输入以下命令查看所有可用密钥：

```bash
gaiacli keys list
```

#### 检查您的余额

收到代币到您的地址后，您可以输入以下命令查看帐户的余额：

```bash
gaiacli account <YOUR_ADDRESS>
```

_注意：当您查询没有 token 帐户的余额时，您将得到以下错误：找不到地址为<YOUR_ADDRESS>的帐户。这是预料之中的！我们正在努力改进我们的错误提示信息。_

#### 通过 CLI 发送代币

以下是通过 CLI 发送代币的命令：

```bash
gaiacli tx send <from_key_or_address> <to_address> <amount> \
    --chain-id=<name_of_testnet_chain>
```

参数：

- `<from_key_or_address>`: 发送账户的名称或地址。
- `<to_address>`: 接收者地址。
- `<amount>`: 接受`<value|coinName>`格式的参数，例如 `10faucetToken`。

标识：

- `--chain-id`: 此标志允许您指定链的 ID，不同的 testnet 链和主链会有不同的 id。

#### 帮助

如果您需要进行其他操作，最合适的命令是：

```bash
gaiacli
```

它将显示所有可用命令。对于每个命令，您可以使用`--help`标识来获取更多信息。

## 设置 Rest 服务器

Rest 服务器充当前端点和全节点之间的媒介。 Rest 服务器不必与全节点同一台计算机上运行。

要启动 Rest 服务器：

```bash
gaiacli rest-server --node=<full_node_address:full_node_port>
```

Flags:

- `--node`: 全节点的 IP 地址和端口。格式为 `<full_node_address:full_node_port>`。如果全节点在同一台机器上，则地址应为 `tcp：// localhost：26657`。
- `--laddr`: 此标识允许您指定 Rest 服务器的地址和端口（默认为“1317”）。通常只使用这个标识指定端口，此时只需输入 “localhost” 作为地址，格式为`<rest_server_address:port>`。

### 监听入向交易

监听入向交易推荐的方法是通过 LCD 的以下端点定期查询区块链：

<!-- [`/bank/balance/{address}`](https://cosmos.network/rpc/#/ICS20/get_bank_balances__address_) -->

## Rest API

Rest API 记录了可用于与全节点交互的所有可用端点，您可以在[这里](https://cosmos.network/rpc/)查看。

API 针对每种类别的端点归纳为 ICS 标准。例如，[ICS20](https://cosmos.network/rpc/#/ICS20/)描述了 API 与 token 的交互。

为了给开发者提供更大的灵活性，我们提供了生成未签名交易、[签名](https://cosmos.network/rpc/#/ICS20/post_tx_sign)和[广播](https://cosmos.network/rpc/#/ICS20/post_tx_broadcast)等不同的 API 端点。这允许服务提供商使用他们自己的签名机制。

为了生成一个未签名交易（例如 [coin transfer](https://cosmos.network/rpc/#/ICS20/post_bank_accounts__address__transfers)），你需要在 `base_req` 的主体中使用 `generate_only` 字段。

## Cosmos SDK 交易签名

Cosmos SDK 签名是一个相当简单的过程。

每个 Cosmos SDK 交易都有一个规范的 JSON 描述。 `gaiacli`和 REST 接口为交易提供规范的 JSON 描述，“广播”功能将提供紧凑的 Amino（类似 protobuf 的格式）编码转换。

签名消息时的注意事项：

格式如下

```json
{
  "account_number": XXX,
  "chain_id": XXX,
  "fee": XXX,
  "sequence": XXX,
  "memo": XXX,
  "msgs": XXX
}
```

签名者必须提供 `"chain_id"`、 `"account number"` 和 `"sequence number"`。

交易构造接口将生成 `"fee"`、 `"msgs"` 和 `"memo"` 等字段.

You can load the mempool of a full node or validator with a sequence of uncommitted transactions with incrementing
sequence numbers and it will mostly do the correct thing.

`"account_number"` 和 `"sequence"` 字段可以直接从区块链或本地缓存中查询。 错误的获取了这些数值和 chainId，是产生无效签名错误的常见原因。您可以通过加载全节点或验证人中的 mempool 来获取未提交交易的自增序号，这样大大增加成功概率。

您可以使用递增序列号的一系列未提交事务加载完整节点或验证器的 mempool，它将主要执行正确的操作。

在签名之前，所有键都要按字典顺序排序，并从 JSON 输出中删除所有空格。

签名编码是 ECDSArands 的 64 字节连结（即`r || s`），其中`s`按字典顺序小于其反转以防止延展性。 这就像以太坊一样，但没有用户公钥恢复的额外字节，因为 Tendermint 假定公钥一定会提供。

已签名交易中的签名和公钥示例:

```json
{
  "type": "cosmos-sdk/StdTx",
  "value": {
    "msg": [...],
    "signatures": [
      {
        "pub_key": {
          "type": "tendermint/PubKeySecp256k1",
          "value": XXX
        },
        "signature": XXX
      }
    ],
  }
}
```

正确生成签名后，将 JSON 插入生成的交易中，然后调用广播端点进行广播。
