Local Testnet
=============

This tutorial demonstrates the basics of setting up a gaia
testnet locally.

If you haven't already made a key, make one now:

::

    gaia client keys new alice

otherwise, use an existing key.

Initialize The Chain
--------------------

Now initialize a gaia chain, using ``alice``'s address:

::

    gaia node init 5D93A6059B6592833CBC8FA3DA90EE0382198985 --home=$HOME/.gaia1 --chain-id=gaia-test

This will create all the files necessary to run a single node chain in
``$HOME/.gaia1``: a ``priv_validator.json`` file with the validators
private key, and a ``genesis.json`` file with the list of validators and
accounts.

We'll add a second node on our local machine by initiating a node in a
new directory, with the same address, and copying in the genesis:

::

    gaia node init 5D93A6059B6592833CBC8FA3DA90EE0382198985 --home=$HOME/.gaia2 --chain-id=gaia-test
    cp $HOME/.gaia1/genesis.json $HOME/.gaia2/genesis.json

We also need to modify ``$HOME/.gaia2/config.toml`` to set new seeds
and ports. It should look like:

::

    proxy_app = "tcp://127.0.0.1:26668"
    moniker = "anonymous"
    fast_sync = true
    db_backend = "leveldb"
    log_level = "state:info,*:error"

    [rpc]
    laddr = "tcp://0.0.0.0:26667"

    [p2p]
    laddr = "tcp://0.0.0.0:26666"
    seeds = "0.0.0.0:26656"

Start Nodes
-----------

Now that we've initialized the chains, we can start both nodes:

NOTE: each command below must be started in seperate terminal windows. Alternatively, to run this testnet across multiple machines, you'd replace the ``seeds = "0.0.0.0"`` in ``~/.gaia2.config.toml`` with the IP of the first node, and could skip the modifications we made to the config file above because port conflicts would be avoided.

::

    gaia node start --home=$HOME/.gaia1
    gaia node start --home=$HOME/.gaia2

Now we can initialize a client for the first node, and look up our
account:

::

    gaia client init --chain-id=gaia-test --node=tcp://localhost:26657
    gaia client query account 5D93A6059B6592833CBC8FA3DA90EE0382198985 

To see what tendermint considers the validator set is, use:

::

    curl localhost:26657/validators

and compare the information in this file: ``~/.gaia1/priv_validator.json``. The ``address`` and ``pub_key`` fields should match.

To add a second validator on your testnet, you'll need to bond some tokens be declaring candidacy.
