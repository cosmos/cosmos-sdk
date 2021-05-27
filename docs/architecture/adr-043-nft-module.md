# ADR 43: NFT Module

## Changelog

- 05.05.2021: Initial Draft

## Status

DRAFT

## Abstract

This ADR defines a generic NFT module of `x/nft` which supports NFTs as a `proto.Any` and contains `NFT` as the default implementation.

## Context
NFTs are more digital assets than only crypto arts, which is very helpful for accruing value to ATOM. As a result, Cosmos Hub should implement NFT functions and enable a unified mechanism for storing and sending the ownership representative of NFTs as discussed in https://github.com/cosmos/cosmos-sdk/discussions/9065.

As was discussed in [#9065](https://github.com/cosmos/cosmos-sdk/discussions/9065), several potential solutions can be considered:
- irismod/nft and modules/incubator/nft
- CW721
- DID NFTs
- interNFT

Since NFTs functions/use cases are tightly connected with their logic, it is almost impossible to support all the NFTs' use cases in one Cosmos SDK module by defining and implementing different transaction types.

Considering generic usage and compatibility of interchain protocols including IBC and Gravity Bridge, it is preferred to have a very simple NFT module design which only stores NFTs by id and owner. 

This design idea can enable composability that application-specific functions should be managed by other modules on Cosmos Hub or on other Zones by importing the NFT module.

For example, NFTs can be grouped into collections in a collectibles module to define the behaviors such as minting, burning, etc.

The current design is based on the work done by [IRISnet team](https://github.com/irisnet/irismod/tree/master/modules/nft) and an older implementation in the [Cosmos repository](https://github.com/cosmos/modules/tree/master/incubator/nft).


## Decision

We will create a module `x/nft` which only stores NFTs by id and owner.

The interface for the `x/nft` module:

```go
// NFT is an interface used to store NFTs at a given id.
type NFT interface {
    GetId() string // define a simple identifier for the NFT 
    GetOwner() sdk.AccAddress // the current owner's address
    GetData() string // specialized metadata to accompany the nft
    Validate() error 
}
```

The NFT conforms to the following specifications:
  * The Id is an immutable field used as a unique identifier. NFT identifiers don't currently have a naming convention but
    can be used in association with existing Hub attributes, e.g., defining an NFT's identifier as an immutable Hub address allows its integration into existing Hub account management modules. 
    We envision that identifiers can accommodate mint and transfer actions.
    The Id is also the primary index for storing NFTs.
    ```
    id (string) --> NFT (bytes)
    ```
  * Owner: This mutable field represents the current account owner (on the Hub) of the NFT, i.e., the account that will have authorization
    to perform various activities in the future. This can be a user, a module account, or potentially a future NFT module that adds capabilities.
    Owner is also the secondary index for storing NFT IDs owned by an address
    ```
    owner (address) --> []string
    ```
  * Data: This mutable field allows attaching special information to the NFT, for example:
    * metadata such as title of the work and URI
    * immutable data and parameters (such actual NFT data, hash or seed for generators)
    * mutable data and parameters that change when transferring or when certain criteria are met (such as provenance)
      
    This ADR doesn't specify values that this field can take; however, best practices recommend upper-level NFT modules clearly specify its contents.
    Although the value of this field doesn't provide additional context required to manage NFT records, which means that the field can technically be removed from the specification, 
    the field's existence allows basic informational/UI functionality.

The logic for transferring the ownership of an NFT to another address (no coins involved) should also be defined in this module
```go
func (k Keeper) TransferOwnership(nftID string, currentOwner, newOwner sdk.AccAddress) error
```

Other behaviors of NFT, including minting, burning and other business-logic implementation should be defined in other upper-level modules that import this NFT module.

We will also create `BaseNFT` as the default implementation of the `NFT` interface:
```proto
message BaseNFT {
  option (gogoproto.equal) = true;

  string id    = 1;
  string owner = 2;
  string data  = 3;
}
```

The query service methods for the `x/nft` module are:
```proto
service Query {

  // NFT queries NFT details based on id.
  rpc NFT(QueryNFTRequest) returns (QueryNFTResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts/{id}";
  }

  // NFTs queries all NFTs based on the optional onwer.
  rpc NFTs(QueryNFTsRequest) returns (QueryNFTsResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts";
  }
}

// QueryNFTRequest is the request type for the Query/NFT RPC method
message QueryNFTRequest {
  string id = 1;
}

// QueryNFTResponse is the response type for the Query/NFT RPC method
message QueryNFTResponse {
  google.protobuf.Any nft = 1 [(cosmos_proto.accepts_interface) = "NFT", (gogoproto.customname) = "NFT"];
}

// QueryNFTsRequest is the request type for the Query/NFTs RPC method
message QueryNFTsRequest {
  string                                 owner      = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

// QueryNFTsResponse is the response type for the Query/NFTs RPC method
message QueryNFTsResponse {
  repeated google.protobuf.Any nfts = 1 [(cosmos_proto.accepts_interface) = "NFT", (gogoproto.customname) = "NFTs"];
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```



## Consequences

### Backward Compatibility

No backward incompatibilities.

### Forward Compatibility

This specification conforms to the ERC-721 smart contract specification for NFT identifiers. Note that ERC-721 defines uniqueness based on (contract address, uint256 tokenId), and we conform to this implicitly 
because a single module is currently aimed to track NFT identifiers. Note: use of the (mutable) data field to determine uniqueness is not safe. 

### Positive

- NFT identifiers available on Cosmos Hub
- Ability to build different NFT modules for the Cosmos Hub, e.g., ERC-721.
- NFT module which supports interoperability with IBC and other cross-chain infrastructures like Gravity Bridge

### Negative

- Currently, no methods are defined for this module except to store and retrieve data.

### Neutral

- Other functions need more modules. For example, a custody module is needed for NFT trading function, a collectible module is needed for defining NFT properties

## Further Discussions

For other kinds of applications on the Hub, more app-specific modules can be developed in the future:
- `x/nft/collectibles`: grouping NFTs into collections and defining properties of NFTs such as minting, burning and transferring, etc.
- `x/nft/custody`: custody of NFTs to support trading functionality
- `x/nft/marketplace`: selling and buying NFTs using sdk.Coins

Other networks in the Cosmos ecosystem could design and implement their own NFT modules for specific NFT applications and use cases.

## References

- Initial discussion: https://github.com/cosmos/cosmos-sdk/discussions/9065
- x/nft: initialize module: https://github.com/cosmos/cosmos-sdk/pull/9174
