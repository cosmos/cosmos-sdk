# Proofs Tendermint

This is an adapter library to convert the `SimpleProof` from
[tendermint/crypto/merkle](https://github.com/tendermint/tendermint/tree/master/crypto/merkle)
into the standard confio/proofs format.

As non-existence proofs depend on ordered keys, and all proofs require the key-value pair
to be encoded in a predictable format in the leaves, we will only support proofs generated
from `SimpleProofsFromMap`, which handles the key-value pairs for leafs in a standard format.

## Library usage

It exposes a top-level function `func ConvertSimpleProof(p *merkle.SimpleProof, key, value []byte) (*proofs.ExistenceProof, error)`
that can convert from `merkle.SimpleProof` with the KVPair encoding for leafs, into confio/proof protobuf objects.

It currently only works for existence proofs. We plan to soon support non-existence proofs.

## CLI usage

We also expose a simple script to generate test data for the confio proofs package.

```shell
make testgen
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
