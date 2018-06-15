# SDK Client

## Instruction

### Remove remaining data(if exists)

```bash
> rm -r ~/.gaiad
> rm -r ~/.gaiacli
```

### Installation

```bash
> go get -u github.com/cosmos/cosmos-sdk
> cd $GOPATH/src/github.com/cosmos/cosmos-sdk
> git checkout develop
> dep ensure
> make install
```

### Run node and server

If you are going to connect to an external node(e.g. testnet), skip this part

```bash

> gaiad init --name myname
> gaiad start
> gaiacli advanced rest-server
```

### Test getting account

```bash
> cat $HOME/.gaiad/config/genesis.json
...
"app_state": {
    "accounts": [
      {
        "address": "1E95740A4F7B21CDD9B3DF84C8FF1BC697173B70",
        "coins": [
          {
            "denom": "mynameToken",
            "amount": 1000
          },
          {
            "denom": "steak",
            "amount": 50
          }
        ]
      }
    ],
    ...
}
> curl localhost:1317/accounts/1E95740A4F7B21CDD9B3DF84C8FF1BC697173B70 # or access from your web browser
{"type":"6C54F73C9F2E08","value":{"address":"1E95740A4F7B21CDD9B3DF84C8FF1BC697173B70","coins":[{"denom":"mynameToken","amount":1000},{"denom":"steak","amount":50}],"public_key":null,"sequence":0}}
```

## Notable Flags

* `--chain-id <STRING>`: chain id of the full node to connect
* `--node <URL>`: address of the full node to connect
* `--laddr <URL>`: address to run the rest server on
* `--trust-node`: flag for disabling merkle proof, default true before launch
