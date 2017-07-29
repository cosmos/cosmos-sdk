# baseserver

baseserver is the REST counterpart to basecli

## Compiling and running it
```shell
$ go get -u -v github.com/tendermint/basecoin/cmd/baseserver
$ baseserver init
$ baseserver serve --port 8888
```

to run the server at localhost:8888, otherwise if you don't specify --port,
by default the server will be run on port 8998.

## Supported routes
Route | Method | Completed | Description
---|---|---|---
/keys|GET|✔️|Lists all keys
/keys|POST|✔️|Generate a new key. It expects fields: "name", "algo", "passphrase"
/keys/{name}|GET|✔️|Retrieves the specific key
/keys/{name}|POST/PUT|✔️|Updates the named key
/keys/{name}|DELETE|✔️|Deletes the named key
/build/send|POST|✔️|Send a transaction
/sign|POST|✔️|Sign a transaction
/tx|POST|✖️|Post a transaction to the blockchain
/seeds/status|GET|✖️|Returns the information on the last seed

## Sample usage
- Generate a key
```shell
$ curl -X POST http://localhost8998/keys --data '{"algo": "ed25519", "name": "SampleX", "passphrase": "Say no more"}'
```

```json
{
  "key": {
    "name": "SampleX",
    "address": "603EE63C41E322FC7A247864A9CD0181282EB458",
    "pubkey": {
      "type": "ed25519",
      "data": "C050948CFC087F5E1068C7E244DDC30E03702621CC9442A28E6C9EDA7771AA0C"
    }
  },
  "seed_phrase": "border almost future parade speak soccer bulk orange real brisk caution body river chapter"
}
```

- Sign a key
```shell
$ curl -X POST http://localhost:8998/sign --data '{
    "name": "matt",
    "password": "Say no more",
    "tx": {
        "type": "sigs/multi",
        "data": {
            "tx": {"type":"coin/send","data":{"inputs":[{"address":{"chain":"","app":"role","addr":"62616E6B32"},"coins":[{"denom":"mycoin","amount":900000}]}],"outputs":[{"address":{"chain":"","app":"sigs","addr":"BDADF167E6CF2CDF2D621E590FF1FED2787A40E0"},"coins":[{"denom":"mycoin","amount":900000}]}]}},
            "signatures": null
        }
    }
}'
```

```json
{
  "type": "sigs/multi",
  "data": {
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
    },
    "signatures": [
      {
        "Sig": {
          "type": "ed25519",
          "data": "F6FE3053F1E6C236F886A0D525C1AF840F7831B6E50F7E1108C345AA524303920F09945DA110AD5184B3F45717D7114E368B12AFE027FECECC2FC193D4906A0C"
        },
        "Pubkey": {
          "type": "ed25519",
          "data": "0D8D19E527BAE9D1256A3D03009E2708171CDCB71CCDEDA2DC52DD9AD23AEE25"
        }
      }
    ]
  }
}
```
