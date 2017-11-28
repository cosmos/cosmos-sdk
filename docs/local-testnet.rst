Local Testnet Example
=====================

This tutorial demonstrates the basics of setting up a gaia
testnet locally

First, generate a new key with a name, and save the address:

::

    gaia client keys new alice
    gaia client keys list

This will output:

::

    Enter a passphrase:
    Repeat the passphrase:
    alice	    E9E103F788AADD9C0842231E496B2139C118FA60
    **Important** write this seed phrase in a safe place.
    It is the only way to recover your account if you ever forget your password.

    inject position weather divorce shine immense middle affair piece oval silver type until spike educate abandon

which has your address and will be re-used throughout this tutorial.
We recommend doing something like ``MYADDR=<your address>``. Writing 
down the recovery phrase is crucial for production keys, however,
for the tutorial you can skip this step.

Now initialize a gaia chain:

::

    gaia node init E9E103F788AADD9C0842231E496B2139C118FA60 --home=$HOME/.gaia1 --chain-id=gaia-test

This will create all the files necessary to run a single node chain in
``$HOME/.gaia1``: a ``priv_validator.json`` file with the validators
private key, and a ``genesis.json`` file with the list of validators and
accounts. In this case, we have one random validator, and ``$MYADDR`` is
an independent account that has a bunch of coins.

We can add a second node on our local machine by initiating a node in a
new directory, and copying in the genesis:

::

    gaia node init E9E103F788AADD9C0842231E496B2139C118FA60 --home=$HOME/.gaia2 --chain-id=gaia-test
    cp $HOME/.gaia1/genesis.json $HOME/.gaia2/genesis.json

We need to also modify ``$HOME/.gaia2/config.toml`` to set new seeds
and ports. It should look like:

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
    seeds = "0.0.0.0:46656"

Great, now that we've initialized the chains, we can start both nodes:

NOTE: each command below must be started in seperate terminal windows.

::

    gaia node start --home=$HOME/.gaia1
    gaia node start --home=$HOME/.gaia2

Now we can initialize a client for the first node, and look up our
account:

::

    gaia client init --chain-id=gaia-test --node=tcp://localhost:46657
    gaia client query account E9E103F788AADD9C0842231E496B2139C118FA60

Nice. We can also lookup the validator set:

::

    gaia client query validators

Notice it's empty! This is because the initial validators are special -
the app doesn't know about them, so they can't be removed. To see what
tendermint itself thinks the validator set is, use:

::

    curl localhost:46657/validators

Ok, let's add the second node as a validator. First, we need the pubkey
data:

::

    cat $HOME/.gaia2/priv_validator.json 

If you have a json parser like ``jq``, you can get just the pubkey:

::

    cat $HOME/.gaia2/priv_validator.json | jq .pub_key.data

Now we can bond some coins to that pubkey:

::

    gaia client tx bond --amount=10fermion --name=alice --pubkey=<validator pubkey>

We should see our account balance decrement, and the pubkey get added to
the app's list of bonds:

::

    gaia client query account E9E103F788AADD9C0842231E496B2139C118FA60
    gaia client query validators

To confirm for certain the new validator is active, check tendermint:

::

    curl localhost:46657/validators

If you now kill your second node, blocks will stop streaming in, because
there aren't enough validators online. Turn it back on and they will
start streaming again.

Finally, to relinquish all your power, unbond some coins. You should see
your VotingPower reduce and your account balance increase.

::

    gaia client tx unbond --amount=10fermion --name=alice
    gaia client query validators
    gaia client query account E9E103F788AADD9C0842231E496B2139C118FA60

Once you unbond enough, you will no longer be needed to make new blocks.

That concludes an overview of the ``gaia`` tooling for local testing.
