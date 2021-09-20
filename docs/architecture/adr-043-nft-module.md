# ADR 43: NFT Module

## Changelog

- 05.05.2021: Initial Draft
- 07.01.2021: Incorporate Billy's feedback
- 07.02.2021: Incorporate feedbacks from Aaron, Shaun, Billy et al.

## Status

DRAFT

## Abstract

This ADR defines the `x/nft` module which is a generic implementation of NFTs, roughly "compatible" with ERC721.

## Context

NFTs are more than just crypto art, which is very helpful for accruing value to the Cosmos ecosystem. As a result, Cosmos Hub should implement NFT functions and enable a unified mechanism for storing and sending the ownership representative of NFTs as discussed in https://github.com/cosmos/cosmos-sdk/discussions/9065.

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

#### Class

We define a model for NFT **Class**, which is comparable to an ERC721 Contract on Ethereum, under which a collection of NFTs can be created and managed.

```protobuf
message Class {
  string id          = 1;
  string name        = 2;
  string symbol      = 3;
  string description = 4;
  string uri         = 5;
}
```

- `id` is an alphanumeric identifier of the NFT class; it is used as the primary index for storing the class; _required_
- `name` is a descriptive name of the NFT class; _optional_
- `symbol` is the symbol usually shown on exchanges for the NFT class; _optional_
- `description` is a detailed description of the NFT class; _optional_
- `uri` is a URL pointing to an off-chain JSON file that contains metadata about this NFT class ([OpenSea example](https://docs.opensea.io/docs/contract-level-metadata)); _optional_

#### NFT

We define a general model for `NFT` as follows.

```protobuf
message NFT {
  string class_id           = 1;
  string id                 = 2;
  string uri                = 3;
  google.protobuf.Any data  = 10;
}
```

- `class_id` is the identifier of the NFT class where the NFT belongs; _required_
- `id` is an alphanumeric identifier of the NFT, unique within the scope of its class. It is specified by the creator of the NFT and may be expanded to use DID in the future. `class_id` combined with `id` uniquely identifies an NFT and is used as the primary index for storing the NFT; _required_
  ```
  {class_id}/{id} --> NFT (bytes)
  ```
- `uri` is a URL pointing to an off-chain JSON file that contains metadata about this NFT (Ref: [ERC721 standard and OpenSea extension](https://docs.opensea.io/docs/metadata-standards)); _required_
- `data` is a field that CAN be used by composing modules to specify additional properties for the NFT; _optional_

This ADR doesn't specify values that `data` can take; however, best practices recommend upper-level NFT modules clearly specify their contents.  Although the value of this field doesn't provide the additional context required to manage NFT records, which means that the field can technically be removed from the specification, the field's existence allows basic informational/UI functionality.

### `Keeper` Interface

```go
type Keeper interface {
  NewClass(class Class)
  UpdateClass(class Class)

  Mint(nft NFT，receiver sdk.AccAddress)   // updates totalSupply
  Burn(classId string, nftId string)    // updates totalSupply
  Update(nft NFT)
  Transfer(classId string, nftId string, receiver sdk.AccAddress)

  GetClass(classId string) Class
  GetClasses() []Class

  GetNFT(classId string, nftId string) NFT
  GetNFTsOfClassByOwner(classId string, owner sdk.AccAddress) []NFT
  GetNFTsOfClass(classId string) []NFT

  GetOwner(classId string, nftId string) sdk.AccAddress
  GetBalance(classId string, owner sdk.AccAddress) uint64
  GetTotalSupply(classId string) uint64
}
```

Other business logic implementations should be defined in composing modules that import `x/nft` and use its `Keeper`.

### `Msg` Service

```protobuf
service Msg {
  rpc Send(MsgSend)         returns (MsgSendResponse);
}

message MsgSend {
  string class_id = 1;
  string id       = 2;
  string sender   = 3;
  string reveiver = 4;
}
message MsgSendResponse {}
```

`MsgSend` can be used to transfer the ownership of an NFT to another address.

The implementation outline of the server is as follows:

```go
type msgServer struct{
  k Keeper
}

func (m msgServer) Send(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
  // check current ownership
  assertEqual(msg.Sender, m.k.GetOwner(msg.ClassId, msg.Id))

  // transfer ownership
  m.k.Transfer(msg.ClassId, msg.Id, msg.Receiver)

  return &types.MsgSendResponse{}, nil
}
```

The query service methods for the `x/nft` module are:

```proto
service Query {

  // Balance queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
  rpc Balance(QueryBalanceRequest) returns (QueryBalanceResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/balance/{class_id}/{owner}";
  }

  // Owner queries the owner of the NFT based on its class and id, same as ownerOf in ERC721
  rpc Owner(QueryOwnerRequest) returns (QueryOwnerResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/owner/{class_id}/{id}";
  }

  // Supply queries the number of NFTs of a given class, same as totalSupply in ERC721Enumerable
  rpc Supply(QuerySupplyRequest) returns (QuerySupplyResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/supply/{class_id}";
  }

  // NFTsOfClassByOwner queries the NFTs of a given class owned by the owner, similar to tokenOfOwnerByIndex in ERC721Enumerable
  rpc NFTsOfClassByOwner(QueryNFTsOfClassByOwnerRequest) returns (QueryNFTsResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/owned_nfts/{class_id}/{owner}";
  }

  // NFTsOfClass queries all NFTs of a given class, similar to tokenByIndex in ERC721Enumerable
  rpc NFTsOfClass(QueryNFTsOfClassRequest) returns (QueryNFTsResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts/{class_id}";
  }

  // NFT queries an NFT based on its class and id.
  rpc NFT(QueryNFTRequest) returns (QueryNFTResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts/{class_id}/{id}";
  }

  // Class queries an NFT class based on its id
  rpc Class(QueryClassRequest) returns (QueryClassResponse) {
      option (google.api.http).get = "/cosmos/nft/v1beta1/classes/{class_id}";
  }

  // Classes queries all NFT classes
  rpc Classes(QueryClassesRequest) returns (QueryClassesResponse) {
      option (google.api.http).get = "/cosmos/nft/v1beta1/classes";
  }
}

// QueryBalanceRequest is the request type for the Query/Balance RPC method
message QueryBalanceRequest {
  string class_id = 1;
  string owner    = 2;
}

// QueryBalanceResponse is the response type for the Query/Balance RPC method
message QueryBalanceResponse{
  uint64 amount = 1;
}

// QueryOwnerRequest is the request type for the Query/Owner RPC method
message QueryOwnerRequest {
  string class_id = 1;
  string id       = 2;
}

// QueryOwnerResponse is the response type for the Query/Owner RPC method
message QueryOwnerResponse{
  string owner = 1;
}

// QuerySupplyRequest is the request type for the Query/Supply RPC method
message QuerySupplyRequest {
  string class_id = 1;
}

// QuerySupplyResponse is the response type for the Query/Supply RPC method
message QuerySupplyResponse {
  uint64 amount = 1;
}

// QueryNFTsOfClassByOwnerRequest is the request type for the Query/NFTsOfClassByOwner RPC method
message QueryNFTsOfClassByOwnerRequest {
  string                                 class_id   = 1;
  string                                 owner      = 2;
  cosmos.base.query.v1beta1.PageResponse pagination = 3;
}

// QueryNFTsOfClassRequest is the request type for the Query/NFTsOfClass RPC method
message QueryNFTsOfClassRequest {
  string                                 class_id   = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

// QueryNFTsResponse is the response type for the Query/NFTsOfClass and Query/NFTsOfClassByOwner RPC methods
message QueryNFTsResponse {
  repeated cosmos.nft.v1beta1.NFT        nfts       = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

// QueryNFTRequest is the request type for the Query/NFT RPC method
message QueryNFTRequest {
  string class_id = 1;
  string id       = 2;
}

// QueryNFTResponse is the response type for the Query/NFT RPC method
message QueryNFTResponse {
  cosmos.nft.v1beta1.NFT nft = 1;
}

// QueryClassRequest is the request type for the Query/Class RPC method
message QueryClassRequest {
  string class_id = 1;
}

// QueryClassResponse is the response type for the Query/Class RPC method
message QueryClassResponse {
  cosmos.nft.v1beta1.Class class = 1;
}

// QueryClassesRequest is the request type for the Query/Classes RPC method
message QueryClassesRequest {
  // pagination defines an optional pagination for the request.
  cosmos.base.query.v1beta1.PageRequest pagination = 1;
}

// QueryClassesResponse is the response type for the Query/Classes RPC method
message QueryClassesResponse {
  repeated cosmos.nft.v1beta1.Class      classes    = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```


### Interoperability

Interoperability is all about reusing assets between modules and chains. The former one is achieved by ADR-33: Protobuf client - server communication. At the time of writing ADR-33 is not finalized. The latter is achieved by IBC. Here we will focus on the IBC side.
IBC is implemented per module. Here, we aligned that NFTs will be recorded and managed in the x/nft. This requires creation of a new IBC standard and implementation of it.

For IBC interoperability, NFT custom modules MUST use the NFT object type understood by the IBC client. So, for x/nft interoperability, custom NFT implementations (example: x/cryptokitty) should use the canonical x/nft module and proxy all NFT balance keeping functionality to x/nft or else re-implement all functionality using the NFT object type understood by the IBC client. In other words: x/nft becomes the standard NFT registry for all Cosmos NFTs (example: x/cryptokitty will register a kitty NFT in x/nft and use x/nft for book keeping). This was [discussed](https://github.com/cosmos/cosmos-sdk/discussions/9065#discussioncomment-873206) in the context of using x/bank as a general asset balance book. Not using x/nft will require implementing another module for IBC.


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

+ New IBC app is required for x/nft

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
