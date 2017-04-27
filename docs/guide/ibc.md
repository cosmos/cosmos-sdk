# InterBlockchain Communication with Basecoin

One of the most exciting elements of the Cosmos Network is the InterBlockchain Communication (IBC) protocol,
which enables interoperability across different blockchains.
The simplest example of using the IBC protocol is to send a data packet from one blockchain to another.

We implemented IBC as a basecoin plugin.
and here we'll show you how to use the Basecoin IBC-plugin to send a packet of data across blockchains!

Please note, this tutorial assumes you are familiar with [Basecoin plugins](/docs/guide/basecoin-plugins.md),
but we'll explain how IBC works. You may also want to see [our repository of example plugins](https://github.com/tendermint/basecoin-examples).

The IBC plugin defines a new set of transactions as subtypes of the `AppTx`.
The plugin's functionality is accessed by setting the `AppTx.Name` field to `"IBC"`, 
and setting the `Data` field to the serialized IBC transaction type.

We'll demonstrate exactly how this works below.

## IBC

Let's review the IBC protocol.
The purpose of IBC is to enable one blockchain to function as a light-client of another.
Since we are using a classical Byzantine Fault Tolerant consensus algorithm,
light-client verification is cheap and easy:
all we have to do is check validator signatures on the latest block,
and verify a Merkle proof of the state.

In Tendermint, validators agree on a block before processing it.  This means
that the signatures and state root for that block aren't included until the
next block.  Thus, each block contains a field called `LastCommit`, which
contains the votes responsible for committing the previous block, and a field
in the block header called `AppHash`, which refers to the Merkle root hash of
the application after processing the transactions from the previous block.  So,
if we want to verify the `AppHash` from height H, we need the signatures from `LastCommit` 
at height H+1. (And remember that this `AppHash` only contains the results from all 
transactions up to and including block H-1)

Unlike Proof-of-Work, the light-client protocol does not need to download and
check all the headers in the blockchain - the client can always jump straight
to the latest header available, so long as the validator set has not changed
much.  If the validator set is changing, the client needs to track these
changes, which requires downloading headers for each block in which there is a
significant change. Here, we will assume the validator set is constant, and
postpone handling validator set changes for another time.

Now we can describe exactly how IBC works.
Suppose we have two blockchains, `chain1` and `chain2`, and we want to send some data from `chain1` to `chain2`.
We need to do the following:
 1. Register the details (ie. chain ID and genesis configuration) of `chain1` on `chain2`
 2. Within `chain1`, broadcast a transaction that creates an outgoing IBC packet destined for `chain2`
 3. Broadcast a transaction to `chain2` informing it of the latest state (ie. header and commit signatures) of `chain1`
 4. Post the outgoing packet from `chain1` to `chain2`, including the proof that
it was indeed committed on `chain1`. Note `chain2` can only verify this proof
because it has a recent header and commit.

Each of these steps involves a separate IBC transaction type. Let's take them up in turn.

### IBCRegisterChainTx

The `IBCRegisterChainTx` is used to register one chain on another.
It contains the chain ID and genesis configuration of the chain to register:

```golang
type IBCRegisterChainTx struct {
	BlockchainGenesis
}

type BlockchainGenesis struct {
	ChainID string
	Genesis string
}
```

This transaction should only be sent once for a given chain ID, and successive sends will return an error.


### IBCUpdateChainTx

The `IBCUpdateChainTx` is used to update the state of one chain on another.
It contains the header and commit signatures for some block in the chain:

```golang
type IBCUpdateChainTx struct {
	Header tm.Header
	Commit tm.Commit
}
```

In the future, it needs to be updated to include changes to the validator set as well.
Anyone can relay an `IBCUpdateChainTx`, and they only need to do so as frequently as packets are being sent or the validator set is changing.

### IBCPacketCreateTx

The `IBCPacketCreateTx` is used to create an outgoing packet on one chain.
The packet itself contains the source and destination chain IDs,
a sequence number (i.e. an integer that increments with every message sent between this pair of chains),
a packet type (e.g. coin, data, etc.),
and a payload.

