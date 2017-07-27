# API

Serve on `http://localhost:2024`

Don't allow cors

## Create Tx

`POST /tx`

Input: signed tx as json, just like --prepare...
Goes to the node, and returns result:

```
{
  "check_tx": {
    "code": 0,
    "data": "",
    "log": ""
  },
  "deliver_tx": {
    "code": 0,
    "data": "",
    "log": ""
  },
  "hash": "4D0EB7853E71AB6E3021990CF733F70F4CC2E001",
  "height": 1494
}
```

`POST /sign`

Input:

{
    "name": "matt",
    "password": "1234567890",
    "tx": {
        // just like prepare, tx (with sigs/multi but no sigs)
    }
}

Example tx:

wrappers = combine(feeWrap, nonceWrap, chainWrap, sigWrap)

UI will have to know about feeWrap to pass info in

```
{
  "type": "sigs/multi",
  "data": {
    "tx": {
      "type": "chain/tx",
      "data": {
        "chain_id": "lakeshore",
        "expires_at": 0,
        "tx": {
          "type": "nonce",
          "data": {
            "sequence": 3,
            "signers": [
              {
                "chain": "",
                "app": "sigs",
                "addr": "91C959ADE03D8973E8F2FBA9FD2EED327DCE2B0A"
              }
            ],
            "tx": {
              "type": "coin/send",
              "data": {
                "inputs": [
                  {
                    "address": {
                      "chain": "",
                      "app": "role",
                      "addr": "62616E6B32"
                    },
                    "coins": [
                      {
                        "denom": "mycoin",
                        "amount": 900000
                      }
                    ]
                  }
                ],
                "outputs": [
                  {
                    "address": {
                      "chain": "",
                      "app": "sigs",
                      "addr": "BDADF167E6CF2CDF2D621E590FF1FED2787A40E0"
                    },
                    "coins": [
                      {
                        "denom": "mycoin",
                        "amount": 900000
                      }
                    ]
                  }
                ]
              }
            }
          }
        }
      }
    },
    "signatures": []
  }
}
```

Matt's proposal:

`POST /build/send`

Input:
```
{
    "fees": {"denom": "atom", "amount": 23},
    "to": {"app": "role", "addr": "62616E6B32" },
    "from": {"app": "sigs", "addr": "BDADF167E6CF2CDF2D621E590FF1FED2787A40E0" },
    "amount": { "denom": "mycoin", "amount": 900000 },
    "multi": true,
}
```

Output: the input for /sign

`POST /build/create-role`

Input:
```
{
    "name": "trduvtqicrtqvrqy",
    "min_sigs": 2,
    "members": [
        {"app": "sigs", "addr": "BDADF167E6CF2CDF2D621E590FF1FED2787A40E0" },
        {"app": "sigs", "addr": "91C959ADE03D8973E8F2FBA9FD2EED327DCE2B0A" }
    ]
}
```


## Query:

`GET /query/account/sigs:BDADF167E6CF2CDF2D621E590FF1FED2787A40E0`

```
{
  "height": 1170,
  "data": {
    "coins": [
      {
        "denom": "mycoin",
        "amount": 12345
      }
    ]
  }
}
```


## Other stuff

`init` / `serve` on cli

keys management under `/keys`:

List, get, create (delete + update)?

proxy mounted as well under `/tendermint`

`/tendermint/status`
`/tendermint/block`

info about self...

`/`

```
{
    "app": "basecli",
    "version": "0.7.1",  // of client, server????
    "modules": {
        "chain": "0.1.0",
        "fees": "0.2.1",
        "coin": "0.3.2",
        "stake": "0.1.2"
    },
    "nodes": [
        "localhost:46657",
        "mercury.interchain.io:443"
    ]
}
```

`/seeds`

```
{
    "last_height": 4555,
    "update_problems": "too much change"
}
```

info on last seed
