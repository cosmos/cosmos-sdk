# Getting Started

To start a rest server, we need to specify the following parameters:

| Parameter       | Type      | Default                 | Required | Description                                          |
| --------------- | --------- | ----------------------- | -------- | ---------------------------------------------------- |
| chain-id        | string    | null                    | true     | ID of chain we connect to, must be specified |
| home            | string    | "$HOME/.gaiacli"        | false    | directory for config and data, such as key and checkpoint |
| node-list       | string    | "tcp://localhost:26657" | false    | Full node list to connect to, example: "tcp://10.10.10.10:26657,tcp://20.20.20.20:26657" |
| laddr           | string    | "tcp://localhost:1317"  | false    | Address for server to listen on |
| trust-node      | bool      | false                   | false    | Trust full nodes or not |
| swagger-host-ip | string    | "localhost"             | false    | The host IP of the Gaia-lite server, swagger-ui will send request to this host |
| modules         | string    | "general,key,token"     | false    | Enabled modules |

Sample command to start gaia-lite node:
```
gaiacli lite-server --chain-id=<chain_id>
```

When the connected full node is trusted, then the proof is not necessary, so you can run gaia-lite node with trust-node option:
```
gaiacli lite-server --chain-id=<chain_id> --trust-node
```

If you have want to run gaia-lite node in a remote server, then you must specify the public ip to swagger-host-ip
```
gaiacli lite-server --chain-id=<chain_id> --swagger-host-ip=<remote_server_ip>
```

The gaia-lite node can connect to multiple full nodes. Then gaia-lite will do load balancing for full nodes which is helpful to improve reliability and TPS. You can do this by this command:
```
gaiacli lite-server --chain-id=<chain_id> --node-list=tcp://10.10.10.10:26657,tcp://20.20.20.20:26657
```

The gaia-lite support modular rest APIs. Now it supports four modules: general, key, token and stake. If you need all of them, just start it with this command:
 ```
 gaiacli lite-server --chain-id=<chain_id> --modules=general,key,token,stake
 ```