```golang
type IBCPacketCreateTx struct {
	Packet
}

type Packet struct {
	SrcChainID string
	DstChainID string
	Sequence   uint64
	Type       string
	Payload    []byte
}
```

We have yet to define the format for the payload, so, for now, it's just arbitrary bytes.

One way to think about this is that `chain2` has an account on `chain1`.
With a `IBCPacketCreateTx` on `chain1`, we send funds to that account.
Then we can prove to `chain2` that there are funds locked up for it in it's
account on `chain1`.
Those funds can only be unlocked with corresponding IBC messages back from
`chain2` to `chain1` sending the locked funds to another account on
`chain1`.

### IBCPacketPostTx

The `IBCPacketPostTx` is used to post an outgoing packet from one chain to another.
It contains the packet and a proof that the packet was committed into the state of the sending chain:

```golang
type IBCPacketPostTx struct {
	FromChainID     string // The immediate source of the packet, not always Packet.SrcChainID
	FromChainHeight uint64 // The block height in which Packet was committed, to check Proof
	Packet
	Proof *merkle.IAVLProof
}
```

The proof is a Merkle proof in an IAVL tree, our implementation of a balanced, Merklized binary search tree.
It contains a list of nodes in the tree, which can be hashed together to get the Merkle root hash.
This hash must match the `AppHash` contained in the header at `FromChainHeight + 1`
- note the `+ 1` is necessary since `FromChainHeight` is the height in which the packet was committed,
and the resulting state root is not included until the next block.

### IBC State

Now that we've seen all the transaction types, let's talk about the state.
Each chain stores some IBC state in its Merkle tree.
For each chain being tracked by our chain, we store:

- Genesis configuration
- Latest state
- Headers for recent heights

We also store all incoming (ingress) and outgoing (egress) packets.

The state of a chain is updated every time an `IBCUpdateChainTx` is committed.
New packets are added to the egress state upon `IBCPacketCreateTx`.
New packets are added to the ingress state upon `IBCPacketPostTx`,
assuming the proof checks out.

## Merkle Queries

The Basecoin application uses a single Merkle tree that is shared across all its state,
including the built-in accounts state and all plugin state. For this reason,
it's important to use explicit key names and/or hashes to ensure there are no collisions.

We can query the Merkle tree using the ABCI Query method.
If we pass in the correct key, it will return the corresponding value,
as well as a proof that the key and value are contained in the Merkle tree.

The results of a query can thus be used as proof in an `IBCPacketPostTx`.

## Try it out

Now that we have all the background knowledge, let's actually walk through the tutorial.

