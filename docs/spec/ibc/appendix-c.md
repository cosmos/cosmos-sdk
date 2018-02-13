## Appendix C: Merkle Proof Formats

([Back to table of contents](specification.md#contents))

A merkle tree (or a trie) generates one hash that can prove every element of the tree. Generating this hash starts with hashing the leaf nodes. Then hashing multiple leaf nodes together to get the hash of an inner node (two or more, based on degree k of the k-ary tree). And continue hashing together the inner nodes at each level of the tree, until it reaches a root hash. Once you have a known root hash, you can prove key/value belongs to this tree by tracing the path to the value and revealing the (k-1) hashes for all the paths we did not take on each level. If this is new to you, you can read a basic introduction[[10](./footnotes.md#10)].

There are a number of different implementations of this basic idea, using different hash functions, as well as prefixes to prevent second preimage attacks (differentiating leaf nodes from inner nodes). Rather than force all chains that wish to participate in IBC to use the same data store, we provide a data structure that can represent merkle proofs from a variety of data stores, and provide for chaining proofs to allow for sub-trees. While searching for a solution, we did find the chainpoint proof format[[11](./footnotes.md#11)], which inspired this design significantly, but didn't (yet) offer the flexibility we needed.

We generalize the left/right idiom to concatenating a (possibly empty) fixed prefix, the (just calculated) last hash, and a (possibly empty) fixed suffix. We must only define two fields on each level and can represent any type, even a 16-ary Patricia tree, with this structure. One must only translate from the store's native proof to this format, and it can be verified by any chain, providing compatibility for arbitrary data stores.

The proof format also allows for chaining of trees, combining multiple merkle stores into a "multi-store". Many applications (such as the EVM) define a data store with a large proof size for internal use. Rather than force them to change the store (impossible), or live with huge proofs (inefficient), we provide the possibility to express merkle proofs connecting multiple subtrees. Thus, one could have one subtree for data, and a second for IBC. Each tree produces their own merkle root, and these are then hashed together to produce the root hash that is stored in the block header.

A valid merkle proof for IBC must either consist of a proof of one tree, and prepend "ibc" to all key names as defined above, or use a subtree named "ibc" in the first section, and store the key names as above in the second tree.

For those who wish to minimize the size of their merkle proofs, we recommend using Tendermint's IAVL+ tree implementation[[12](./footnotes.md#12)], which is designed for optimal proof size, and freely available for use. It uses an AVL tree (a type of binary tree) with ripemd160 as the hashing algorithm at each stage. This produces optimally compact proofs, ideal for posting in blockchain transactions. For a data store of _n_ values, there will be _log<sub>2</sub>(n)_ levels, each requiring one 20-byte hash for proving the branch not taken (plus possible metadata for the level). We can express a proof in a tree of 1 million elements in something around 400 bytes. If we further store all IBC messages in a separate subtree, we should expect the count of nodes in this tree to be a few thousand, and require less than 400 bytes, even for blockchains with a quite large state.

```
 // HashOp is the hashing algorithm we use at each level
 enum HashOp {
     RIPEMD160 = 0;
     SHA224 = 1;
     SHA256 = 2;
     SHA384 = 3;
     SHA512 = 4;
     SHA3_224 = 5;
     SHA3_256 = 6;
     SHA3_384 = 7;
     SHA3_512 = 8;
     SHA256_X2 = 9;
 };
 // Op represents one hash in a chain of hashes.
 // An operation takes the output of the last level and returns
 // a hash for the next level:
 // Op(last) => Operation(prefix + last + sufix)
 //
 // A simple left/right hash would simply set prefix=left or
 // suffix=right and leave the other blank. However, one could
 // also represent the a Patricia trie proof by setting
 // prefix to the rlp encoding of all nodes before the branch
 // we select, and suffix to all those after the one we select.
 message Op {
     bytes prefix = 1;
     bytes suffix = 2;
     HashOp op = 3;
 }
 // Data is the end value stored, used to generate the initial hash
 message Data {
     bytes prefix = 1;
     bytes key = 2;
     bytes value = 3;
     HashOp op = 4;
     // If it is KeyValue, this is the data we want
     // If it is SubTree, key is name of the tree, value is root hash
     // Expect another branch to follow
     enum DataType {
         KeyValue = 0;
         SubTree = 1;
     }
     DataType dataType = 5;
 }
 // Branch will hash data and then pass it through operations from
 // last to first in order to calculate the root node.
 //
 // Visualize Branch as representing the data closest to root as the
 // first item, and the leaf as the last item.
 message Branch {
     repeated Op operations = 1;
     Data data = 2;
 }
 // MerkleProof shows a veriable path from the data to
 // a root hash (potentially spanning multiple sub-trees).
 message MerkleProof {
  // identify the header this is rooted in
  string chainId = 1;
  uint64 height = 2;
  // this hash must match the header as well as the
  // calculation from below
  bytes rootHash = 3;
  // branches start from the value, and then may
  // include multiple subtree branches to embed it
  //
  // The first branch must have dataType KeyValue
  // Following branches must have dataType SubTree
  repeated Branch branches = 1;
 }
 ```

