IBC
===

TODO: update in light of latest SDK (this document is currently out of date)

One of the most exciting elements of the Cosmos Network is the
InterBlockchain Communication (IBC) protocol, which enables
interoperability across different blockchains. We implemented IBC as a
basecoin plugin, and we'll show you how to use it to send tokens across
blockchains!

Please note: this tutorial assumes familiarity with the Cosmos SDK.

The IBC plugin defines a new set of transactions as subtypes of the
``AppTx``. The plugin's functionality is accessed by setting the
``AppTx.Name`` field to ``"IBC"``, and setting the ``Data`` field to the
serialized IBC transaction type.

We'll demonstrate exactly how this works below.

Inter BlockChain Communication
------------------------------

Let's review the IBC protocol. The purpose of IBC is to enable one
blockchain to function as a light-client of another. Since we are using
a classical Byzantine Fault Tolerant consensus algorithm, light-client
verification is cheap and easy: all we have to do is check validator
signatures on the latest block, and verify a Merkle proof of the state.

In Tendermint, validators agree on a block before processing it. This
means that the signatures and state root for that block aren't included
until the next block. Thus, each block contains a field called
``LastCommit``, which contains the votes responsible for committing the
previous block, and a field in the block header called ``AppHash``,
which refers to the Merkle root hash of the application after processing
the transactions from the previous block. So, if we want to verify the
``AppHash`` from height H, we need the signatures from ``LastCommit`` at
height H+1. (And remember that this ``AppHash`` only contains the
results from all transactions up to and including block H-1)

Unlike Proof-of-Work, the light-client protocol does not need to
download and check all the headers in the blockchain - the client can
always jump straight to the latest header available, so long as the
validator set has not changed much. If the validator set is changing,
the client needs to track these changes, which requires downloading
headers for each block in which there is a significant change. Here, we
will assume the validator set is constant, and postpone handling
validator set changes for another time.

Now we can describe exactly how IBC works. Suppose we have two
blockchains, ``chain1`` and ``chain2``, and we want to send some data
from ``chain1`` to ``chain2``. We need to do the following: 1. Register
the details (ie. chain ID and genesis configuration) of ``chain1`` on
``chain2`` 2. Within ``chain1``, broadcast a transaction that creates an
outgoing IBC packet destined for ``chain2`` 3. Broadcast a transaction
to ``chain2`` informing it of the latest state (ie. header and commit
signatures) of ``chain1`` 4. Post the outgoing packet from ``chain1`` to
``chain2``, including the proof that it was indeed committed on
``chain1``. Note ``chain2`` can only verify this proof because it has a
recent header and commit.

Each of these steps involves a separate IBC transaction type. Let's take
them up in turn.

IBCRegisterChainTx
~~~~~~~~~~~~~~~~~~

The ``IBCRegisterChainTx`` is used to register one chain on another. It
contains the chain ID and genesis configuration of the chain to
register:

::

    type IBCRegisterChainTx struct { BlockchainGenesis }

    type BlockchainGenesis struct { ChainID string Genesis string }

This transaction should only be sent once for a given chain ID, and
successive sends will return an error.

IBCUpdateChainTx
~~~~~~~~~~~~~~~~

The ``IBCUpdateChainTx`` is used to update the state of one chain on
another. It contains the header and commit signatures for some block in
the chain:

::

    type IBCUpdateChainTx struct {
      Header tm.Header
      Commit tm.Commit
    }

In the future, it needs to be updated to include changes to the
validator set as well. Anyone can relay an ``IBCUpdateChainTx``, and
they only need to do so as frequently as packets are being sent or the
validator set is changing.

IBCPacketCreateTx
~~~~~~~~~~~~~~~~~

The ``IBCPacketCreateTx`` is used to create an outgoing packet on one
chain. The packet itself contains the source and destination chain IDs,
a sequence number (i.e. an integer that increments with every message
sent between this pair of chains), a packet type (e.g. coin, data,
etc.), and a payload.

::

    type IBCPacketCreateTx struct {
      Packet
    }

    type Packet struct {
      SrcChainID string
      DstChainID string
      Sequence   uint64
      Type string
      Payload    []byte
    }

We have yet to define the format for the payload, so, for now, it's just
arbitrary bytes.

One way to think about this is that ``chain2`` has an account on
``chain1``. With a ``IBCPacketCreateTx`` on ``chain1``, we send funds to
that account. Then we can prove to ``chain2`` that there are funds
locked up for it in it's account on ``chain1``. Those funds can only be
unlocked with corresponding IBC messages back from ``chain2`` to
``chain1`` sending the locked funds to another account on ``chain1``.

IBCPacketPostTx
~~~~~~~~~~~~~~~

