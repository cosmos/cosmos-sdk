## Getting Started

To start a rest server, we need to specify the following parameters:

| Parameter   | Type      | Default                 | Required | Description                                          |
| ----------- | --------- | ----------------------- | -------- | ---------------------------------------------------- |
| chain-id    | string    | null                    | true     | chain id of the full node to connect                 |
| node        | URL       | "tcp://localhost:46657" | true     | address of the full node to connect                  |
| laddr       | URL       | "tcp://localhost:1317"  | true     | address to run the rest server on                    |
| trust-node  | bool      | "false"                 | true     | Whether this LCD is connected to a trusted full node |
| trust-store | DIRECTORY | "$HOME/.lcd"            | false    | directory for save checkpoints and validator sets    |

**Sample command** :

```
gaiacli light-client --chain-id=test --laddr=tcp://localhost:1317  --node tcp://localhost:46657 --trust-node=false
```

## LCD Use Cases

LCD could be very helpful for related service providers. For a wallet service provider, LCD could make transaction faster and more reliable in the following cases. 

1. User Creates An Account

![deposit](https://github.com/irisnet/cosmos-sdk/raw/bianjie/lcd-spec/docs/spec/lcd/pics/create-account.png)

1. User Makes a Transfer

![withdraw](https://github.com/irisnet/cosmos-sdk/raw/bianjie/lcd-spec/docs/spec/lcd/pics/transfer.png)
