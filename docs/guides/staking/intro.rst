Using The Staking Module
========================

This project is a demonstration of the Cosmos Hub staking functionality; it is
designed to get validator acquianted with staking concepts and procedures.

Potential validators will be declaring their candidacy, after which users can
delegate and, if they so wish, unbond. This can be practiced using a local or
public testnet.

This example covers initial setup of a two-node testnet between a server in the cloud and a local machine. Begin this tutorial from a cloud machine that you've ``ssh``'d into.

Install
-------

The ``gaiad`` and ``gaiacli`` binaries:

::

    go get github.com/cosmos/cosmos-sdk
    cd $GOPATH/src/github.com/cosmos/cosmos-sdk
    make get_vendor_deps
    make install

Let's jump right into it. First, we initialize some default files:

::

    gaiad init

which will output:

::

    I[03-30|11:20:13.365] Found private validator                      module=main path=/root/.gaiad/config/priv_validator.json
    I[03-30|11:20:13.365] Found genesis file                           module=main path=/root/.gaiad/config/genesis.json
    Secret phrase to access coins:
    citizen hungry tennis noise park hire glory exercise link glow dolphin labor design grit apple abandon

This tell us we have a ``priv_validator.json`` and ``genesis.json`` in the ``~/.gaiad/config`` directory. A ``config.toml`` was also created in the same directory. It is a good idea to get familiar with those files. Write down the seed.

The next thing we'll need to is add the key from ``priv_validator.json`` to the ``gaiacli`` key manager. For this we need a seed and a password:

::

    gaiacli keys add alice --recover

which will give you three prompts:

::

    Enter a passphrase for your key:
    Repeat the passphrase:
    Enter your recovery seed phrase:

create a password and copy in your seed phrase. The name and address of the key will be output:

::
    NAME:   ADDRESS:                                    PUBKEY:
    alice	67997DD03D527EB439B7193F2B813B05B219CC02	1624DE6220BB89786C1D597050438C728202436552C6226AB67453CDB2A4D2703402FB52B6

You can see all available keys with:

::

    gaiacli keys list

Setup Testnet
-------------

Next, we start the daemon (do this in another window):

::

    gaiad start

and you'll see blocks start streaming through.

For this example, we're doing the above on a cloud machine. The next steps should be done on your local machine or another server in the cloud, which will join the running testnet then bond/unbond.

Accounts
--------

We have:

- ``alice`` the initial validator (in the cloud)
- ``bob``  receives tokens from ``alice`` then declares candidacy (from local machine)
- ``charlie`` will bond and unbond to ``bob`` (from local machine)

Remember that ``alice`` was already created. On your second machine, install the binaries and create two new keys:

::

    gaiacli keys add bob
    gaiacli keys add charlie

both of which will prompt you for a password. Now we need to copy the ``genesis.json`` and ``config.toml`` from the first machine (with ``alice``) to the second machine. This is a good time to look at both these files.

The ``genesis.json`` should look something like:

::

    {
      "app_state": {
        "accounts": [
          {
            "address": "1D9B2356CAADF46D3EE3488E3CCE3028B4283DEE",
            "coins": [
              {
                "denom": "steak",
                "amount": 100000
              }
            ]
          }
        ],
        "stake": {
          "pool": {
            "total_supply": 0,
            "bonded_shares": {
              "num": 0,
              "denom": 1
            },
            "unbonded_shares": {
              "num": 0,
              "denom": 1
            },
            "bonded_pool": 0,
            "unbonded_pool": 0,
            "inflation_last_time": 0,
            "inflation": {
              "num": 7,
              "denom": 100
            }
          },
          "params": {
            "inflation_rate_change": {
              "num": 13,
              "denom": 100
            },
            "inflation_max": {
              "num": 20,
              "denom": 100
            },
            "inflation_min": {
              "num": 7,
              "denom": 100
            },
            "goal_bonded": {
              "num": 67,
              "denom": 100
            },
            "max_validators": 100,
            "bond_denom": "steak"
          }
        }
      },
      "validators": [
        {
          "pub_key": {
            "type": "AC26791624DE60",
            "value": "rgpc/ctVld6RpSfwN5yxGBF17R1PwMTdhQ9gKVUZp5g="
          },
          "power": 10,
          "name": ""
        }
      ],
      "app_hash": "",
      "genesis_time": "0001-01-01T00:00:00Z",
      "chain_id": "test-chain-Uv1EVU"
    }


