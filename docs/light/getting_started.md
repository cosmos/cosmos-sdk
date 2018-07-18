# Getting Started

To start a rest server, we need to specify the following parameters:

| Parameter       | Type      | Default                 | Required | Description                                          |
| -----------     | --------- | ----------------------- | -------- | ---------------------------------------------------- |
| home            | string    | "$HOME/.gaiacli"        | false    | directory for save checkpoints and validator sets    |
| chain-id        | string    | null                    | true     | chain id of the full node to connect                 |
| node-list       | URL       | "tcp://localhost:26657" | false    | addresses of the full node to connect                |
| laddr           | URL       | "tcp://localhost:1317"  | false    | address to run the rest server on                    |
| trust-node      | bool      | "false"                 | false    | Whether this LCD trust full nodes or not             |
| swagger-host-ip | string    | "localhost"             | false    | The IP of the server which Cosmos-LCD is running on  |
| modules         | string    | "general,key,token"     | false    | enabled modules.                                     |

* When the connected full node is trusted, then the proof is not necessary, so you can run Cosmos-LCD with trust-node option:
```
gaiacli advanced rest-server-swagger --chain-id {your chain id} --trust-node
```
You can't specify true or false for this option. Once this option exist, then it is trust mode; otherwise, it is distrust mode. If you want Cosmos-LCD to run in distrust mode, just remove this option.

If you have gaiad running on your local machine, and its listening port is 26657, then you can start Cosmos-LCD just with the following command:
```bash
gaiacli advanced rest-server-swagger --chain-id {your chain id}
```

## Gaia Light Use Cases

LCD could be very helpful for related service providers. For a wallet service provider, LCD could
make transaction faster and more reliable in the following cases.

### Create an account

![deposit](pics/create-account.png)

First you need to get a new seed phrase :[get-seed](api.md#keysseed---get)

After having new seed, you could generate a new account with it : [keys](api.md#keys---post)

### Transfer a token

![transfer](pics/transfer-tokens.png)

The first step is to build an asset transfer transaction. Here we can post all necessary parameters
to /create_transfer to get the unsigned transaction byte array. Refer to this link for detailed
operation: [build transaction](api.md#create_transfer---post)

Then sign the returned transaction byte array with users' private key. Finally broadcast the signed
transaction. Refer to this link for how to broadcast the signed transaction: [broadcast transaction](api.md#create_transfer---post)
