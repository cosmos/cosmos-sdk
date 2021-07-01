# ADR 43: NFT Module

## Changelog

- 05.05.2021: Initial Draft
- 07.01.2021: Incorporate Billy's feedback

## Status

DRAFT

## Abstract

This ADR defines the `x/nft` module which is a generic storage of NFTs, roughly "compatible" with ERC721.

## Context

NFTs are more digital assets than only crypto arts, which is very helpful for accruing value to the Cosmos ecosystem. As a result, Cosmos Hub should implement NFT functions and enable a unified mechanism for storing and sending the ownership representative of NFTs as discussed in https://github.com/cosmos/cosmos-sdk/discussions/9065.

As was discussed in [#9065](https://github.com/cosmos/cosmos-sdk/discussions/9065), several potential solutions can be considered:

- irismod/nft and modules/incubator/nft
- CW721
- DID NFTs
- interNFT

Since functions/use cases of NFTs are tightly connected with their logic, it is almost impossible to support all the NFTs' use cases in one Cosmos SDK module by defining and implementing different transaction types.

Considering generic usage and compatibility of interchain protocols including IBC and Gravity Bridge, it is preferred to have a generic NFT module design which handles the generic NFTs logic.

This design idea can enable composability that application-specific functions should be managed by other modules on Cosmos Hub or on other Zones by importing the NFT module.

The current design is based on the work done by [IRISnet team](https://github.com/irisnet/irismod/tree/master/modules/nft) and an older implementation in the [Cosmos repository](https://github.com/cosmos/modules/tree/master/incubator/nft).

## Decision

We will create a module `x/nft`, which contains the following functionality:

- Store NFTs and track their ownership.
- Expose `Keeper` interface for composing modules to mint and burn NFTs.
- Expose external `Message` interface for users to transfer ownership of their NFTs.
- Query NFTs and their supply information.

### Types

#### Genre

We define a model for NFT **Genre**, which is comparable to an ERC721 Contract on Ethereum, under which a collection of NFTs can be created and managed.

```protobuf
message Genre {
  string id              = 1;
  string name            = 2;
  string symbol          = 3;
  string description     = 4;
  string uri             = 5;
  bool mint_restricted   = 10;
  bool update_restricted = 11;
}
```

- `id` is an alphanumeric identifier of the NFT genre; it is used as the primary index for storing the genre; _required_
- `name` is a descriptive name of the NFT genre; _optional_
- `symbol` is the symbol usually shown on exchanges for the NFT genre; _optional_
- `description` is a detailed description of the NFT genre; _optional_
- `uri` is a URL pointing to an off-chain JSON file that contains metadata about this NFT genre ([OpenSea example](https://docs.opensea.io/docs/contract-level-metadata)); _optional_
- `mint_restricted` flag, if set to true, indicates that only the genre owner can mint NFTs, otherwise anyone can do so; _required_
- `udpate_restricted` flag, if set to true, indicates that no one can update NFTs, otherwise only NFT owners can do so; _required_

#### NFT

We define a general model for `NFT` as follows.

```protobuf
message NFT {
  string genre = 1;
  string id    = 2;
  string uri   = 3;
  string data  = 10;
}
```

- `genre` is identifier of genre where the NFT belongs; _required_
- `id` is an alphanumeric identifier of the NFT, unique within the scope of its genre. It is specified by the creator of the NFT and may be expanded to use DID in the future. `genre` combined with `id` uniquely identifies an NFT and is used as the primary index for storing the NFT; _required_
  ```
  {genre}/{id} --> NFT (bytes)
  ```
- `uri` is a URL pointing to an off-chain JSON file that contains metadata about this NFT (Ref: [ERC721 standard and OpenSea extension](https://docs.opensea.io/docs/metadata-standards)).
- `data` is a field that CAN be used by composing modules to specify additional properties for the NFT; _optional_

This ADR doesn't specify values that `data` can take; however, best practices recommend upper-level NFT modules clearly specify their contents.  Although the value of this field doesn't provide the additional context required to manage NFT records, which means that the field can technically be removed from the specification, the field's existence allows basic informational/UI functionality.

### `Keeper` Interface (TODO)
Other business logic implementations should be defined in composing modules that import this NFT module and use its `Keeper`.

### `Msg` Service

```protobuf
service Msg {
  rpc Send(MsgSend)         returns (MsgSendResponse);
}

message MsgSend {
  string genre    = 1;
  string id       = 2;
  string sender   = 3;
  string reveiver = 4;
}
message MsgSendResponse {}

`MsgSend` can be used to transfer the ownership of an NFT to another address.

The implementation outline of the server is as follows:

```go
type msgServer struct{
  k Keeper
}

func (m msgServer) Send(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
  // check current ownership
  assertEqual(msg.Sender, m.k.GetNftOwner(msg.Genre, msg.Id))

  // change ownership mapping
  m.k.SetNftOwner(msg.Genre, msg.Id, msg.Receiver)

  return &types.MsgSendResponse{}, nil
}

The query service methods for the `x/nft` module are:

```proto
service Query {

  // Balance queries the number of NFTs based on the genre and owner, same as balanceOf in ERC721
  rpc Balance(QueryBalanceRequest) returns (QueryBalanceResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/balance/{genre}/{owner}";
  }

  // Owner queries the owner of the NFT based on the genre and id, same as ownerOf in ERC721
  rpc Owner(QueryOwnerRequest) returns (QueryOwnerResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/owner/{genre}/{id}";
  }

  // Supply queries the number of NFTs based on the genre, same as totalSupply in ERC721Enumerable
  rpc Supply(QuerySupplyRequest) returns (QuerySupplyResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/supply/{genre}";
  }

  // NFTsOf queries all NFTs based on the genre, similar to tokenByIndex in ERC721Enumerable
  rpc NFTs(QueryNFTsRequest) returns (QueryNFTsResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts/{genre}";
  }

  // NFTsOfOwner queries the NFTs based on the genre and owner, similar to tokenOfOwnerByIndex in ERC721Enumerable
  rpc NFTsOfOwner(QueryNFTsOfOwnerRequest) returns (QueryNFTsResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/balance/{genre}/{owner}";
  }

  // NFT queries NFT details based on genre and id.
  rpc NFT(QueryNFTRequest) returns (QueryNFTResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts/{genre}/{id}";
  }

  // Genre queries the definition of a given genre
  rpc Genre(QueryGenreRequest) returns (QueryGenreResponse) {
      option (google.api.http).get = "/cosmos/nft/v1beta1/genres/{genre}";
  }

  // Types queries all the genres
  rpc Genres(QueryGenresRequest) returns (QueryGenresResponse) {
      option (google.api.http).get = "/cosmos/nft/v1beta1/genres";
  }
}

// QueryBalanceRequest is the request type for the Query/Balance RPC method
message QueryBalanceRequest {
  string genre = 1;
  string owner = 2;
}

// QueryBalanceResponse is the response type for the Query/Balance RPC method
message QueryBalanceResponse{
  uint64 amount = 1;
}

// QueryOwnerRequest is the request type for the Query/Owner RPC method
message QueryOwnerRequest {
  string genre = 1;
  string id    = 2;
}

// QueryOwnerResponse is the response type for the Query/Owner RPC method
message QueryOwnerResponse{
  string owner = 1;
}

// QuerySupplyRequest is the request type for the Query/Supply RPC method
message QuerySupplyRequest {
  string genre = 1;
}

// QuerySupplyResponse is the response type for the Query/Supply RPC method
message QuerySupplyResponse {
  uint64 amount = 1;
}

// QueryNFTsRequest is the request type for the Query/NFTs RPC method
message QueryNFTsRequest {
  string                                 genre      = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

// QueryNFTsOfOwnerRequest is the request type for the Query/NFTsOfOwner RPC method
message QueryNFTsOfOwnerRequest {
  string                                 genre      = 1;
  string                                 owner      = 2;
  cosmos.base.query.v1beta1.PageResponse pagination = 3;
}

// QueryNFTsResponse is the response type for the Query/NFTs and Query/NFTsOfOwner RPC method
message QueryNFTsResponse {
  repeated cosmos.nft.v1beta1.NFT        nfts       = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

// QueryNFTRequest is the request type for the Query/NFT RPC method
message QueryNFTRequest {
  string genre = 1;
  string id    = 2;
}

// QueryNFTResponse is the response type for the Query/NFT RPC method
message QueryNFTResponse {
  cosmos.nft.v1beta1.NFT nft = 1;
}

// QueryGenreRequest is the request type for the Query/Genre RPC method
message QueryGenreRequest {
  string genre = 1;
}

// QueryGenreResponse is the response type for the Query/Genre RPC method
message QueryGenreResponse {
  cosmos.nft.v1beta1.Genre genre = 1;
}

// QueryGenresRequest is the request type for the Query/Genres RPC method
message QueryGenresRequest {
  // pagination defines an optional pagination for the request.
  cosmos.base.query.v1beta1.PageRequest pagination = 1;
}

// QueryGenresResponse is the response type for the Query/Genres RPC method
message QueryGenresResponse {
  repeated cosmos.nft.v1beta1.Genre      genres     = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

## Consequences

### Backward Compatibility

No backward incompatibilities.

### Forward Compatibility

This specification conforms to the ERC-721 smart contract specification for NFT identifiers. Note that ERC-721 defines uniqueness based on (contract address, uint256 tokenId), and we conform to this implicitly because a single module is currently aimed to track NFT identifiers. Note: use of the (mutable) data field to determine uniqueness is not safe.s

### Positive

- NFT identifiers available on Cosmos Hub.
- Ability to build different NFT modules for the Cosmos Hub, e.g., ERC-721.
- NFT module which supports interoperability with IBC and other cross-chain infrastructures like Gravity Bridge

### Negative


### Neutral

- Other functions need more modules. For example, a custody module is needed for NFT trading function, a collectible module is needed for defining NFT properties.

## Further Discussions

For other kinds of applications on the Hub, more app-specific modules can be developed in the future:

- `x/nft/custody`: custody of NFTs to support trading functionality.
- `x/nft/marketplace`: selling and buying NFTs using sdk.Coins.

Other networks in the Cosmos ecosystem could design and implement their own NFT modules for specific NFT applications and use cases.

## References

- Initial discussion: https://github.com/cosmos/cosmos-sdk/discussions/9065
- x/nft: initialize module: https://github.com/cosmos/cosmos-sdk/pull/9174
- [ADR 033](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-033-protobuf-inter-module-comm.md)
