# Glossary

This defines many of the terms that are used in the other documents.  If there
is every a concept that seems unclear, check here.

## Transaction

A transaction is a packet of binary data that contains all information to
validate and perform an action on the blockchain. The only other data that
it interacts with is the current state of the chain (kv store), and it must
have a deterministic action. The transaction is the main piece of one request.

We currently make heavy use of go-wire and go-data to provide automatic binary
and json encodings (and decodings) for objects, even when they embed many
interfaces inside. There is one public `TxMapper` in the basecoin root package,
and all modules can register their own transaction types there. This allows us
to deserialize the entire tx in one location (even with types defined in other
repos), to easily embed an arbitrary Tx inside another without specifying
the specific type, and provide an automatic json representation to provide to
users (or apps) to inspect the chain.

## Context

As the request passes through the system, it can pick up information, that must
be carried along with it.  Like the authorized it has received from another
middleware, or the block height it runs at.  This is all deterministic
information from the context in which the request runs (based on the tx and
the block it was included in) and can be used to validate the tx.

## Data Store

To be able to provide proofs to tendermint, we keep all data in one key-value
store, indexed with a merkle tree.  This allows us to easily provide a root
hash and proofs for queries without requiring complex logic inside each
module. Standarizing this also allows powerful light-client tooling as it knows
how to verify all data in the store.

The downside is there is one quite simple interface that the application has
to `Get` and `Set` data. There is not even a range query.  Although there are
some data structures like queues and range queries that are also in the `state`
package to provide higher-level functionality in a standard format.

## Isolation

One of the main arguments for blockchain is security.  So while we encourage
the use of third-party modules, we must be vigilant against security holes.
If you use the `stack` package, it will provide two different types of
sandboxing for you.

The first step, is that when `DeliverTx` is called on a module, it is never
given the entire data store, but rather only its own prefixed section. This
is achieved by prefixing all keys transparently with `<module name> + 0x0`,
using the null byte as a separator.  Since module name must be a string, no
clever naming scheme can lead to a collision. Inside the module, we can write
anywhere we want, without worry that we have to touch some data that is not ours.

The second step involves the permissions in the context.  The context can say
that this tx was signed by eg. Rigel.  But if any module can add that permission,
it would be too easy to forge accounts.  Thus, each permission is associated
with the module that granted it (in this case `auth`), and if a module tries
to add a permission for another module, it will panic. There is also
protection if a module creates a brand new fake context to trick the downstream
modules.

This means that modules can confidently write to their local section of the
database and trust the permissions associated with the context, without concern
of interferance from other modules.  (Okay, if you see a bunch of C-code in
the module traversing through all the memory space of the application, then
get worried....)

## Handler

## Middleware

## Modules

## Dispathcer

## IPC (Inter-Plugin Communication)

## IBC (Inter-Blockchain Communication)

Wow, this is a big topic.  Also a WIP.  Add more here...
