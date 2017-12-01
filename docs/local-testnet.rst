Local Testnet Example
=====================

This tutorial demonstrates the basics of setting up a gaia
testnet locally

Generate Keys
-------------

First, generate a new key with a name:

::

    gaia client keys new alice

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
for this tutorial you can skip this step.

Because ``alice`` will be delclaring candidacy to be a validator, we need another key, ``bob`` who will delegate his tokens to ``alice``.

::

    gaia client keys new bob

Now we can see the keys we've created:

::

    gaia client keys list

which show's something like:

::

    All keys:
    alice               E9E103F788AADD9C0842231E496B2139C118FA60
    bob                 7E00832E8CC9D15E3AE6EEBAE09C3CB83AA04361

Try add the ``--output json`` flag to the above command to get more information.
We've got our keys made, so let's move on to the next step.

Initialize the chain
--------------------

Now initialize a gaia chain, using ``alice``'s address:

::

    gaia node init E9E103F788AADD9C0842231E496B2139C118FA60 --home=$HOME/.gaia1 --chain-id=gaia-test

This will create all the files necessary to run a single node chain in
``$HOME/.gaia1``: a ``priv_validator.json`` file with the validators
private key, and a ``genesis.json`` file with the list of validators and
accounts.

We'll add a second node on our local machine by initiating a node in a
new directory, and copying in the genesis:

::

    gaia node init E9E103F788AADD9C0842231E496B2139C118FA60 --home=$HOME/.gaia2 --chain-id=gaia-test
    cp $HOME/.gaia1/genesis.json $HOME/.gaia2/genesis.json

We also need to modify ``$HOME/.gaia2/config.toml`` to set new seeds
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

NOTE: each command below must be started in seperate terminal windows. Alternatively, to run this testnet across multiple machines, you'd replace the ``seeds = "0.0.0.0"`` in ``~/.gaia2.config.toml``, and could skip the modifications we made to the config file above because port conflicts would be avoided.

::

    gaia node start --home=$HOME/.gaia1
    gaia node start --home=$HOME/.gaia2

Now we can initialize a client for the first node, and look up our
account:

::

    gaia client init --chain-id=gaia-test --node=tcp://localhost:46657
    gaia client query account E9E103F788AADD9C0842231E496B2139C118FA60

To see what tendermint considers the validator set is, use:

::

    curl localhost:46657/validators

and compare the information in this file: ``~/.gaia1/priv_validator.json``. The ``address`` and ``pub_key`` fields should match.

Add Second Validator
--------------------

Next, let's add the second node as a validator.

First, we need the pub_key data:

::

    cat $HOME/.gaia2/priv_validator.json 

the first part will look like:

::

    {"address":"7B78527942C831E16907F10C3263D5ED933F7E99","pub_key":{"type":"ed25519","data":"96864CE7085B2E342B0F96F2E92B54B18C6CC700186238810D5AA7DFDAFDD3B2"},

and you want the ``pub_key`` ``data`` that starts with ``96864CE``.

Now we can declare candidacy to that pubkey:

::

    gaia client tx declare-candidacy --amount=10fermion --name=alice --pubkey=<pub_key data>

We should see our account balance decrement:

::

    gaia client query account E9E103F788AADD9C0842231E496B2139C118FA60

To confirm for certain the new validator is active, ask the tendermint node:

::

    curl localhost:46657/validators

If you now kill either node, blocks will stop streaming in, because
there aren't enough validators online. Turn it back on and they will
start streaming again.

Now that ``alice`` has declared her candidacy, which essentially bonded 10 fermions and made her a validator, we're going to get ``bob`` to delegate some coins to ``alice``.

Delegate
--------



Unbond
------

Finally, to relinquish all your power, unbond some coins. You should see
your VotingPower reduce and your account balance increase.

::

    gaia client tx unbond --amount=10fermion --name=alice
    gaia client query account E9E103F788AADD9C0842231E496B2139C118FA60

Once you unbond enough, you will no longer be needed to make new blocks.

That concludes an overview of the ``gaia`` tooling for local testing.
