# Standard Library

The quarks framework comes with a number of standard modules that provide a lot
of common functionality that is useful to a wide variety of applications,
and also provide good examples to use when developing your own modules. Before
starting to write code, see if the functionality is already here.

## Basic Middleware

### Logging

`modules.base.Logger` is a middleware that records basic info on CheckTx,
DeliverTx, and SetOption, along with timing in microseconds. It can be installed
standard at the top of all middleware stacks, or replace it with your own
Middleware if you want to record more custom information with each request.

### Recovery

To avoid accidental panics (eg. bad go-wire decoding) killing the abci app,
wrap the stack with `stack.Recovery`, which catches all panics and returns
them as errors, so they can be handled normally.

### Signatures

The first layer of the tx contains the signatures to authorize it.  This is then
verfied by `modules.auth.Signatures`.  All tx may have one or multiple signatures
which are then processed and verified by this middleware and then passed down
the stack.

### Chain

The next layer of a tx (in the standard stack) binds the tx to a specific chain
with an optional expiration height.  This keeps the tx from being replayed on
a fork or other such chain, as well as a partially signed multisig being delayed
months before being committed to the chain. This functionality is provided in
`modules.base.Chain`

### Nonce

To avoid replay protection within one chain, we want a nonce associated
with each account. Rather than force everything to use coins as a payment,or force each module to implement its own replay protection, each tx is wraped with a nonce and
the account it belongs to.  This must be one higher than the last request or
the request is rejected. This is implemented in `modules.nonce.ReplayCheck`

You can also take a look at the [design discussion](https://github.com/tendermint/basecoin/issues/160)

### Fees

An optional feature, but useful on many chains, is charging a fee for every
transaction. A simple implementation of this is provided in
`modules.fee.SimpleFeeMiddleware`. A fee currency and minimum amount are
defined in the constructor (eg. in code).  If the minimum amount is 0, then
the fee is optional. If it is above 0, then every tx with insufficient fee is
rejected. This fee is deducted from the payers account before executing any
other transaction.

This module depends on the `coin` module.

## Other Apps

### Coin

What would a crypto-currency be without tokens? The sendtx logic from basecoin
was extracted into one module, which is now optional, meaning most of the other
functionality would also work in a system with no built-in tokens, such as
a private network that provides another access control mechanism.

`modules.coin.Handler` defines a Handler that maintains a number of accounts
along with a set of various tokens, supporting multiple denominations. The
main access is `SendTx`, which can support any type of actor (other apps as
well as public key addresses), and is a building block for any other app that
requires some payment solution, like fees or trader.

### Roles

Roles encapsulates what are typically called N-of-M multi-signatures accounts
in the crypto world. However, I view this as a type of role or group, which can
be the basis for building a permision system. For example, a set of people could
be called registrars, which can authorize a new IBC chain, and need eg. 2 out
of 7 signatures to approve it.

Currently, one can create a role with `modules.roles.Handler`, and assume one
of those roles by wrapping another transaction with `AssumeRoleTx`, which is
processed by `modules.roles.Middleware`. Updating the set of actors in
a role is planned in the near future.

### IBC

IBC, or inter-blockchain communication, is the cornerstone of cosmos, and built
into the quark framework as a basic primative. To properly understand these
concepts requires a much longer explanation, but in short, the chain works
as a light-client to another chain and maintains input and output queue to
send packets with that chain.

Most functionality is implemented in `modules.ibc.Handler`. Registering a chain
is a seed of trust that requires verification of the proper seed (or genesis
block), and this generally requires approval of an authorized registrar (which
may be a multi-sig role).  Updating a registered chain can be done by anyone,
as the new header can be completely verified by the existing knowledge of the
chain.  Also, modules can initiate an outgoing IBC message to another chain
by calling `CreatePacketTx` over IPC (inter-plugin communication) with a tx
that belongs to their module. (This must be explicitly authorized by the
same module, so only the eg. coin module can authorize a sendtx to another
chain).

`PostPacketTx` can post a tx that was created on another chain along with the
merkle proof, which must match an already registered header. If this chain
can verify the authenticity, it will accept the packet, along with all the
permissions from the other chain, and execute it on this stack. This is the
only way to get permissions that belong to another chain.

These various pieces can be combined in a relay, which polls for new packets
on one chain, and then posts the packets along with the new headers on the
other chain.

## Planned Apps

### Staking

Straight-forward PoS as used for cosmos.
Based on [basecoin-stake](https://github.com/tendermint/basecoin-stake)

### Voting

Simple elections that can authorize other tx, like roles. A building block for
governance.

### Trader

Escrow, OTC option, Order book.  Based on [basecoin-examples](https://github.com/tendermint/basecoin-examples/tree/develop/trader).  This may be more appropriate
for an external repo.

