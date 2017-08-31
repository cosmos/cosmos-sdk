# API

In order to allow for quick client development, with the security
of the light-client and ease of use the cli (embedding go-wire and
go-crypto), we will provide a REST API, which can run on localhost
as part of `basecli`, securely managing your keys, signing transactions,
and validating queries.  It will be exposed for a locally deployed
application to make use of (eg. electron app, web server...)

By default we will serve on `http://localhost:2024`.  CORS will be disabled by default.  Potentially we can add a flag to enable it for one domain.

## MVP

The MVP will allow us to move around money.  This involves the
following functions:

## Construct an unsigned transaction

`POST /build/send`

Input:
```
{
    "to": {"app": "role", "addr": "62616E6B32" },
    "from": {"app": "sigs", "addr": "BDADF167E6CF2CDF2D621E590FF1FED2787A40E0" },
    "amount": { "denom": "mycoin", "amount": 900000 },
    "sequence": 1,
    "multi": true,
}
```

Output (a json encoding of basecoin.Tx):

`basecli tx send --to=role:62616E6B32 --from=sigs:91C959ADE03D8973E8F2FBA9FD2EED327DCE2B0A --amount=900000mycoin  --sequence=1 --multi --prepare=- --no-sign`


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
            "sequence": 1,
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
    "signatures": null
  }
}
```

## Sign a Tx

Once you construct a proper json-encoded `basecoin.Tx`, you can sign it once, or (if you constructed it with `multi=true`), multiple times.


`POST /sign`

Input:

```
{
    "name": "matt",
    "password": "1234567890",
    "tx": {
        "type": "sigs/multi",
        "data": {
            "tx": // see output of /build/send,
            "signatures": nil,
        }
    }
}
```

Output:

`basecli tx send --to=role:62616E6B32 --from=sigs:91C959ADE03D8973E8F2FBA9FD2EED327DCE2B0A --amount=900000mycoin  --sequence=1 --multi --no-sign --prepare=unsigned.json`

`echo 1234567890 | basecli tx --in=unsigned.json --prepare=- --name=matt`

```
{
    "type": "sigs/multi",
    "data": {
        "tx": // see output of /build/send,
        "signatures": [
            {
                "Sig": {
                    "type": "ed25519",
                    "data": "436188FAC4668DDF6729022454AFBA5DA0B44E516C4EC7013C6B00BD877F255CDE0355F3FBFE9CCF88C9F519C192D498BF087AFE0D531351813432A100857803"
                },
                "Pubkey": {
                    "type": "ed25519",
                    "data": "B01508EB073C0823E2CE6ABF4538BA02EAEC39B02113290BBFCEC7E1B07F575A"
                }
            }
        ]
    }
}
```

## Send Tx to the Blockchain

This will encode the transaction as binary and post it to the tendermint node, waiting until it is committed to the blockchain.
(TODO: make this async? return when recevied, notify when committed?)

`POST /tx`

Input:

Signed tx as json, directly copy output of `/sign`

Output:


`echo 1234567890 | basecli tx send --to=role:62616E6B32 --from=sigs:91C959ADE03D8973E8F2FBA9FD2EED327DCE2B0A --amount=900000mycoin  --sequence=1 --multi --name=matt --prepare=signed.json`

`basecli tx --in=signed.json --no-sign`

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

## Query account balance

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

## Other stuff for MVP

You must run `basecli init` int he cli to set things up.

When you run `basecli serve`, it will start the local rest api server, with the above endpoints.

Also, support keys endpoints from go-crypto as they currently are and mount them under `/keys`.

## Future Stuff

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
