# IBC Doubble Hubble

## Remove remaining data

```console
> rm -r ~/.chain1
> rm -r ~/.chain2
> rm -r ~/.basecli
```

## Initialize both chains

```console
> basecoind init --home ~/.chain1
I[04-02|14:03:33.704] Generated private validator                  module=main path=/home/mossid/.chain1/config/priv_validator.json
I[04-02|14:03:33.705] Generated genesis file                       module=main path=/home/mossid/.chain1/config/genesis.json
{
  "secret": "crunch ignore trigger neither differ dance cheap brick situate floor luxury citizen husband decline arrow abandon",
  "account": "C69FEB398A29AAB1B3C4F07DE22208F35E711BCC",
  "validator": {
    "pub_key": {
      "type": "ed25519",
      "data": "8C9917D5E982E221F5A1450103102B44BBFC1E8768126C606246CB37B5794F4D"
    },
    "power": 10,
    "name": ""
  },
  "node_id": "3ac8e6242315fd62143dc3e52c161edaaa6b1a64",
  "chain_id": "test-chain-ZajMfr"
}
> ADDR1=C69FEB398A29AAB1B3C4F07DE22208F35E711BCC
> ID1=test-chain-ZajMfr
> NODE1=tcp://0.0.0.0:36657
> basecli keys add key1 --recover
Enter a passphrase for your key:
Repeat the passphrase:
Enter your recovery seed phrase:
crunch ignore trigger neither differ dance cheap brick situate floor luxury citizen husband decline arrow abandon
key1        C69FEB398A29AAB1B3C4F07DE22208F35E711BCC


> basecoind init --home ~/.chain2
I[04-02|14:09:14.453] Generated private validator                  module=main path=/home/mossid/.chain2/config/priv_validator.json
I[04-02|14:09:14.453] Generated genesis file                       module=main path=/home/mossid/.chain2/config/genesis.json
{
  "secret": "age guide awesome month female left oxygen soccer define high grocery work desert dinner arena abandon",
  "account": "DC26002735D3AA9573707CFA6D77C12349E49868",
  "validator": {
    "pub_key": {
      "type": "ed25519",
      "data": "A94FE4B9AD763D301F4DD5A2766009812495FB7A79F1275FB8A5AF09B44FD5F3"
    },
    "power": 10,
    "name": ""
  },
  "node_id": "ad26831330e1c72b85276d53c20f0680e6fd4cf5"
  "chain_id": "test-chain-4XHTPn"
}
> ADDR2=DC26002735D3AA9573707CFA6D77C12349E49868
> ID2=test-chain-4XHTPn
> NODE2=tcp://0.0.0.0:46657
> basecli keys add key2 --recover
Enter a passphrase for your key:
Repeat the passphrase:
Enter your recovery seed phrase:
age guide awesome month female left oxygen soccer define high grocery work desert dinner arena abandon
key2        DC26002735D3AA9573707CFA6D77C12349E49868


> basecoind start --home ~/.chain1 --address tcp://0.0.0.0:36658 --rpc.laddr tcp://0.0.0.0:36657 --p2p.laddr tcp://0.0.0.0:36656
...

> basecoind start --home ~/.chain2 # --address tcp://0.0.0.0:46658 --rpc.laddr tcp://0.0.0.0:46657 --p2p.laddr tcp://0.0.0.0:46656
...
```
## Check balance

```console
> basecli account $ADDR1 --node $NODE1
{
  "address": "C69FEB398A29AAB1B3C4F07DE22208F35E711BCC",
  "coins": [
    {
      "denom": "mycoin",
      "amount": 9007199254740992
    }
  ],
  "public_key": null,
  "sequence": 0,
  "name": ""
}

> basecli account $ADDR2 --node $NODE2
{
  "address": "DC26002735D3AA9573707CFA6D77C12349E49868",
  "coins": [
    {
      "denom": "mycoin",
      "amount": 9007199254740992
    }
  ],
  "public_key": null,
  "sequence": 0,
  "name": ""
}

```

## Transfer coins (addr1:chain1 -> addr2:chain2)

```console
> basecli transfer --name key1 --to $ADDR2 --amount 10mycoin --chain $ID2 --chain-id $ID1 --node $NODE1
Password to sign with 'key1':
Committed at block 1022. Hash: E16019DCC4AA08CA70AFCFBC96028ABCC51B6AD0
> basecli account $ADDR1 --node $NODE1
{
  "address": "C69FEB398A29AAB1B3C4F07DE22208F35E711BCC",
  "coins": [
    {
      "denom": "mycoin",
      "amount": 9007199254740982
    }
  ],
  "public_key": {
    "type": "ed25519",
    "data": "9828FF1780A066A0D93D840737566B697035448D6C880807322BED8919348B2B"
  },
  "sequence": 1,
  "name": ""
}
```

## Relay IBC packets

```console
> basecli relay --name key2 --from-chain-id $ID1 --from-chain-node $NODE1 --to-chain-id $ID2 --to-chain-node $NODE2 --chain-id $ID2
Password to sign with 'key2':
I[04-03|16:18:59.984] Detected IBC packet                          number=0
I[04-03|16:19:00.869] Relayed IBC packet                           number=0
> basecli account $ADDR2 --node $NODE2
{
  "address": "DC26002735D3AA9573707CFA6D77C12349E49868",
  "coins": [
    {
      "denom": "mycoin",
      "amount": 9007199254741002
    }
  ],
  "public_key": {
    "type": "ed25519",
    "data": "F52B4FA545F4E9BFE5D7AF1DD2236899FDEF905F9B3057C38D7C01BF1B8EB52E"
  },
  "sequence": 1,
  "name": ""
}

```
