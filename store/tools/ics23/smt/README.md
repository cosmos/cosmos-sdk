# Proofs SMT

This project demonstrates the generation and validation of ICS-23 proofs for a sparse Merkle tree (SMT) as implemented by [Celestia](https://github.com/celestiaorg/smt).

## Library usage

It exposes a two main functions :

`func CreateMembershipProof(tree *smt.SparseMerkleTree, key []byte) (*ics23.CommitmentProof, error)`
produces a CommitmentProof that the given key exists in the SMT (and contains the current value). This returns an error if the key does not exist in the tree.

`func CreateNonMembershipProof(tree *smt.SparseMerkleTree, key []byte, preimages PreimageMap) (*ics23.CommitmentProof, error)`
produces a CommitmentProof that the given key doesn't exist in the SMT. This returns an error if the key does not exist in the tree.
This relies on an auxiliary `PreimageMap` object which provides access to the preimages of all keys in the tree based on their (hashed) path ordering.


## CLI usage

We also expose a simple script to generate test data for the confio proofs package.

```shell
go install ./cmd/testgen-smt
testgen-smt exist left 10
```

Will output some json data, from a randomly generated Merkle tree each time.

```json
{
  "key": "574f516c4364415274743845444d397347484937",
  "proof": "0a9d010a2024910c64b5b74b6b72e6b9d3310a1d0bd599032e05e8abc43112d194e1a78f30121e76616c75655f666f725f574f516c4364415274743845444d3973474849371a07080118012a0100222708011201011a20b51557119b6985d54a48a4510e528d5f929f0b1c8b57914bb6cd8f9eab035d75222708011201011a20fff8248ca9e98cbb05c81612d38e74780b2c02d9c88ee628cfbdb8ca44769a63",
  "root": "f69ef3599b7f0471b61735490636608a8ff43a327b2b5a3a5528ca7f7059ffa5",
  "value": "76616c75655f666f725f574f516c4364415274743845444d397347484937"
}
```

`"root"` is the hex-encoded root hash of the Merkle tree.

`"proof"` is the hex-encoding of the protobuf binary encoding of a `proofs.ExistenceProof` object. This contains a (key, value) pair, along with all steps to reach the root hash. This provides a non-trivial test case, to ensure clients in multiple languages can verify the protobuf proofs we generate from the SMT.
