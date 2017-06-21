# The Basecoin Tool

In previous tutorials we learned the [basics of the Basecoin
CLI](/docs/guide/basecoin-basics.md) and [how to implement a
plugin](/docs/guide/basecoin-plugins.md).  In this tutorial, we provide more
details on using the Basecoin tool.

# Generate a Key

Generate a key using the `basecli` tool:

```
basecli keys new mykey
ME=$(basecli keys get mykey | awk '{print $2}')
```

# Data Directory

By default, `basecoin` works out of `~/.basecoin`. To change this, set the
`BCHOME` environment variable:

```
export BCHOME=~/.my_basecoin_data
basecoin init $ME
basecoin start
```

or

```
BCHOME=~/.my_basecoin_data basecoin init $ME
BCHOME=~/.my_basecoin_data basecoin start
```

# ABCI Server

So far we have run Basecoin and Tendermint in a single process.  However, since
we use ABCI, we can actually run them in different processes.  First,
initialize them:

```
basecoin init $ME
```

This will create a single `genesis.json` file in `~/.basecoin` with the
information for both Basecoin and Tendermint.

Now, In one window, run

```
basecoin start --without-tendermint
```

and in another,

```
TMROOT=~/.basecoin tendermint node
```

You should see Tendermint start making blocks!

Alternatively, you could ignore the Tendermint details in
`~/.basecoin/genesis.json` and use a separate directory by running:

```
tendermint init
tendermint node
```

For more details on using `tendermint`, see [the guide](https://tendermint.com/docs/guides/using-tendermint).

# Keys and Genesis

In previous tutorials we used `basecoin init` to initialize `~/.basecoin` with
the default configuration.  This command creates files both for Tendermint and
for Basecoin, and a single `genesis.json` file for both of them.  For more
information on these files, see the [guide to using
Tendermint](https://tendermint.com/docs/guides/using-tendermint).

Now let's make our own custom Basecoin data.

First, create a new directory:

```
mkdir example-data
```

We can tell `basecoin` to use this directory by exporting the `BCHOME`
environment variable:

```
export BCHOME=$(pwd)/example-data
```

If you're going to be using multiple terminal windows, make sure to add this
variable to your shell startup scripts (eg. `~/.bashrc`).

Now, let's create a new key:

```
basecli keys new foobar
```

The key's info can be retrieved with

```
basecli keys get foobar -o=json
```

You should get output which looks similar to the following:

```json
{
  "name": "foobar",
  "address": "404C5003A703C7DA888C96A2E901FCE65A6869D9",
  "pubkey": {
    "type": "ed25519",
    "data": "8786B7812AB3B27892D8E14505EEFDBB609699E936F6A4871B1983F210736EEA"
  }
}
```

Yours will look different - each key is randomly derived. Now we can make a
`genesis.json` file and add an account with our public key:

```json
{
  "app_hash": "",
  "chain_id": "example-chain",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "amount": 10,
      "name": "",
      "pub_key": {
        "type": "ed25519",
        "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      }
    }
  ],
  "app_options": {
    "accounts": [
      {
        "pub_key": {
          "type": "ed25519",
          "data": "8786B7812AB3B27892D8E14505EEFDBB609699E936F6A4871B1983F210736EEA"
        },
        "coins": [
          {
            "denom": "gold",
            "amount": 1000000000
          }
        ]
      }
    ]
  }
}
```

Here we've granted ourselves `1000000000` units of the `gold` token.  Note that
we've also set the `chain-id` to be `example-chain`.  All transactions must
therefore include the `--chain-id example-chain` in order to make sure they are
valid for this chain.  Previously, we didn't need this flag because we were
using the default chain ID ("test_chain_id").  Now that we're using a custom
chain, we need to specify the chain explicitly on the command line.

Note we have also left out the details of the Tendermint genesis. These are
documented in the [Tendermint
guide](https://tendermint.com/docs/guides/using-tendermint).


# Reset

You can reset all blockchain data by running:

```
basecoin unsafe_reset_all
```

Similarly, you can reset client data by running:
 
```
basecli reset_all
```

# Genesis

Any required plugin initialization should be constructed using `SetOption` on
genesis.  When starting a new chain for the first time, `SetOption` will be
called for each item the genesis file.  Within genesis.json file entries are
made in the format: `"<plugin>/<key>", "<value>"`, where `<plugin>` is the
plugin name, and `<key>` and `<value>` are the strings passed into the plugin
SetOption function.  This function is intended to be used to set plugin
specific information such as the plugin state.