Make sure you have installed
[Tendermint](https://tendermint.com/intro/getting-started/download) and
[basecoin](/docs/guide/install.md).

`basecoin` is a framework for creating new cryptocurrency applications.
It comes with an `IBC` plugin enabled by default.

You will also want to install the [jq](https://stedolan.github.io/jq/) for handling JSON at the command line.

Now let's start the two blockchains.
In this tutorial, each chain will have only a single validator,
where the initial configuration files are already generated.
Let's change directory so these files are easily accessible:

```
cd $GOPATH/src/github.com/tendermint/basecoin/demo
```

The relevant data is now in the `data` directory.
Before we begin, let's set some environment variables for convenience:

```
export BCHOME="."
BCHOME1="./data/chain1"
BCHOME2="./data/chain2"

export CHAIN_ID1=test_chain_1
export CHAIN_ID2=test_chain_2

CHAIN_FLAGS1="--chain_id $CHAIN_ID1 --from $BCHOME1/key.json"
CHAIN_FLAGS2="--chain_id $CHAIN_ID2 --from $BCHOME2/key.json --node tcp://localhost:36657"
```

In previous examples, we started basecoin in-process with tendermint.
Here, we will run them in different processes, using the `--without-tendermint` flag,
as described in the [guide to the basecoin tool](basecoin-tool.md).
We can start the two chains as follows:

```
TMROOT=$BCHOME1 tendermint node --log_level=info &> chain1_tendermint.log &
BCHOME=$BCHOME1 basecoin start --without-tendermint &> chain1_basecoin.log &
```

and

```
TMROOT=$BCHOME2 tendermint node --log_level=info --node_laddr tcp://localhost:36656 --rpc_laddr tcp://localhost:36657 --proxy_app tcp://localhost:36658 &> chain2_tendermint.log &
BCHOME=$BCHOME2 basecoin start --address tcp://localhost:36658 --without-tendermint &> chain2_basecoin.log &
```

Note how we refer to the relevant data directories, and how we set the various addresses for the second node so as not to conflict with the first.

We can now check on the status of the two chains:

```
curl localhost:46657/status
curl localhost:36657/status
```

If either command fails, the nodes may not have finished starting up. Wait a couple seconds and try again.
Once you see the status of both chains, it's time to move on.

In this tutorial, we're going to send some data from `test_chain_1` to `test_chain_2`.
We begin by registering `test_chain_1` on `test_chain_2`:

```
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS2 register --ibc_chain_id $CHAIN_ID1 --genesis $BCHOME1/genesis.json
```

Now we can create the outgoing packet on `test_chain_1`:

```
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS1 packet create --ibc_from $CHAIN_ID1 --to $CHAIN_ID2 --type coin --payload 0xDEADBEEF --ibc_sequence 1
```

Note our payload is just `DEADBEEF`.
Now that the packet is committed in the chain, let's get some proof by querying:

```
QUERY=$(basecoin query ibc,egress,$CHAIN_ID1,$CHAIN_ID2,1)
echo $QUERY
```

The result contains the latest height, a value (i.e. the hex-encoded binary serialization of our packet),
and a proof (i.e. hex-encoded binary serialization of a list of nodes from the Merkle tree) that the value is in the Merkle tree.
We keep the result in the `QUERY` variable so we can easily reference subfields using the `jq` tool.

If we want to send this data to `test_chain_2`, we first have to update what it knows about `test_chain_1`.
We'll need a recent block header and a set of commit signatures.
Fortunately, we can get them with the `block` command:

```
BLOCK=$(basecoin block $(echo $QUERY | jq .height))
echo $BLOCK
```

Here, we are passing `basecoin block` the `height` from our earlier query.
Note the result contains both a hex-encoded and json-encoded version of the header and the commit.
The former is used as input for later commands; the latter is human-readable, so you know what's going on!

Let's send this updated information about `test_chain_1` to `test_chain_2`.
First, output the header and commit for reference:

```
echo $BLOCK | jq .hex.header
echo $BLOCK | jq .hex.commit
```

And now forward those values to `test_chain_2`:

```
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS2 update --header 0x<header> --commit 0x<commit>
```

Now that `test_chain_2` knows about some recent state of `test_chain_1`, we can post the packet to `test_chain_2`,
along with proof the packet was committed on `test_chain_1`. Since `test_chain_2` knows about some recent state
of `test_chain_1`, it will be able to verify the proof!

First, output the height, packet, and proof for reference:

```
echo $QUERY | jq .height
echo $QUERY | jq .value
echo $QUERY | jq .proof
```

And forward those values to `test_chain_2`:

```
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS2 packet post --ibc_from $CHAIN_ID1 --height <height> --packet 0x<packet> --proof 0x<proof>
```

If the command does not return an error, then we have successfuly transfered data from `test_chain_1` to `test_chain_2`. Tada!

## Conclusion

In this tutorial we explained how IBC works, and demonstrated how to use it to communicate between two chains.
We did the simplest communciation possible: a one way transfer of data from chain1 to chain2.
The most important part was that we updated chain2 with the latest state (i.e. header and commit) of chain1,
and then were able to post a proof to chain2 that a packet was committed to the outgoing state of chain1.

In a future tutorial, we will demonstrate how to use IBC to actually transfer tokens between two blockchains,
but we'll do it with real testnets deployed across multiple nodes on the network. Stay tuned!

