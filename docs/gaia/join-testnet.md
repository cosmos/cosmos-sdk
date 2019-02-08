# Join the public testnet

::: tip Current Testnet
See the [testnet repo](https://github.com/cosmos/testnets) for
information on the latest testnet, including the correct version
of the Cosmos-SDK to use and details about the genesis file.
:::

::: warning
**You need to [install gaia](./installation.md) before you go further**
:::

## Setting Up a New Node

> NOTE: If you ran a full node on a previous testnet, please skip to [Upgrading From Previous Testnet](#upgrading-from-previous-testnet).

These instructions are for setting up a brand new full node from scratch.

First, initialize the node and create the necessary config files:

```bash
gaiad init <your_custom_moniker>
```

::: warning Note
Monikers can contain only ASCII characters. Using Unicode characters will render your node unreachable.
:::

You can edit this `moniker` later, in the `~/.gaiad/config/config.toml` file:

```toml
# A custom human readable name for this node
moniker = "<your_custom_moniker>"
```

You can edit the `~/.gaiad/config/gaiad.toml` file in order to enable the anti spam mechanism and reject incoming transactions with less than a minimum fee:

```
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# Validators reject any tx from the mempool with less than the minimum fee per gas.
minimum_fees = ""
```


Your full node has been initialized! Please skip to [Genesis & Seeds](#genesis-seeds).

## Upgrading From Previous Testnet

These instructions are for full nodes that have ran on previous testnets and would like to upgrade to the latest testnet.

### Reset Data

First, remove the outdated files and reset the data.

```bash
rm $HOME/.gaiad/config/addrbook.json $HOME/.gaiad/config/genesis.json
gaiad unsafe-reset-all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before,
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

::: danger Warning
Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to double sign.
:::

### Software Upgrade

Now it is time to upgrade the software:

```bash
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
git fetch --all && git checkout master
make update_tools install
```

::: tip
*NOTE*: If you have issues at this step, please check that you have the latest stable version of GO installed.
:::

Note we use `master` here since it contains the latest stable release.
See the [testnet repo](https://github.com/cosmos/testnets)
for details on which version is needed for which testnet,
and the [SDK release page](https://github.com/cosmos/cosmos-sdk/releases)
for details on each release.

Your full node has been cleanly upgraded!

## Genesis & Seeds

### Copy the Genesis File

Fetch the testnet's `genesis.json` file into `gaiad`'s config directory.

```bash
mkdir -p $HOME/.gaiad/config
curl https://raw.githubusercontent.com/cosmos/testnets/master/latest/genesis.json > $HOME/.gaiad/config/genesis.json
```

Note we use the `latest` directory in the [testnets repo](https://github.com/cosmos/testnets)
which contains details for the latest testnet. If you are connecting to a different testnet, ensure you get the right files.

To verify the correctness of the configuration run:

```bash
gaiad start
```

### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.gaiad/config/config.toml`. The `testnets` repo contains links to the seed nodes for each testnet. If you are looking to join the running testnet please [check the repository for details](https://github.com/cosmos/testnets) on which nodes to use.

If those seeds aren't working, you can find more seeds and persistent peers on the [Cosmos Explorer](https://explorer.cosmos.network/nodes). Open the the `Full Nodes` pane and select nodes that do not have private (`10.x.x.x`) or [local IP addresses](https://en.wikipedia.org/wiki/Private_network). The `Persistent Peer` field contains the connection string. For best results use 4-6.

You can also ask for peers on the [Validators Riot Room](https://riot.im/app/#/room/#cosmos-validators:matrix.org)

For more information on seeds and peers, you can [read this](https://github.com/tendermint/tendermint/blob/develop/docs/tendermint-core/using-tendermint.md#peers).

## Run a Full Node

Start the full node with this command:

```bash
gaiad start
```

Check that everything is running smoothly:

```bash
gaiacli status
```

View the status of the network with the [Cosmos Explorer](https://explorecosmos.network). Once your full node syncs up to the current block height, you should see it appear on the [list of full nodes](https://explorecosmos.network/validators). If it doesn't show up, that's ok--the Explorer does not connect to every node.

## Export State

Gaia can dump the entire application state to a JSON file, which could be useful for manual analysis and can also be used as the genesis file of a new network.

Export state with:

```bash
gaiad export > [filename].json
```

You can also export state from a particular height (at the end of processing the block of that height):

```bash
gaiad export --height [height] > [filename].json
```

If you plan to start a new network from the exported state, export with the `--for-zero-height` flag:

```bash
gaiad export --height [height] --for-zero-height > [filename].json
```

## Upgrade to Validator Node

You now have an active full node. What's the next step? You can upgrade your full node to become a Cosmos Validator. The top 100 validators have the ability to propose new blocks to the Cosmos Hub. Continue onto [the Validator Setup](./validators/validator-setup.md).
