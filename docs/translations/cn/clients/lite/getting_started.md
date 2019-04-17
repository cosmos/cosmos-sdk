# 入门

要启动 REST 服务器，我们需要指定以下参数：

| 参数        | 类型      | 默认值                  | 必填  | 描述                         |
| ----------- | --------- | ----------------------- | ----- | ---------------------------- |
| chain-id    | string    | null                    | true  | 要链接全节点的 chain id      |
| node        | URL       | "tcp://localhost:46657" | true  | 要链接全节点的地址和端口号   |
| laddr       | URL       | "tcp://localhost:1317"  | true  | 提供 REST 服务的地址和端口号 |
| trust-node  | bool      | "false"                 | true  | 是否信任 LCD 连接的全节点    |
| trust-store | DIRECTORY | "$HOME/.lcd"            | false | 保存检查点和验证人集的目录   |

示例：

```bash
gaiacli rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=false
```

有关Gaia-Lite RPC的更多信息，请参阅 [swagger documentation](https://cosmos.network/rpc/)

