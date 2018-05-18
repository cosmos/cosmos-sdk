Testnet Setup
=============

**Note:** This document is incomplete and may not be up-to-date with the state of the code.

See the `installation guide <../sdk/install.html>`__ for details on installation.

Here is a quick example to get you off your feet:

First, generate a couple of genesis transactions to be incorparated into the genesis file, this will create two keys with the password ``1234567890``

::

    gaiad init gen-tx --name=foo --home=$HOME/.gaiad1
    gaiad init gen-tx --name=bar --home=$HOME/.gaiad2
    gaiacli keys list

**Note:** If you've already run these tests you may need to overwrite keys using the ``--OWK`` flag
When you list the keys you should see two addresses, we'll need these later so take note.
Now let's actually create the genesis files for both nodes:

::

    cp -a ~/.gaiad2/config/gentx/. ~/.gaiad1/config/gentx/
    cp -a ~/.gaiad1/config/gentx/. ~/.gaiad2/config/gentx/
    gaiad init --gen-txs --home=$HOME/.gaiad1 --chain-id=test-chain
    gaiad init --gen-txs --home=$HOME/.gaiad2 --chain-id=test-chain

**Note:** If you've already run these tests you may need to overwrite genesis using the ``-o`` flag
What we just did is copy the genesis transactions between each of the nodes so there is a common genesis transaction set; then we created both genesis files independently from each home directory. Importantly both nodes have independently created their ``genesis.json`` and ``config.toml`` files, which should be identical between nodes.

Great, now that we've initialized the chains, we can start both nodes in the background:

::

    gaiad start --home=$HOME/.gaiad1  &> gaia1.log &
    NODE1_PID=$!
    gaia start --home=$HOME/.gaiad2  &> gaia2.log &
    NODE2_PID=$!

Note that we save the PID so we can later kill the processes. You can peak at your logs with ``tail gaia1.log``, or follow them for a bit with ``tail -f gaia1.log``.

Nice. We can also lookup the validator set:

::

    gaiacli advanced tendermint validator-set

Then, we try to transfer some ``steak`` to another account:

::

    gaiacli account <FOO-ADDR>
    gaiacli account <BAR-ADDR>
    gaiacli send --amount=10steak --to=<BAR-ADDR> --name=foo --chain-id=test-chain

**Note:** We need to be careful with the ``chain-id`` and ``sequence``

Check the balance & sequence with:

::

    gaiacli account <BAR-ADDR>

To confirm for certain the new validator is active, check tendermint:

::

    curl localhost:46657/validators

Finally, to relinquish all your power, unbond some coins. You should see your VotingPower reduce and your account balance increase.

::

    gaiacli stake unbond --chain-id=<chain-id> --name=test

That's it!

**Note:** TODO demonstrate edit-candidacy
**Note:** TODO demonstrate delegation
**Note:** TODO demonstrate unbond of delegation
**Note:** TODO demonstrate unbond candidate