The ``IBCPacketPostTx`` is used to post an outgoing packet from one
chain to another. It contains the packet and a proof that the packet was
committed into the state of the sending chain:

::

    type IBCPacketPostTx struct {
      FromChainID     string // The immediate source of the packet, not always Packet.SrcChainID
      FromChainHeight uint64 // The block height in which Packet was committed, to check Proof Packet
      Proof *merkle.IAVLProof
    }

The proof is a Merkle proof in an IAVL tree, our implementation of a
balanced, Merklized binary search tree. It contains a list of nodes in
the tree, which can be hashed together to get the Merkle root hash. This
hash must match the ``AppHash`` contained in the header at
``FromChainHeight + 1``

-  note the ``+ 1`` is necessary since ``FromChainHeight`` is the height
   in which the packet was committed, and the resulting state root is
   not included until the next block.

IBC State
~~~~~~~~~

Now that we've seen all the transaction types, let's talk about the
state. Each chain stores some IBC state in its Merkle tree. For each
chain being tracked by our chain, we store:

-  Genesis configuration
-  Latest state
-  Headers for recent heights

We also store all incoming (ingress) and outgoing (egress) packets.

The state of a chain is updated every time an ``IBCUpdateChainTx`` is
committed. New packets are added to the egress state upon
``IBCPacketCreateTx``. New packets are added to the ingress state upon
``IBCPacketPostTx``, assuming the proof checks out.

Merkle Queries
--------------

The Basecoin application uses a single Merkle tree that is shared across
all its state, including the built-in accounts state and all plugin
state. For this reason, it's important to use explicit key names and/or
hashes to ensure there are no collisions.

We can query the Merkle tree using the ABCI Query method. If we pass in
the correct key, it will return the corresponding value, as well as a
proof that the key and value are contained in the Merkle tree.

The results of a query can thus be used as proof in an
``IBCPacketPostTx``.

Relay
-----

While we need all these packet types internally to keep track of all the
proofs on both chains in a secure manner, for the normal work-flow, we
can run a relay node that handles the cross-chain interaction.

In this case, there are only two steps. First ``basecoin relay init``,
which must be run once to register each chain with the other one, and
make sure they are ready to send and recieve. And then
``basecoin relay start``, which is a long-running process polling the
queue on each side, and relaying all new message to the other block.

This requires that the relay has access to accounts with some funds on
both chains to pay for all the ibc packets it will be forwarding.

Try it out
----------

Now that we have all the background knowledge, let's actually walk
through the tutorial.

Make sure you have installed `basecoin and
basecli </docs/guide/install.md>`__.

Basecoin is a framework for creating new cryptocurrency applications. It
comes with an ``IBC`` plugin enabled by default.

You will also want to install the
`jq <https://stedolan.github.io/jq/>`__ for handling JSON at the command
line.

If you have any trouble with this, you can also look at the `test
scripts </tests/cli/ibc.sh>`__ or just run ``make test_cli`` in basecoin
repo. Otherwise, open up 5 (yes 5!) terminal tabs....

Preliminaries
~~~~~~~~~~~~~

::

    # first, clean up any old garbage for a fresh slate...
    rm -rf ~/.ibcdemo/

Let's start by setting up some environment variables and aliases:

::

    export BCHOME1_CLIENT=~/.ibcdemo/chain1/client
    export BCHOME1_SERVER=~/.ibcdemo/chain1/server
    export BCHOME2_CLIENT=~/.ibcdemo/chain2/client
    export BCHOME2_SERVER=~/.ibcdemo/chain2/server
    alias basecli1="basecli --home $BCHOME1_CLIENT"
    alias basecli2="basecli --home $BCHOME2_CLIENT"
    alias basecoin1="basecoin --home $BCHOME1_SERVER"
    alias basecoin2="basecoin --home $BCHOME2_SERVER"

This will give us some new commands to use instead of raw ``basecli``
and ``basecoin`` to ensure we're using the right configuration for the
chain we want to talk to.

We also want to set some chain IDs:

::

    export CHAINID1="test-chain-1"
    export CHAINID2="test-chain-2"

And since we will run two different chains on one machine, we need to
maintain different sets of ports:

::

    export PORT_PREFIX1=1234
    export PORT_PREFIX2=2345
    export RPC_PORT1=${PORT_PREFIX1}7
    export RPC_PORT2=${PORT_PREFIX2}7

Setup Chain 1
~~~~~~~~~~~~~

Now, let's create some keys that we can use for accounts on
test-chain-1:

::

    basecli1 keys new money
    basecli1 keys new gotnone
    export MONEY=$(basecli1 keys get money | awk '{print $2}')
    export GOTNONE=$(basecli1 keys get gotnone | awk '{print $2}')

and create an initial configuration giving lots of coins to the $MONEY
key:

