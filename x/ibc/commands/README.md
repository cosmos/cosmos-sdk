# IBC CLI Usage

## initialize

```bash
basecoind init # copy the recover key
basecli keys add keyname --recover
basecoind start
```

## transfer

`transfer` sends coins from one chain to another(or itself).

```bash
basecli transfer --name keyname --to address_of_destination --amount 10mycoin --chain test-chain-AAAAAA --chain-id AAAAAA
```

The id of the chain can be found in `$HOME/.basecoind/config/genesis.json`

## relay

```bash
basecli relay --name keyname --from-chain-id test-chain-AAAAAA --from-chain-node=tcp://0.0.0.0:46657 --to-chain-id test-chain-AAAAAA --to-chain-node=tcp://0.0.0.0:46657
```
