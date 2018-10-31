# Appendices

([Back to table of contents](README.md#contents))

## Appendix A: Encoding Libraries

The specification has focused on semantics and functionality of the IBC protocol. However in order to facilitate the communication between multiple implementations of the protocol, we seek to define a standard syntax, or binary encoding, of the data structures defined above. Many structures are universal and for these, we provide one standard syntax. Other structures, such as _H<sub>h </sub>, U<sub>h </sub>, _and _X<sub>h</sub>_ are tied to the consensus engine and we can define the standard encoding for tendermint, but support for additional consensus engines must be added separately. Finally, there are some aspects of the messaging, such as the envelope to post this data (fees, nonce, signatures, etc.), which is different for every chain, and must be known to the relay, but are not important to the IBC algorithm itself and left undefined.

In defining a standard binary encoding for all the "universal" components, we wish to make use of a standardized library, with efficient serialization and support in multiple languages. We considered two main formats: Ethereum's RLP[[6](./references.md#6)] and Google's Protobuf[[7](./references.md#7)]. We decided for protobuf, as it is more widely supported, is more expressive for different data types, and supports code generation for very efficient (de)serialization codecs. It does have a learning curve and more setup to generate the code from the type specifications, but the ibc data types should not change often and this code generation setup only needs to happen once per language (and can be exposed in a common repo), so this is not a strong counter-argument. Efficiency, expressiveness, and wider support rule in its favor. It is also widely used in gRPC and in many microservice architectures.

The tendermint-specific data structures are encoded with go-amino[[8](./references.md#8)], the native binary encoding used inside of tendermint. Most blockchains define their own formats, and until some universal format for headers and signatures among blockchains emerge, it seems very premature to enforce any encoding here. These are defined as arbitrary byte slices in the protocol, to be parsed in an consensus engine-dependent manner.

For the following appendixes, the data structure specifications will be in proto3[[9](./references.md#9)] format.

## Appendix B: IBC Queue Format

The foundational data structure of the IBC protocol are the packet queues stored inside each chain. We start with a well-defined binary representation of the keys and values used in these queues. The encodings mirror the semantics defined above:

_key = _(_remote id, [send|receipt], [head|tail|index])_

_V<sub>send</sub> = (maxHeight, maxTime, type, data)_

_V<sub>receipt</sub> = (result, [success|error code])_

Keys and values are binary encoded and stored as bytes in the Merkle tree in order to generate the root hash stored in the block header, which validates all proofs. They are treated as arrays of bytes by the Merkle proofs for deterministically generating the sequence of hashes and passed as such in all interchain messages. Once the validity of a key value pair has been determined from the Merkle proof and header, the payload bytes can be deserialized and interpreted by the protocol.

See [binary format as protobuf specification](./protobuf/queue.proto)

## Appendix C: Merkle Proof Formats

A Merkle tree (or a trie) generates a single hash that can be used to prove any element of the tree. In order to generate this hash, we first hash the leaf nodes, then hash multiple leaf nodes together to get the hash of an inner node (two or more, based on degree k of the k-ary tree), and repeat for each level of the tree until we end up with one root hash.
With a known root hash (which is included in the block headers), the existence of a particular key/value in the tree can be proven by tracing the path to the value and revealing the (k-1) hashes for the paths not taken on each level ([[10](./references.md#10)]).

There are a number of different implementations of this basic idea, using different hash functions, as well as prefixes to prevent second preimage attacks (differentiating leaf nodes from inner nodes). Rather than force all chains that wish to participate in IBC to use the same data store, we provide a data structure that can represent Merkle proofs from a variety of data stores, and provide for chaining proofs to allow for subtrees. While searching for a solution, we did find the chainpoint proof format[[11](./references.md#11)], which inspired this design significantly, but didn't (yet) offer the flexibility we needed.

We generalize the left/right idiom to the concatenation a (possibly empty) fixed prefix, the (just calculated) last hash, and a (possibly empty) fixed suffix. We must only define two fields on each level and can represent any type, even a 16-ary Patricia tree, with this structure. One must only translate from the store's native proof to this format, and it can be verified by any chain, providing compatibility with arbitrary data stores.

The proof format also allows for chaining of trees, combining multiple Merkle stores into a "multi-store". Many applications (such as the EVM) define a data store with a large proof size for internal use. Rather than force them to change the store (impossible), or live with huge proofs (inefficient), we provide the possibility to express Merkle proofs connecting multiple subtrees. Thus, one could have one subtree for data, and a second for IBC. Each tree produces its own Merkle root, and these are then hashed together to produce the root hash that is stored in the block header.

A valid Merkle proof for IBC must either consist of a proof of one tree, and prepend `ibc` to all key names as defined above, or use a subtree named `ibc` in the first section, and store the key names as above in the second tree.

In order to minimize the size of their Merkle proofs, we recommend using Tendermint's IAVL+ tree implementation[[12](./references.md#12)], which is designed for optimal proof size and released under a permissive license. It uses an AVL tree (a type of binary tree) with ripemd160 as the hashing algorithm at each stage. This produces optimally compact proofs, ideal for posting in blockchain transactions. For a data store of _n_ values, there will be _log<sub>2</sub>(n)_ levels, each requiring one 20-byte hash for proving the branch not taken (plus possible metadata for the level). We can express a proof in a tree of 1 million elements in something around 400 bytes. If we further store all IBC messages in a separate subtree, we should expect the count of nodes in this tree to be a few thousand, and require less than 400 bytes, even for blockchains with a large state.

See [binary format as protobuf specification](./protobuf/merkle.proto)

## Appendix D: Byzantine Recovery Strategies

IBC guarantees reliable, ordered packet delivery in the face of malicious nodes or relays, on top of which application invariants can be ensured. However, all guarantees break down when the blockchain on the other end of the connection exhibits Byzantine behavior. This can take two forms: a failure of the consensus mechanism (reverting previously finalized blocks), or a failure at the application level (not correctly performing the application-level functions on the packet).

The IBC protocol can detect a limited class of Byzantine faults at the consensus level by identifying duplicate headers -- if an IBC module ever sees two different headers for the same height (or any evidence that headers belong to different forks), then it can freeze the connection immediately. State reconciliation (e.g. restoring token balances to owners of vouchers on the other chain) must be handled by blockchain governance.

If there is a big divide in the remote chain and the validation set splits (e.g. 60-40 weighted) as to the direction of the chain, then the light-client header update protocol will refuses to follow either fork. If both sides declare a hard fork and continue with new validator sets that are not compatible with the consensus engine (they don't have ⅔ support from the previous block), then the connection(s) will need to be reopened manually (by governance on the local chain) and set to the new header set(s). The IBC protocol doesn't have the option to follow both chains as the queue and associated state must map to exactly one remote chain. In a fork, the chain can continue the connection with one fork, and optionally make a fresh connection with the other fork.

Another kind of Byzantine action is at the application level. Let us assume packets represent transfer of value. If chain `A` sends a message with `x` tokens to chain `B`, then it promises to remove `x` tokens from the local supply. And if chain `B` handles this message successfully, it promises to credit `x` token vouchers to the account indicated in the packet. If chain `A` does not remove tokens from supply, or chain `B` does not generate vouchers, the application invariants (conservation of supply & fungibility) break down.

The IBC protocol does not handle these kinds of errors. They must be handled individually by each application. Applications could use Plasma-like fraud proofs to allow state recovery on one chain if fraud can be proved on the other chain. Although complex to implement, a correct implementation would allow applications to guarantee their invariants as long as *either* blockchain's consensus algorithm behaves correctly (and this could be extended to `n` chains). Economic incentives can additionally be used to disincentivize any kind of provable fraud.

## Appendix E: Tendermint Header Proofs

{ Ensure this is correct. }

**TODO: clean this all up**

This is a mess now, we need to figure out what formats we use, define go-amino, etc. or just point to the source???? Will do more later, need help here from the tendermint core team.

In order to prove a merkle root, we must fully define the headers, signatures, and validator information returned from the Tendermint consensus engine, as well as the rules by which to verify a header. We also define here the messages used for creating and removing connections to other blockchains as well as how to handle forks.

Building Blocks: Header, PubKey, Signature, Commit, ValidatorSet

→ needs input/support from Tendermint Core team (and go-crypto)

Registering Chain

Updating Header

Validator Changes

**ROOT of trust**

As mentioned in the definitions, all proofs are based on an original assumption. The root of trust here is either the genesis block (if it is newer than the unbonding period) or any signed header of the other chain.

When governance on a pair of chain, the respective chains must agree to a root of trust on the counterparty chain. This can be the genesis block on a chain that launches with an IBC channel or a later block header.

From this signed header, one can check the validator set against the validator hash stored in the header, and then verify the signatures match. This provides internal consistency and accountability, but if 5 nodes provide you different headers (eg. of forks), you must make a subjective decision which one to trust. This should be performed by on-chain governance to avoid an exploitable position of trust.

**VERIFYING HEADERS**

Once we have a trusted header with a known validator set, we can quickly validate any new header with the same validator set. To validate a new header, simply verifying that the validator hash has not changed, and that over 2/3 of the voting power in that set has properly signed a commit for that header. We can skip all intervening headers, as we have complete finality (no forks) and accountability (to punish a double-sign).

This is safe as long as we have a valid signed header by the trusted validator set that is within the unbonding period for staking. In that case, if we were given a false (forked) header, we could use this as proof to slash the stake of all the double-signing validators. This demonstrates the importance of attribution and is the same security guarantee of any non-validating full node. Even in the presence of some ultra-powerful malicious actors, this makes the cost of creating a fake proof for a header equal to at least one third of all staked tokens, which should be significantly higher than any gain of a false message.

**UPDATING VALIDATORS SET**

If the validator hash is different than the trusted one, we must simultaneously both verify that if the change is valid while, as well as use using the new set to validate the header.  Since the entire validator set is not provided by default when we give a header and commit votes, this must be provided as extra data to the certifier.

A validator change in Tendermint can be securely verified with the following checks:

*   First, that the new header, validators, and signatures are internally consistent
    *   We have a new set of validators that matches the hash on the new header
    *   At least 2/3 of the voting power of the new set validates the new header
*   Second, that the new header is also valid in the eyes of our trust set
    *   Verify at least 2/3 of the voting power of our trusted set, which are also in the new set, properly signed a commit to the new header

In that case, we can update to this header, and update the trusted validator set, with the same guarantees as above (the ability to slash at least one third of all staked tokens on any false proof).
