<!--
order: 2
-->

# State

## Class

Class is mainly composed of `id`, `name`, `symbol`, `description`, `uri`, `uri_hash`,`data` where `id` is the unique identifier of the class, similar to the Ethereum ERC721 contract address, the others are optional.

* Class: `0x01 | classID | -> ProtocolBuffer(Class)`

## NFT

NFT is mainly composed of `class_id`, `id`, `uri`, `uri_hash` and `data`. Among them, `class_id` and `id` are two-tuples that identify the uniqueness of nft, `uri` and `uri_hash` is optional, which identifies the off-chain storage location of the nft, and `data` is an Any type. Use Any chain of `x/nft` modules can be customized by extending this field

* NFT: `0x02 | classID | 0x00 | nftID |-> ProtocolBuffer(NFT)`

## NFTOfClassByOwner

NFTOfClassByOwner is mainly to realize the function of querying all nfts using classID and owner, without other redundant functions.

* NFTOfClassByOwner: `0x03 | owner | 0x00 | classID | 0x00 | nftID |-> 0x01`

## Owner

Since there is no extra field in NFT to indicate the owner of nft, an additional key-value pair is used to save the ownership of nft. With the transfer of nft, the key-value pair is updated synchronously.

* OwnerKey: `0x04 | classID | 0x00  | nftID |-> owner`

## TotalSupply

TotalSupply is responsible for tracking the number of all nfts under a certain class. Mint operation is performed under the changed class, supply increases by one, burn operation, and supply decreases by one.

* OwnerKey: `0x05 | classID |-> totalSupply`
