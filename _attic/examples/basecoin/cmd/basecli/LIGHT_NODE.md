# Run your own (super) lightweight node

In addition to providing command-line tooling that goes cryptographic verification
on all the data your receive from the node, we have implemented a proxy mode, that
allows you to run a super lightweight node.  It does not follow the chain on
every block or even every header, but only as needed. But still providing the
same security as running a full non-validator node on your local machine.

Basically, it runs as a proxy that exposes the same rpc interface as the full node
and connects to a (potentially untrusted) full node. Every response is cryptographically
verified before being passed through, returning an error if it doesn't match.

You can expect 2 rpc calls for every query plus <= 1 query for each validator set
change. Going offline for a while allows you to verify multiple validator set changes
with one call. Cuz at 1 block/sec and 1000 tx/block, it just doesn't make sense
to run a full node just to get security

## Setup

Just initialize your client with the proper validator set as in the [README](README.md)

```
$ export BCHOME=~/.lightnode
$ basecli init --node tcp://<host>:<port> --chain-id <chain>
```

## Running

```
$ basecli proxy --serve tcp://localhost:7890
...
curl localhost:7890/status
curl localhost:7890/block\?height=20
```

You can even subscribe to events over websockets and they are all verified
before passing them though.  Though if you want every block, you might as
well run a full (nonvalidating) node.

## Seeds

Every time the validator set changes, the light node verifies if it is legal,
and then creates a seed at that point.  These "seeds" are verified checkpoints
that we can trace any proof back to, starting with one on `init`.

To make sure you are based on the most recent header, you can run:

```
basecli seeds update
basecli seeds show
```

## Feedback

This is the first release of basecli and the light-weight proxy. It is secure, but
may not be useful for your workflow. Please try it out and open github issues
for any enhancements or bugs you find.  I am aiming to make this a very useful
tool by tendermint 0.11, for which I need community feedback.
