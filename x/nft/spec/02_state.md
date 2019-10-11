# State

## Collections

As all NFTs belong to a specific `Collection`, they are kept on store in an array
within each `Collection`. Every time an NFT that belongs to a collection is updated,
it needs to be updated on the corresponding NFT array on the corresponding `Collection`.
`denomHash` is used as part of the key to limit the length of the `denomBytes` which is
 a hash of `denomBytes` made from the tendermint [tmhash library](https://github.com/tendermint/tendermint/tree/master/crypto/tmhash).

- Collections: `0x00 | denomHash -> amino(Collection)`
- denomHash: `tmhash(denomBytes)`

## Owners

The ownership of an NFT is set initially when an NFT is minted and needs to be
updated every time there's a transfer or when an NFT is burned.

- Owners: `0x01 | addressBytes | denomHash -> amino(Owner)`
- denomHash: `tmhash(denomBytes)`
