# Bech32 on Cosmos

The Cosmos network prefers to use the Bech32 address format wherever users must handle binary data. Bech32 encoding provides robust integrity checks on data and the human readable part(HRP) provides contextual hints that can assist UI developers with providing informative error messages.

In the Cosmos network, keys and addresses may refer to a number of different roles in the network like accounts, validators etc.

## HRP table

| HRP              | Definition                            |
| ---------------- | ------------------------------------- |
| cosmos           | Cosmos Account Address                |
| cosmospub        | Cosmos Account Public Key             |
| cosmosvalcons    | Cosmos Validator Consensus Address    |
| cosmosvalconspub | Cosmos Validator Consensus Public Key |
| cosmosvaloper    | Cosmos Validator Operator Address     |
| cosmosvaloperpub | Cosmos Validator Operator Public Key  |

## Encoding

While all user facing interfaces to Cosmos software should exposed Bech32 interfaces, many internal interfaces encode binary value in hex or base64 encoded form.

To covert between other binary representation of addresses and keys, it is important to first apply the Amino encoding process before bech32 encoding.

A complete implementation of the Amino serialization format is unnecessary in most cases. Simply prepending bytes from this [table](https://github.com/tendermint/spec/blob/master/spec/blockchain/encoding.md#public-key-cryptography) to the byte string payload before bech32 encoding will sufficient for compatible representation.
