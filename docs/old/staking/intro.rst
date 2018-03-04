Using Gaia
==========

This project is a demonstration of the Cosmos Hub with staking functionality; it is
designed to get validator acquianted with staking concepts and procedure.

Potential validators will be declaring their candidacy, after which users can
delegate and, if they so wish, unbond. This can be practiced using a local or
public testnet.

Install
-------

The ``gaia`` tooling is an extension of the Cosmos-SDK; to install:

::

    go get github.com/cosmos/gaia
    cd $GOPATH/src/github.com/cosmos/gaia
    make get_vendor_deps
    make install

It has three primary commands:

::

    Available Commands:
      node        The Cosmos Network delegation-game blockchain test
      rest-server REST client for gaia commands
      client      Gaia light client
                        
      version     Show version info
      help        Help about any command

and a handful of flags that are highlighted only as necessary.

The ``gaia node`` command is a proxt for running a tendermint node. You'll be using
this command to either initialize a new node, or - using existing files - joining
the testnet. 

The ``gaia rest-server`` command is used by the `cosmos UI <https://github.com/cosmos/cosmos-ui>`__.

Lastly, the ``gaia client`` command is the workhorse of the staking module. It allows
for sending various transactions and other types of interaction with a running chain.
that you've setup or joined a testnet.

Generating Keys
---------------

Review the `key management tutorial <../key-management.html>`__ and create one key
if you'll be joining the public testnet, and three keys if you'll be trying out a local
testnet.

Setup Testnet
-------------

The first thing you'll want to do is either `create a local testnet <./local-testnet.html>`__ or
join a `public testnet <./public-testnet.html>`__. Either step is required before proceeding.

The rest of this tutorial will assume a local testnet with three participants: ``alice`` will be
the initial validator, ``bob`` will first receives tokens from ``alice`` then declare candidacy
as a validator, and ``charlie`` will bond then unbond to ``bob``. If you're joining the public
testnet, the token amounts will need to be adjusted.

Sending Tokens
--------------

We'll have ``alice`` who is currently quite rich, send some ``fermions`` to ``bob``:

::

    gaia client tx send --amount=1000fermion --sequence=1 --name=alice --to=5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6

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

Adding a Second Validator
-------------------------

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

Delegating
----------

First let's have ``alice`` send some coins to ``charlie``:

::

    gaia client tx send --amount=1000fermion --sequence=2 --name=alice --to=48F74F48281C89E5E4BE9092F735EA519768E8EF

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

It's also possible the query the delegator's bond like so:

::

    gaia client query delegator-bond --delegator-address 48F74F48281C89E5E4BE9092F735EA519768E8EF --pubkey 52D6FCD8C92A97F7CCB01205ADF310A18411EA8FDCC10E65BF2FCDB05AD1689B

with an output similar to:

::

    {
      "height": 325782,
      "data": {
        "PubKey": {
          "type": "ed25519",
          "data": "52D6FCD8C92A97F7CCB01205ADF310A18411EA8FDCC10E65BF2FCDB05AD1689B"
        },
        "Shares": 20
      }
    }
 

where the ``--delegator-address`` is ``charlie``'s address and the ``-pubkey`` is the same as we've been using.


Unbonding
---------

Finally, to relinquish your voting power, unbond some coins. You should see
your VotingPower reduce and your account balance increase.

::

    gaia client tx unbond --amount=5fermion --name=charlie --pubkey=<pub_key data>
    gaia client query account 48F74F48281C89E5E4BE9092F735EA519768E8EF

See the bond decrease with ``gaia client query delegator-bond`` like above.

That concludes an overview of the ``gaia`` tooling for local testing.
