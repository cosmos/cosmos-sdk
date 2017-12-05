Local Testnet Example
=====================

This tutorial demonstrates the basics of: setting up a gaia
testnet locally, sending transactions, declaring candidacy,
bonding, and unbonding. Various other commands of the ``gaia``
tooling are also introduced.

Generate Keys
-------------

First, let's generate a key named ``alice``:

::

    gaia client keys new alice

This will output:

::

    Enter a passphrase:
    Repeat the passphrase:
    alice	    5D93A6059B6592833CBC8FA3DA90EE0382198985
    **Important** write this seed phrase in a safe place.
    It is the only way to recover your account if you ever forget your password.

    inject position weather divorce shine immense middle affair piece oval silver type until spike educate abandon

which has your address and will be re-used throughout this tutorial.
We recommend doing something like ``MYADDR=<your address>``. Writing 
down the recovery phrase is crucial for production keys, however,
for this tutorial you can skip this step.

Because ``alice`` will be the initial validator, we need another key, ``bob`` who will first receives tokens from ``alice``, then declare candidacy as a validator. We also need an account for ``charlie`` who will bond and unbond to ``bob``.

::

    gaia client keys new bob
    gaia client keys new charlie

Now we can see the keys we've created:

::

    gaia client keys list

which shows something like:

::

    All keys:
    alice           5D93A6059B6592833CBC8FA3DA90EE0382198985
    bob             5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6
    charlie         48F74F48281C89E5E4BE9092F735EA519768E8EF

Try adding the ``--output json`` flag to the above command to get more information.
The information for these keys is stored in ``~/.cosmos-gaia-cli``.
We've got our keys made, so let's move on to the next step.

Initialize the chain
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

NOTE: each command below must be started in seperate terminal windows. Alternatively, to run this testnet across multiple machines, you'd replace the ``seeds = "0.0.0.0"`` in ``~/.gaia2.config.toml`` with the IP of the first node, and could skip the modifications we made to the config file above because port conflicts would be avoided.

::

    gaia node start --home=$HOME/.gaia1
    gaia node start --home=$HOME/.gaia2

Now we can initialize a client for the first node, and look up our
account:

::

    gaia client init --chain-id=gaia-test --node=tcp://localhost:46657
    gaia client query account 5D93A6059B6592833CBC8FA3DA90EE0382198985 

To see what tendermint considers the validator set is, use:

::

    curl localhost:46657/validators

and compare the information in this file: ``~/.gaia1/priv_validator.json``. The ``address`` and ``pub_key`` fields should match.

Send Tokens
-----------

We'll have ``alice`` who is currently quite rich, send some ``fermions`` to ``bob``:

::

    gaia client tx send --amount=992fermion --sequence=1 --name=alice --to=5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6

where the ``--sequence`` flag is to be incremented for each transaction, the ``--name`` flag names the sender, and the ``--to`` flag takes ``bob``'s address. You'll see something like:

::

    Please enter passphrase for alice: 
    {
      "check_tx": {
        "gas": 30
      },
      "deliver_tx": {
        "tags": [
          {
            "key": "height",
            "value_type": 1,
            "value_int": 2963
          },
          {
            "key": "coin.sender",
            "value_string": "5D93A6059B6592833CBC8FA3DA90EE0382198985"
          },
          {
            "key": "coin.receiver",
            "value_string": "5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6"
          }
        ]
      },
      "hash": "423BD7EA3C4B36AF8AFCCA381C0771F8A698BA77",
      "height": 2963
    }

Check out ``bob``'s account, which should now have 992 fermions:

::

    gaia client query account 5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6

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

Now ``bob`` can declare candidacy to that pubkey:

::

    gaia client tx declare-candidacy --amount=10fermion --name=bob --pubkey=<pub_key data> --moniker=bobby

with an output like:

::

    Please enter passphrase for bob: 
    {
      "check_tx": {
        "gas": 30
      },
      "deliver_tx": {},
      "hash": "2A2A61FFBA1D7A59138E0068C82CC830E5103799",
      "height": 4075
    }


We should see ``bob``'s account balance decrease by 10 fermions:

::

    gaia client query account 5D93A6059B6592833CBC8FA3DA90EE0382198985 

To confirm for certain the new validator is active, ask the tendermint node:

::

    curl localhost:46657/validators

If you now kill either node, blocks will stop streaming in, because
there aren't enough validators online. Turn it back on and they will
start streaming again.

Now that ``bob`` has declared candidacy, which essentially bonded 10 fermions and made him a validator, we're going to get ``charlie`` to delegate some coins to ``bob``.

Delegate
--------

First let's have ``alice`` send some coins to ``charlie``:

::

    gaia client tx send --amount=999fermion --sequence=2 --name=alice --to=48F74F48281C89E5E4BE9092F735EA519768E8EF

Then ``charlie`` will delegate some fermions to ``bob``:

::

    gaia client tx delegate --amount=10fermion --name=charlie --pubkey=<pub_key data>

You'll see output like:

::

    Please enter passphrase for charlie: 
    {
      "check_tx": {
        "gas": 30
      },
      "deliver_tx": {},
      "hash": "C3443BA30FCCC1F6E3A3D6AAAEE885244F8554F0",
      "height": 51585
    }

And that's it. You can query ``charlie``'s account to see the decrease in fermions.

To get more information about the candidate, try:

::

    gaia client query candidate --pubkey=<pub_key data>

and you'll see output similar to:

::

    {
      "height": 51899,
      "data": {
        "pub_key": {
          "type": "ed25519",
          "data": "52D6FCD8C92A97F7CCB01205ADF310A18411EA8FDCC10E65BF2FCDB05AD1689B"
        },
        "owner": {
          "chain": "",
          "app": "sigs",
          "addr": "5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6"
        },
        "shares": 20,
        "voting_power": 20,
        "description": {
          "moniker": "bobby",
          "identity": "",
          "website": "",
          "details": ""
        }
      }
    }


Unbond
------

Finally, to relinquish all your power, unbond some coins. You should see
your VotingPower reduce and your account balance increase.

::

    gaia client tx unbond --amount=5fermion --name=charlie --pubkey=<pub_key data>
    gaia client query account 48F74F48281C89E5E4BE9092F735EA519768E8EF

That concludes an overview of the ``gaia`` tooling for local testing.