::

    basecoin1 init --chain-id $CHAINID1 $MONEY

Now start basecoin:

::

    sed -ie "s/4665/$PORT_PREFIX1/" $BCHOME1_SERVER/config.toml

    basecoin1 start &> basecoin1.log &

Note the ``sed`` command to replace the ports in the config file. You
can follow the logs with ``tail -f basecoin1.log``

Now we can attach the client to the chain and verify the state. The
first account should have money, the second none:

::

    basecli1 init --node=tcp://localhost:${RPC_PORT1} --genesis=${BCHOME1_SERVER}/genesis.json
    basecli1 query account $MONEY
    basecli1 query account $GOTNONE

Setup Chain 2
~~~~~~~~~~~~~

This is the same as above, except with ``basecli2``, ``basecoin2``, and
``$CHAINID2``. We will also need to change the ports, since we're
running another chain on the same local machine.

Let's create new keys for test-chain-2:

::

    basecli2 keys new moremoney
    basecli2 keys new broke
    MOREMONEY=$(basecli2 keys get moremoney | awk '{print $2}')
    BROKE=$(basecli2 keys get broke | awk '{print $2}')

And prepare the genesis block, and start the server:

::

    basecoin2 init --chain-id $CHAINID2 $(basecli2 keys get moremoney | awk '{print $2}')

    sed -ie "s/4665/$PORT_PREFIX2/" $BCHOME2_SERVER/config.toml

    basecoin2 start &> basecoin2.log &

Now attach the client to the chain and verify the state. The first
account should have money, the second none:

::

    basecli2 init --node=tcp://localhost:${RPC_PORT2} --genesis=${BCHOME2_SERVER}/genesis.json
    basecli2 query account $MOREMONEY
    basecli2 query account $BROKE

Connect these chains
~~~~~~~~~~~~~~~~~~~~

OK! So we have two chains running on your local machine, with different
keys on each. Let's hook them up together by starting a relay process to
forward messages from one chain to the other.

The relay account needs some money in it to pay for the ibc messages, so
for now, we have to transfer some cash from the rich accounts before we
start the actual relay.

::

    # note that this key.json file is a hardcoded demo for all chains, this will
    # be updated in a future release
    RELAY_KEY=$BCHOME1_SERVER/key.json
    RELAY_ADDR=$(cat $RELAY_KEY | jq .address | tr -d \")

    basecli1 tx send --amount=100000mycoin --sequence=1 --to=$RELAY_ADDR--name=money
    basecli1 query account $RELAY_ADDR

    basecli2 tx send --amount=100000mycoin --sequence=1 --to=$RELAY_ADDR --name=moremoney
    basecli2 query account $RELAY_ADDR

Now we can start the relay process.

::

    basecoin relay init --chain1-id=$CHAINID1 --chain2-id=$CHAINID2 \
      --chain1-addr=tcp://localhost:${RPC_PORT1} --chain2-addr=tcp://localhost:${RPC_PORT2} \
      --genesis1=${BCHOME1_SERVER}/genesis.json --genesis2=${BCHOME2_SERVER}/genesis.json \
      --from=$RELAY_KEY

    basecoin relay start --chain1-id=$CHAINID1 --chain2-id=$CHAINID2 \
      --chain1-addr=tcp://localhost:${RPC_PORT1} --chain2-addr=tcp://localhost:${RPC_PORT2} \
      --from=$RELAY_KEY &> relay.log &

This should start up the relay, and assuming no error messages came out,
the two chains are now fully connected over IBC. Let's use this to send
our first tx accross the chains...

Sending cross-chain payments
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The hard part is over, we set up two blockchains, a few private keys,
and a secure relay between them. Now we can enjoy the fruits of our
labor...

::

    # Here's an empty account on test-chain-2
    basecli2 query account $BROKE

::

    # Let's send some funds from test-chain-1
    basecli1 tx send --amount=12345mycoin --sequence=2 --to=test-chain-2/$BROKE --name=money

::

    # give it time to arrive...
    sleep 2
    # now you should see 12345 coins!
    basecli2 query account $BROKE

You're no longer broke! Cool, huh? Now have fun exploring and sending
coins across the chains. And making more accounts as you want to.

Conclusion
----------

In this tutorial we explained how IBC works, and demonstrated how to use
it to communicate between two chains. We did the simplest communciation
possible: a one way transfer of data from chain1 to chain2. The most
important part was that we updated chain2 with the latest state (i.e.
header and commit) of chain1, and then were able to post a proof to
chain2 that a packet was committed to the outgoing state of chain1.

In a future tutorial, we will demonstrate how to use IBC to actually
transfer tokens between two blockchains, but we'll do it with real
testnets deployed across multiple nodes on the network. Stay tuned!
