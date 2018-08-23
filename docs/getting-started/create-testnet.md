## Create your Own Testnet

To create your own testnet, first each validator will need to install gaiad and run gen-tx

```bash
gaiad init gen-tx --name <account_name>
```

This populations `$HOME/.gaiad/gen-tx/` with a json file.

Now these json files need to be aggregated together via Github, a Google form, pastebin or other methods.

Place all files on one computer in `$HOME/.gaiad/gen-tx/`

```bash
gaiad init --gen-txs -o --chain=<chain-name>
```

This will generate a `genesis.json` in `$HOME/.gaiad/config/genesis.json` distribute this file to all validators on your testnet.

### Export state

To export state and reload (useful for testing purposes):

```
gaiad export > genesis.json; cp genesis.json ~/.gaiad/config/genesis.json; gaiad start
```
