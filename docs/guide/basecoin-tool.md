# The Basecoin Tool

In previous tutorials we learned the [basics of the `basecoin` CLI](/docs/guides/basecoin-basics)
and [how to implement a plugin](/docs/guides/example-plugin).
In this tutorial, we provide more details on using the `basecoin` tool.

# Data Directory

By default, `basecoin` works out of `~/.basecoin`. To change this, set the `BCHOME` environment variable:

```
export BCHOME=~/.my_basecoin_data
basecoin init
basecoin start
```

or 

```
BCHOME=~/.my_basecoin_data basecoin init
BCHOME=~/.my_basecoin_data basecoin start
```

# ABCI Server

So far we have run Basecoin and Tendermint in a single process.
However, since we use ABCI, we can actually run them in different processes.
First, initialize them:

```
basecoin init
```

This will create a single `genesis.json` file in `~/.basecoin` with the information for both Basecoin and Tendermint.

Now, In one window, run 

```
basecoin start --without-tendermint
```

and in another,

```
TMROOT=~/.basecoin tendermint node
```

You should see Tendermint start making blocks!

Alternatively, you could ignore the Tendermint details in `~/.basecoin/genesis.json` and use a separate directory by running:

```
tendermint init
tendermint node
```

For more details on using `tendermint`, see [the guide](https://tendermint.com/docs/guides/using-tendermint).

# Keys and Genesis

In previous tutorials we used `basecoin init` to initialize `~/.basecoin` with the default configuration.
This command creates files both for Tendermint and for Basecoin, and a single `genesis.json` file for both of them.
For more information on these files, see the [guide to using tendermint](https://tendermint.com/docs/guides/using-tendermint).

Now let's make our own custom Basecoin data.

First, create a new directory:

```
mkdir example-data
```

We can tell `basecoin` to use this directory by exporting the `BCHOME` environment variable:

```
export BCHOME=$(pwd)/example-data
```

If you're going to be using multiple terminal windows, make sure to add this variable to your shell startup scripts (eg. `~/.bashrc`).

Now, let's create a new private key:

```
basecoin key new > $BCHOME/key.json
```

Here's what my `key.json looks like (TODO: change `keys` so it looks like this ...):

```json
{
        "address": "4EGEhnqOw/gX326c7KARUkY1kic=",
        "pub_key": {
                "type": "ed25519",
                "data": "a20d48b5caff42892d0ac67ccdeee38c1dcbbe42b15b486057d16244541e8141"
        },
        "priv_key": {
                "type": "ed25519",
                "data": "654c845f4b36d1a881deb0ff09381165d3ccd156b4aabb5b51267e91f1d024a5a20d48b5caff42892d0ac67ccdeee38c1dcbbe42b15b486057d16244541e8141"
        }
}
```

Yours will look different - each key is randomly derrived.

Now we can make a `genesis.json` file and add an account with our public key:

```json
{
  "chain_id": "example-chain",
  "app_options": {
    "accounts": [{
      "pub_key": {
        "type": "ed25519",
        "data": "a20d48b5caff42892d0ac67ccdeee38c1dcbbe42b15b486057d16244541e8141"
      },
      "coins": [
        {
          "denom": "gold",
          "amount": 1000000000
        }
      ]
    }]
  }
}
```

Here we've granted ourselves `1000000000` units of the `gold` token.
Note that we've also set the `chain_id` to be `example-chain`.
All transactions must therefore include the `--chain_id example-chain` in order to make sure they are valid for this chain.
Previously, we didn't need this flag because we were using the default chain ID ("test_chain_id").
Now that we're using a custom chain, we need to specify the chain explicitly on the command line.

Note we have also left out the details of the tendermint genesis. These are documented in the [tendermint guide](https://tendermint.com/docs/guides/using-tendermint).


# Reset

You can reset all blockchain data by running:

```
basecoin unsafe_reset_all
```
