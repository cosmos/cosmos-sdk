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
package github.com/cosmos/cosmos-sdk: no Go files in /home/mossid/go/src/github.com/cosmos/cosmos-sdk
> cd $GOPATH/src/github.com/cosmos/cosmos-sdk
> git checkout develop
Branch develop set up to track remote branch develop from origin.
Switched to a new branch 'develop'
> dep ensure
> make install
go install -ldflags "-X github.com/cosmos/cosmos-sdk/version.GitCommit=be7ec5b" ./cmd/gaia/cmd/gaiad
go install -ldflags "-X github.com/cosmos/cosmos-sdk/version.GitCommit=be7ec5b" ./cmd/gaia/cmd/gaiacli
```

### Run node and server

```bash
# If your are going to connect to an external node(e.g. testnet), skip gaiad init and gaiad start
> gaiad init --name myname
{
  "chain_id": "test-chain-vFxi71",
  "node_id": "ef0dcdf3b4c90b9f5a1f09975730ae7708861ed6",
  "app_message": null
}
> gaiad start
...
# If your are going to connect to an external node, speficy node address with --node flag
> gaiacli advanced rest-server
I[06-05|04:26:30.739] Starting RPC HTTP server on tcp://localhost:1317 module=rest-server 
I[06-05|04:26:30.739] REST server started                          module=rest-server
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
...
> curl localhost:1317/accounts/1E95740A4F7B21CDD9B3DF84C8FF1BC697173B70 # or access from your web browser
{"type":"6C54F73C9F2E08","value":{"address":"1E95740A4F7B21CDD9B3DF84C8FF1BC697173B70","coins":[{"denom":"mynameToken","amount":1000},{"denom":"steak","amount":50}],"public_key":null,"sequence":0}}
```

## Notable Flags

* `--chain-id <STRING>`: chain id of the full node to connect
* `--node <URL>`: address of the full node to connect
* `--laddr <URL>`: address to run the rest server on
