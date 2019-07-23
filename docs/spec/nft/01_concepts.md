# Concepts

## NFT

The `NFT` Interface inherits the BaseNFT struct and includes getter functions for the asset data. It also lincludes a Stringer function in order to print the struct. The interface may change if metadata is moved to it’s own module as it might no longer be necessary for the flexibility of an interface.

```go
// NFT non fungible token interface
type NFT interface {
  GetID() string
  GetOwner() sdk.AccAddress
  SetOwner(address sdk.AccAddress)
  GetName() string
  GetDescription() string
  GetImage() string
  GetTokenURI() string
  EditMetadata(name, description, image, tokenURI string)
  String() string
}
```

## Collections

A Collection is used to organized sets of NFTs. It contains the denomination of the NFT instead of storing it within each NFT. This saves storage space by removing redundancy.

```go
// Collection of non fungible tokens
type Collection struct {
  Denom string `json:"denom,omitempty"` // name of the collection; not exported to clients
  NFTs  NFTs   `json:"nfts"`            // NFTs that belong to a collection
}
```

## Owner

An Owner is a struct that includes information about all NFTs owned by a single account. It would be possible to retrieve this information by looping through all Collections but that process could become computationaly prohibitive so a more efficient retrieval system is to store redundant information limited to the token ID by owner.

```go
// Owner of non fungible tokens
type Owner struct {
  Address       sdk.AccAddress `json:"address"`
  IDCollections IDCollections  `json:"IDCollections"`
}
```
