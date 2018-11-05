# Getting Started

To start a REST server, we need to specify the following parameters:

| Parameter   | Type      | Default                 | Required | Description                                          |
| ----------- | --------- | ----------------------- | -------- | ---------------------------------------------------- |
| chain-id    | string    | null                    | true     | chain id of the full node to connect                 |
| node        | URL       | "tcp://localhost:46657" | true     | address of the full node to connect                  |
| laddr       | URL       | "tcp://localhost:1317"  | true     | address to run the rest server on                    |
| trust-node  | bool      | "false"                 | true     | Whether this LCD is connected to a trusted full node |
| trust-store | DIRECTORY | "$HOME/.lcd"            | false    | directory for save checkpoints and validator sets    |

For example::

```bash
gaiacli rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=false
```

The server listens on HTTPS by default. You can set the SSL certificate to be used by the server with these additional flags:

```bash
gaiacli rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=false \
    --certfile=mycert.pem --keyfile=mykey.key
```

If no certificate/keyfile pair is supplied, a self-signed certificate will be generated and its fingerprint printed out.
Append `--insecure` to the command line if you want to disable the secure layer and listen on an insecure HTTP port.

For more information about the Gaia-Lite RPC, see the [swagger documentation](https://cosmos.network/rpc/)
