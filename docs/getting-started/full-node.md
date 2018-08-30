# Join the Testnet

Please ensure you have the [Cosmos SDK](/getting-started/installation.md) installed. If you ran a full node on a previous testnet, please skip to [Upgrading From Previous Testnet](#upgrading-from-previous-testnet).

## Setting Up a New Node

These instructions are for setting up a brand new full node from scratch. 

First, initialize the node and create the necessary config files:

```bash
gaiad init --name <your_custom_name>
```

::: warning Note
Only ASCII characters are supported for the `--name`. Using Unicode characters will render your node unreachable.
:::

You can edit this `name` later, in the `~/.gaiad/config/config.toml` file:

```toml
# A custom human readable name for this node
moniker = "<your_custom_name>"
```

Your full node has been initialized! Please skip to [Genesis & Seeds](#genesis-seeds).

## Upgrading From Previous Testnet

These instructions are for full nodes that have ran on previous testnets and would like to upgrade to the latest testnet.

### Reset Data

First, remove the outdated files and reset the data.

```bash
rm $HOME/.gaiad/config/addrbook.json $HOME/.gaiad/config/genesis.json
gaiad unsafe_reset_all
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
git fetch --all && git checkout v0.23.1
make update_tools && make get_vendor_deps && make install
```

Your full node has been cleanly upgraded!

## Genesis & Seeds

### Copy the Genesis File

Fetch the testnet's `genesis.json` file into `gaiad`'s config directory.

```bash
mkdir -p $HOME/.gaiad/config
curl https://raw.githubusercontent.com/cosmos/testnets/master/gaia-7002/genesis.json > $HOME/.gaiad/config/genesis.json
```

### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.gaiad/config/config.toml`. Here are some seed nodes you can use:

```toml
# Comma separated list of seed nodes to connect to
seeds = "718145d422a823fd2a4e1e36e91b92bb0c4ddf8e@gaia-testnet.coinculture.net:26656,5922bf29b48a18c2300b85cc53f424fce23927ab@67.207.73.206:26656,7c8b8fd03577cd4817f5be1f03d506f879df98d8@gaia-7000-seed1.interblock.io:26656,a28737ff02391a6e00a1d3b79befd57e68e8264c@gaia-7000-seed2.interblock.io:26656,987ffd26640cd03d08ed7e53b24dfaa7956e612d@gaia-7000-seed3.interblock.io:26656"
```

If those seeds aren't working, you can find more seeds and persistent peers on the [Cosmos Explorer](https://explorecosmos.network/nodes). Open the the `Full Nodes` pane and select nodes that do not have private (`10.x.x.x`) or [local IP addresses](https://en.wikipedia.org/wiki/Private_network). The `Persistent Peer` field contains the connection string. For best results use 4-6.

For more information on seeds and peers, you can [read this](https://github.com/tendermint/tendermint/blob/develop/docs/using-tendermint.md#peers).

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


## Upgrade to Validator Node

You now have an active full node. What's the next step? You can upgrade your full node to become a Cosmos Validator. The top 100 validators have the ability to propose new blocks to the Cosmos Hub. Continue onto [the Validator Setup](../validators/validator-setup.md).
