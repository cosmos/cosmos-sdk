# ADR 43: BaseNFT Module

## Changelog

- 05.05.2021: Initial Draft

## Status

DRAFT

## Abstract

This ADR defines a generic NFT module of `x/nft` which supports NFTs as a `proto.Any` and contains `BaseNFT` as the default implementation.

## Context

As was discussed in [#9065](https://github.com/cosmos/cosmos-sdk/discussions/9065), several potential solutions can be considered:
- irismod/nft and modules/incubator/nft
- CW721
- DID NFTs
- interNFT

Considering generic usage and compatibility of interchain protocols including IBC and Gravity Bridge, it is preferred to have a very simple NFT module design which only stores NFTs by id and owner. 

Application-specific functions (minting, burning, transferring, etc.) should be managed by other modules on Cosmos Hub or on other Zones.

The current design is based on the work done by [IRISnet team](https://github.com/irisnet/irismod/tree/master/modules/nft) and an older implementation in the [Cosmos repository](https://github.com/cosmos/modules/tree/master/incubator/nft).


## Decision

We will create a module `x/nft` which only stores NFTs by id and owner.

The interface for the `x/nft` module:

```go
// NFTI is an interface used to store NFTs at a given id.
type NFTI interface {
	GetId() string // define a simple identifier for the NFT 
	GetOwner() sdk.AccAddress // the current owner's address
	GetData() string // specialized metadata to accompany the nft
}
```

The NFT conforms to the following specifications:
  * The Id is an immutable field used as a unique identifier. NFT identifiers don't currently have a naming convention but
    can be used in association with existing Hub attributes, e.g., defining an NFT's identifier as an immutable Hub address allows its integration into existing Hub account management modules. 
    We envision that identifiers can accommodate mint and transfer actions. 
  * Owner: This mutable field represents the current account owner (on the Hub) of the NFT, i.e., the account that will have authorization
    to perform various activities in the future. This can be a user, a module account, or potentially a future NFT module that adds capabilities.
  * Data: This mutable field allows for attaching pertinent metadata to the NFT, e.g., to record special parameter change upon transferring or criteria being met.
    The data field is used to maintain special information, such as name and URI. Currently, there is no convention for the data placed into this field,
    however, best practices recommend defining clear specifications that apply to each specific NFT type.

We will also create `BaseNFT` as the default implementation of the `NFTI` interface:
```proto
message BaseNFT {
  option (gogoproto.equal) = true;

  string id    = 1;
  string owner = 2;
  string data  = 3;
}
```

Queries functions for `BaseNFT` is:
```proto
service Query {

  // NFT queries NFT details based on id.
  rpc NFT(QueryNFTRequest) returns (QueryNFTResponse) {
    option (google.api.http).get = "/cosmos/nft/v1beta1/nfts/{id}";
  }

  // NFTs queries all proposals based on the optional onwer.
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
  google.protobuf.Any nft = 1 [(cosmos_proto.accepts_interface) = "NFTI", (gogoproto.customname) = "NFT"];
}

// QueryNFTsRequest is the request type for the Query/NFTs RPC method
message QueryNFTsRequest {
  string                                 owner      = 1;
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}

// QueryNFTsResponse is the response type for the Query/NFTs RPC method
message QueryNFTsResponse {
  repeated google.protobuf.Any nfts = 1 [(cosmos_proto.accepts_interface) = "NFTI", (gogoproto.customname) = "NFTs"];
  cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```



## Consequences

### Backward Compatibility

No backwards incompatibilities.

### Forwards Compatibility

This specification conforms to the ERC-721 smart contract specification for NFT identifiers. Note that ERC-721 defines uniqueness based on (contract address, uint256 tokenId), and we conform to this implicitly 
because a single module is currently aimed to track NFT identifiers. Note: use of the (mutable) data field to determine uniqueness is not safe. 

### Positive

- NFT identifiers available on Cosmos Hub
- Ability to build different NFT modules for the Cosmos Hub, e.g., ERC-721.
- NFT module which supports interoperability with IBC and other cross-chain infrastractures like Gravity Bridge

### Negative

- Currently, no methods are defined for this module except to store and retrieve data.

### Neutral

- Other functions need more modules. For example, a custody module is needed for NFT trading function, a collectible module is needed for defining NFT properties

## Further Discussions

For other kinds of applications on the Hub, more app-specific modules can be developed in the future:
- `x/nft/collectibles`: grouping NFTs into collections and defining properties of NFTs such as minting, burning and transferring, etc.
- `x/nft/custody`: custody of NFTs to support trading functionality
- `x/nft/marketplace`: selling and buying NFTs using sdk.Coins

Other networks in the Cosmos ecosystem could design and implement their own NFT modules for specific NFT applications and usecases.

## References

- Initial discussion: https://github.com/cosmos/cosmos-sdk/discussions/9065
- x/nft: initialize module: https://github.com/cosmos/cosmos-sdk/pull/9174