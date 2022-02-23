# Proofs IAVL

This is a demo project to show converting proofs from cosmos/iavl into the format
specified in confio/proofs and validating that they still work.

## Library usage

It exposes a two main functions :

`func CreateMembershipProof(tree *iavl.MutableTree, key []byte) (*proofs.CommitmentProof, error)`
produces a CommitmentProof that the given key exists in the iavl tree (and contains the
current value). This returns an error if the key does not exist in the tree.

`func CreateNonMembershipProof(tree *iavl.MutableTree, key []byte) (*proofs.CommitmentProof, error)`
produces a CommitmentProof that the given key doesn't exist in the iavl tree.
This returns an error if the key does not exist in the tree.

Generalized range proofs are lower in priority, as they are just an optimization of the
two basic proof types, and don't provide any additional capabilities.
We will soon add some `Batch` capabilities that can represent these.

## CLI usage

We also expose a simple script to generate test data for the confio proofs package.

```shell
go install ./cmd/testgen-iavl
testgen-iavl
```

Will output some json data, from a randomly generated merkle tree each time.

```json
{
  "existence": "0a146f65436a684273735a34567543774b567a435963121e76616c75655f666f725f6f65436a684273735a34567543774b567a4359631a0d0a0b0801180120012a030002021a2d122b08011204020402201a2120d307032505383dee34ea9eadf7649c31d1ce294b6d62b273d804da478ac161da1a2d122b08011204040802201a2120306b7d51213bd93bac17c5ee3d727ec666300370b19fd55cc13d7341dc589a991a2b12290801122508160220857103d59863ac55d1f34008a681f837c01975a223c0f54883a05a446d49c7c6201a2b1229080112250a2202204498eb5c93e40934bc8bad9626f19e333c1c0be4541b9098f139585c3471bae2201a2d122b080112040e6c02201a212022648db12dbf830485cc41435ecfe37bcac26c6c305ac4304f649977ddc339d51a2c122a0801122610c60102204e0b7996a7104f5b1ac1a2caa0704c4b63f60112e0e13763b2ba03f40a54e845201a2c122a08011226129003022017858e28e0563f7252eaca19acfc1c3828c892e635f76f971b3fbdc9bbd2742e20",
  "root": "cea07656c77e8655521f4c904730cf4649242b8e482be786b2b220a15150d5f0"
}
```

`"root"` is the hex-encoded root hash of the merkle tree

`"existence"` is the hex-encoding of the protobuf binary encoding of a `proofs.ExistenceProof` object. This contains a (key, value) pair,
along with all steps to reach the root hash. This provides a non-trivial test case, to ensure client in multiple languages can verify the
protobuf proofs we generate from the iavl tree