To notice is that the ``accounts`` field has a an address and a whole bunch of "mycoin". This is ``alice``'s address (todo: dbl check). Under ``validators`` we see the ``pub_key.data`` field, which will match the same field in the ``priv_validator.json`` file.

The ``config.toml`` is long so let's focus on one field:

::

    # Comma separated list of seed nodes to connect to
    seeds = ""

On the ``alice`` cloud machine, we don't need to do anything here. Instead, we need its IP address. After copying this file (and the ``genesis.json`` to your local machine, you'll want to put the IP in the ``seeds =  "138.197.161.74"`` field, in this case, we have a made-up IP. For joining testnets with many nodes, you can add more comma-seperated IPs to the list.


Now that your files are all setup, it's time to join the network. On your local machine, run:

::

    gaiad start

and your new node will connect to the running validator (``alice``).

Sending Tokens
--------------

We'll have ``alice`` send some ``mycoin`` to ``bob``, who has now joined the network:

::

    gaiacli send --amount=1000mycoin --sequence=0 --name=alice --to=5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6 --chain-id=test-chain-Uv1EVU

where the ``--sequence`` flag is to be incremented for each transaction, the ``--name`` flag is the sender (alice), and the ``--to`` flag takes ``bob``'s address. You'll see something like:

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

TODO: check the above with current actual output.

Check out ``bob``'s account, which should now have 1000 mycoin:

::

    gaiacli account 5A35E4CC7B7DC0A5CB49CEA91763213A9AE92AD6

Adding a Second Validator
-------------------------

**This section is wrong/needs to be updated**

Next, let's add the second node as a validator.

First, we need the pub_key data:

** need to make bob a priv_Val above?

::

    cat $HOME/.gaia2/priv_validator.json

the first part will look like:

::

    {"address":"7B78527942C831E16907F10C3263D5ED933F7E99","pub_key":{"type":"ed25519","data":"96864CE7085B2E342B0F96F2E92B54B18C6CC700186238810D5AA7DFDAFDD3B2"},

and you want the ``pub_key`` ``data`` that starts with ``96864CE``.

Now ``bob`` can create a validator with that pubkey.

::

    gaiacli stake create-validator --amount=10mycoin --name=bob --address-validator=<address> --pub-key=<pubkey> --moniker=bobby

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


We should see ``bob``'s account balance decrease by 10 mycoin:

::

    gaiacli account 5D93A6059B6592833CBC8FA3DA90EE0382198985

To confirm for certain the new validator is active, ask the tendermint node:

::

    curl localhost:26657/validators

If you now kill either node, blocks will stop streaming in, because
there aren't enough validators online. Turn it back on and they will
start streaming again.

Now that ``bob`` has declared candidacy, which essentially bonded 10 mycoin and made him a validator, we're going to get ``charlie`` to delegate some coins to ``bob``.

Delegating
----------

First let's have ``alice`` send some coins to ``charlie``:

::

    gaiacli send --amount=1000mycoin --sequence=2 --name=alice --to=48F74F48281C89E5E4BE9092F735EA519768E8EF

Then ``charlie`` will delegate some mycoin to ``bob``:

::

    gaiacli stake delegate --amount=10mycoin --address-delegator=<charlie's address> --address-validator=<bob's address> --name=charlie

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

And that's it. You can query ``charlie``'s account to see the decrease in mycoin.

To get more information about the candidate, try:

::

    gaiacli stake validator <address>

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

    gaiacli stake delegation --address-delegator=<address> --address-validator=<address>

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


where the ``--address-delegator`` is ``charlie``'s address and the ``--address-validator`` is ``bob``'s address.


Unbonding
---------

Finally, to relinquish your voting power, unbond some coins. You should see
your VotingPower reduce and your account balance increase.

::

    gaiacli stake unbond --amount=5mycoin --name=charlie --address-delegator=<address> --address-validator=<address>
    gaiacli account 48F74F48281C89E5E4BE9092F735EA519768E8EF

See the bond decrease with ``gaiacli stake delegation`` like above.
