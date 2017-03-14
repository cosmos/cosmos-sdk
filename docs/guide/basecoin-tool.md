# The Basecoin Tool

In previous tutorials we learned the [basics of the `basecoin` CLI](/docs/guides/basecoin-basics)
and [how to implement a plugin](/docs/guides/example-plugin).
In this tutorial, we provide more details on using the `basecoin` tool.

# ABCI Server

So far we have run Basecoin and Tendermint in a single process.
However, since we use ABCI, we can actually run them in different processes.
First, initialize both Basecoin and Tendermint:

```
basecoin init
tendermint init
```

In one window, run 

```
basecoin start --abci-server
```

and in another,

```
tendermint node
```

You should see Tendermint start making blocks!


# Keys and Genesis

In previous tutorials we used `basecoin init` to initialize `~/.basecoin` with the default configuration.
This command creates files both for Tendermint and for Basecoin.
The Tendermint files are stored in `~/.basecoin/tendermint`, and are the same type of files that would exist in `~/.tendermint` after running `tendermint init`.
For more information on these files, see the [guide to using tendermint](https://tendermint.com/docs/guides/using-tendermint).

Now let's make our own custom Basecoin data.

First, create a new directory:

```
mkdir example-data
```

We can tell `basecoin` to use this directory by exporting the `BASECOIN_ROOT` environment variable:

```
export BASECOIN_ROOT=$(pwd)/example-data
```

If you're going to be using multiple terminal windows, make sure to add this variable to your shell startup scripts (eg. `~/.bashrc`).

Now, let's create a new private key:

```
basecoin key new > $BASECOIN_ROOT/key.json
```

Here's what my `key.json looks like:

```json
{
	"address": "15F591CA434CFCCBDEC1D206F3ED3EBA207BFE7D",
	"priv_key": [
		1,
		"737C629667A9EAADBB8E7CF792D5A8F63AA4BB51E06457DDD7FDCC6D7412AAAD43AA6C88034F9EB8D2717CA4BBFCBA745EFF19B13EFCD6F339EDBAAAFCD2F7B3"
	],
	"pub_key": [
		1,
		"43AA6C88034F9EB8D2717CA4BBFCBA745EFF19B13EFCD6F339EDBAAAFCD2F7B3"
	]
}
```

Yours will look different - each key is randomly derrived.

Now we can make a `genesis.json` file and add an account with our public key:

```json
[
  "base/chainID", "example-chain",
  "base/account", {
    "pub_key": [1, "43AA6C88034F9EB8D2717CA4BBFCBA745EFF19B13EFCD6F339EDBAAAFCD2F7B3"],
    "coins": [
	{
	  "denom": "gold",
	  "amount": 1000000000,
	}
    ]
  }
]
```

Here we've granted ourselves `1000000000` units of the `gold` token.
Note that we've also set the `base/chainID` to be `example-chain`.
All transactions must therefore include the `--chain_id example-chain` in order to make sure they are valid for this chain.
Previously, we didn't need this flag because we were using the default chain ID ("test_chain_id").
Now that we're using a custom chain, we need to specify the chain explicitly on the command line.


# Reset

You can reset all blockchain data by running:

```
basecoin unsafe_reset_all
```
