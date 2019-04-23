# Upgrade Your Node

::: warning
The detailed procedure to upgrade a mainnet node from `cosmoshub-1` to `cosmoshub-2` can be found [here](https://gist.github.com/alexanderbez/5e87886221eb304b9e85ad4b167c99c8). 
:::

This document describes the upgrade procedure of a `gaiad` full-node to a new version.

## Software Upgrade

First, stop your instance of `gaiad`. Next, upgrade the software:

```bash
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
git fetch --all && git checkout <new_version>
make tools install
```

::: tip
*NOTE*: If you have issues at this step, please check that you have the latest stable version of GO installed.
:::

See the [testnet repo](https://github.com/cosmos/testnets) for details on which version is needed for which public testnet, and the [SDK release page](https://github.com/cosmos/cosmos-sdk/releases) for details on each release.

Your full node has been cleanly upgraded!

## Fetch new genesis 

:::warning 
If the new version you are upgrading to has breaking changes, you will have to restart your chain. If it is not breaking, you can skip to [Restart](#restart)
:::

The first step is to remove your current genesis:

```bash
rm $HOME/.gaiad/config/addrbook.json $HOME/.gaiad/config/genesis.json
```

The procedure varies depending on the network you want to connect to. 

### Mainnet

Follow the [official upgrade guide](https://gist.github.com/alexanderbez/5e87886221eb304b9e85ad4b167c99c8). 

### Public Testnet

If you are joining a new public testnet, fetch the genesis from the appropriate testnet in the [testnet repo](https://github.com/cosmos/testnets). Save the new genesis as `new_genesis.json`. Then replace the old `genesis.json` with `new_genesis.json`

```bash
cd $HOME/.gaiad/config
cp -f genesis.json new-_enesis.json
mv new_genesis.json genesis.json
```

### Local Testnet

If you are running your own local testnet, you can either start with a brand new genesis using `gaiad init`, or export the state from you previous network as a new genesis. To do so, use the following command

```bash
cd $HOME/.gaiad/config
gaiad export --for-zero-height --height=<export-height> > new_genesis.json
```

Then, replace the old `genesis.json` with `new_genesis.json`.

```bash
cp -f genesis.json new-_enesis.json
mv new_genesis.json genesis.json
```

## Reset Data

:::warning 
If the version <new_version> you are upgrading to is not breaking from the previous one, you should not reset the data. If it is not breaking, you can skip to [Restart](#restart)
:::

::: warning 
If you are running a **validator node** on the mainnet, always be careful when doing `gaiad unsafe-reset-all`. You should never use this command if you are not switching `chain-id`.
:::

::: danger IMPORTANT
Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to get slashed due to double sign !
:::

First, remove the outdated files and reset the data. **If you are running a validator node, make sure you understand what you are doing before resetting**. 

```bash
gaiad unsafe-reset-all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before, your node will still try to connect to them, but may fail if they haven't also been upgraded.

## Restart

To restart your node, just type:

```bash
gaiad start
```
