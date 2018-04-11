Testnet Setup
=============

Install
-------

See the `installation guide <../sdk/install.html>`__ for details on installation.


### Local-Test Example

Here is a quick example to get you off your feet:

First, generate a new key with a name, and save the address:

::

    MYNAME=<your name>
    basecli keys new $MYNAME
    basecli keys list
    MYADDR=<your newly generated address>


Now initialize a gaia chain:

::
    gaiad init --home=$HOME/.gaiad1

you should see seed phrase for genesis account in the output & config & data folder in the home directory.

In the config folder, there will be the following files: ``config.toml``, ``genesis.json``, ``node_key.json``, and ``priv_validator.json``.

The genesis file should look like this:

::

    {
      "genesis_time": "0001-01-01T00:00:00Z",
      "chain_id": "test-chain-0TRiTa",
      "validators": [
        {
          "pub_key": {
            "type": "AC26791624DE60",
            "value": "<value>"
          },
          "power": 10,
          "name": ""
        }
      ],
      "app_hash": "",
      "app_state": {
        "accounts": [
          {
            "address": "<ADDR>",
            "coins": [
              {
                "denom": "steak",
                "amount": 9007199254740992
              }
            ]
          }
        ]
      }
    }


**Note:** We need to change the denomination of token from default to ``steak`` in genesis file.

Then, recover the genesis account with ``basecli``:

::

    basecli keys add <name> --recover

By now, you have set up the first node. This is great!

We can add a second node on our local machine by initiating a node in a new directory, and copying in the ``genesis.json``:

::

    gaiad init --home=$HOME/.gaiad2

and replace the ``genesis.json`` and ``config.toml`` file:

::

    cp $HOME/.gaiad/config/genesis.json $HOME/.gaiad2/config
    cp $HOME/.gaiad/config/config.toml $HOME/.gaiad2/config

then, get the node id of first node:

::

    gaiad show_node_id --home=$HOME/.gaiad1

We need to also modify $HOME/.gaiad2/config.toml to set new seeds and ports. It should look like:

::

    proxy_app = "tcp://127.0.0.1:46668"
    moniker = "anonymous"
    fast_sync = true
    db_backend = "leveldb"
    log_level = "state:info,*:error"
    
    [rpc]
    laddr = "tcp://0.0.0.0:46667"
    
    [p2p]
    laddr = "tcp://0.0.0.0:46666"
    persistent_peers = "<node1-ID>@0.0.0.0:46656"


Great, now that we've initialized the chains, we can start both nodes in the background:

::

    gaiad start --home=$HOME/.gaiad1  &> gaia1.log &
    NODE1_PID=$!
    gaia start --home=$HOME/.gaiad2  &> gaia2.log &
    NODE2_PID=$!

Note that we save the PID so we can later kill the processes. You can peak at your logs with ``tail gaia1.log``, or follow them for a bit with ``tail -f gaia1.log``.

Nice. We can also lookup the validator set:

::

    basecli validatorset

There is only **one** validator now. Let's add another one!

First, we need to create a new account:

::

    basecli keys new <NAME>

Check that we now have two accounts:

::
    basecli keys list 

Then, we try to transfer some ``steak`` to another account:

::

    basecli send --amount=1000steak --to=$MYADDR2 --name=$NAME --chain-id=<CHAIN-ID> --node=tcp://localhost:46657 --sequence=0

**Note** We need to be careful with the ``chain-id`` and ``sequence``

Check the balance & sequence with:

::

    basecli account $MYADDR

We can see the balance of ``$MYADDR2`` is 1000 now. 

Finally, let's bond the validator in ``$HOME/gaiad2``. Get the pubkey first:

::

    cat $HOME/.gaiad2/config/priv_validator.json | jq .pub_key.value

Go to [this website](http://tomeko.net/online_tools/base64.php?lang=en) to change pubkey from base64 to Hex. 

Ok, now we can bond some coins to that pubkey:

::

    basecli bond --stake=1steak --validator=<validator-pubkey-hex> --sequence=0 --chain-id=<chain-id> --name=test

Nice. We can see there are now two validators:

::

    basecli validatorset

Check the balance of ``$MYADDR2`` to see the difference: it has 1 less ``steak``!

::
    basecli account $MYADDR2

To confirm for certain the new validator is active, check tendermint:

::

    curl localhost:46657/validators

Finally, to relinquish all your power, unbond some coins. You should see your VotingPower reduce and your account balance increase.

::

    basecli unbond  --sequence=# --chain-id=<chain-id> --name=test

That's it!
